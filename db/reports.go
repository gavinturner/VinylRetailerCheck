package db

import (
	"fmt"
	"github.com/gavinturner/vinylretailers/retailers"
	"github.com/gavinturner/vinylretailers/util/postgres"
	"github.com/pkg/errors"
	"gopkg.in/guregu/null.v3"
	"time"
)

type Batch struct {
	ID                   int64     `db:"id" json:"id"`
	NumRequiredSearches  int       `db:"req_searches" json:"numRequiredSearches"`
	NumCompletedSearches int       `db:"completed_searches" json:"numCompletedSearches"`
	CreatedAt            time.Time `db:"created_at"`
	UpdatedAt            time.Time `db:"updated_at"`
	ReportedAt           time.Time `db:"reported_at" json:"reportedAt"`
}
type Report struct {
	ID          int64     `db:"id" json:"id"`
	UserID      int64     `db:"user_id" json:"userId"`
	BatchID     int64     `db:"batch_id" json:"batchId"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time `db:"updated_at"`
	CompletedAt null.Time `db:"completed_at" json:"completedAt"`
	SentAt      null.Time `db:"sent_at" json:"sentAt"`
}

type BatchedReport struct {
	ReportID  int64  `db:"report_id" json:"reportId"`
	UserID    int64  `db:"user_id" json:"userId"`
	UserName  string `db:"user_name" json:"userName"`
	UserEmail string `db:"user_email" json:"userEmail"`
	BatchID   int64  `db:"batch_id" json:"batchId"`
}

//
// AddNewBatch
// Create a new batch instance, representing one full set of scans to be completed over the set of users
// currently watching a set of artists. This batch instance will have a set of empty reports connected to
// it, one per watching user, and each such report will cover the set of artists currently being watched
// by the user for whom the report pertains. The batch also stores the required number of searches to be
// performed (artist x retailer) so that we can keep track of whether a batch has been completed or not.
//
func (v *VinylDB) AddNewBatch(tx *postgres.Tx, numRequiredSearches int, userArtists map[int64][]WatchedArtist) (batchId int64, err error) {
	querier := v.Q(tx)

	// create a new batch instance
	err = querier.Get(&batchId, querier.Rebind(`
		INSERT INTO batches (req_searches) VALUES (?) RETURNING id
	`), numRequiredSearches)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to add new batch")
	}

	//
	// for each user currently watching a set of artists, create a report instance for the batch
	// to populate with results. attach the report to the batch and record all the artists that the
	// specific report covers
	//
	for userId, artists := range userArtists {
		var reportId int64
		err = querier.Get(&reportId, querier.Rebind(
			`INSERT INTO reports (user_id, batch_id) VALUES (?, ?) RETURNING id
		`), userId, batchId)
		if err != nil {
			return 0, errors.Wrapf(err, "failed to add new report to batch for user %v", userId)
		}
		for _, watch := range artists {
			_, err = querier.Exec(querier.Rebind(`
				INSERT INTO report_artists(batch_id, report_id, artist_id) VALUES (?, ?, ?)
			`), batchId, reportId, watch.ArtistID)
			if err != nil {
				return 0, errors.Wrapf(err, "failed to add new artist to report for artict %v", watch.ArtistID)
			}
		}
	}
	return batchId, err
}

func (v *VinylDB) DeleteReportsForBatch(tx *postgres.Tx, batchId int64) error {
	querier := v.Q(tx)
	_, err := querier.Exec(querier.Rebind(`
		DELETE FROM report_artists WHERE report_id in (
			SELECT id FROM reports WHERE batch_id = ?
		)`), batchId)
	if err != nil {
		return errors.Wrapf(err, "failed to delete report artists for batch reports")
	}
	_, err = querier.Exec(querier.Rebind(`
		DELETE FROM reports WHERE batch_id = ?
	`), batchId)
	if err != nil {
		return errors.Wrapf(err, "failed to delete reports for batch")
	}
	return nil
}

func (v *VinylDB) DeleteBatch(tx *postgres.Tx, batchId int64) error {
	querier := v.Q(tx)
	err := v.DeleteReportsForBatch(tx, batchId)
	if err != nil {
		return errors.Wrapf(err, "failed to delete reports for batch %v", batchId)
	}
	_, err = querier.Exec(querier.Rebind(`
		DELETE FROM batches WHERE id = ?
	`), batchId)
	if err != nil {
		return errors.Wrapf(err, "failed to delete batch %v", batchId)
	}
	return nil
}

//
// IncrementBatchSearchCompletedCount
// When scanning is complete for a specific artist + retailer, this method allows the scanner to increment
// the number of scans completed that is stored against the batch. We use UPDATE here as an atomic operation
// on the batches table and as such this method will be thread-safe in terms of getting all increments.
//
func (v *VinylDB) IncrementBatchSearchCompletedCount(tx *postgres.Tx, batchId int64) error {
	querier := v.Q(tx)
	rows, err := querier.Exec(querier.Rebind(`
		UPDATE batches SET completed_searches = completed_searches+1 WHERE id = ?
	`), batchId)
	if err != nil {
		return errors.Wrapf(err, "failed to add new batch")
	}
	affected, err := rows.RowsAffected()
	if err != nil {
		return errors.Wrapf(err, "failed to get rows affected")
	}
	if affected == 0 {
		return fmt.Errorf("batch %v was not incremented. does it exist?")
	}
	return nil
}

//
// AddSKUToReportsForBatch
// Given a particular SKU sound for an artist + retailer, this method looks at all the reports attached to the
// current batch, determines if the SKU artist is covered by the report, and if so attaches the SKU to the final
// report. We assume that it has already been determined whether the SKU represents a valid result for the report
// (for example the price has changed).
//
func (v *VinylDB) AddSKUToReportsForBatch(tx *postgres.Tx, batchId int64, sku *SKU) error {
	querier := v.Q(tx)
	reportIds := []int64{}
	err := querier.Select(&reportIds, querier.Rebind(`
		SELECT report_id FROM report_artists WHERE batch_id = ? AND artist_id = ?
	`), batchId, sku.ArtistID)
	if err != nil {
		return errors.Wrapf(err, "failed to get reports for batch artist")
	}
	if len(reportIds) == 0 {
		return fmt.Errorf("no reports for artist %v in batch %v", sku.ArtistID, batchId)
	}

	query := "INSERT INTO report_skus (report_id, sku_id) VALUES "
	args := []interface{}{}
	for idx, reportId := range reportIds {
		args = append(args, reportId, sku.ID)
		query += "(?, ?)"
		if idx < len(reportIds)-1 {
			query += ","
		}
	}
	_, err = querier.Exec(querier.Rebind(query), args...)
	if err != nil {
		return errors.Wrapf(err, "failed to insert sku into batch reports")
	}
	return nil
}

func (v *VinylDB) GetAllCompletedUnsetReports(tx *postgres.Tx) ([]BatchedReport, error) {
	querier := v.Q(tx)
	batches := []BatchedReport{}
	err := querier.Select(&batches, `
		SELECT 
			b.id as batch_id,
			r.id as report_id,
			u.id as user_id,
			u.name as user_name,
			u.email as user_email
		FROM 
			batches b
			JOIN reports r ON r.batch_id = b.id
			JOIN users u ON r.user_id = u.id
		WHERE 
			b.reported_at IS NULL 
			AND b.req_searches = b.completed_searches
			AND r.completed_at IS NULL
	`)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get reports for batch artist")
	}
	return batches, nil
}

func (v *VinylDB) MarkBatchReported(tx *postgres.Tx, batchId int64) error {
	querier := v.Q(tx)
	rows, err := querier.Exec(querier.Rebind(`
		UPDATE batches SET reported_at = CURRENT_TIMESTAMP WHERE id = ?
	`), batchId)
	if err != nil {
		return errors.Wrapf(err, "failed to update batch reported_at")
	}
	affected, err := rows.RowsAffected()
	if err != nil {
		return errors.Wrapf(err, "failed to get rows affected")
	}
	if affected == 0 {
		return fmt.Errorf("batch %v was not set as reported does it exist?")
	}
	return nil
}

func (v *VinylDB) MarkReportSent(tx *postgres.Tx, reportId int64) error {
	querier := v.Q(tx)
	rows, err := querier.Exec(querier.Rebind(`
		UPDATE reports SET sent_at = CURRENT_TIMESTAMP, completed_at = CURRENT_TIMESTAMP WHERE id = ?
	`), reportId)
	if err != nil {
		return errors.Wrapf(err, "failed to update report sent_at")
	}
	affected, err := rows.RowsAffected()
	if err != nil {
		return errors.Wrapf(err, "failed to get rows affected")
	}
	if affected == 0 {
		return fmt.Errorf("report %v was not set as sent. does it exist?")
	}
	return nil
}

func (v *VinylDB) DeleteReport(tx *postgres.Tx, reportId int64) error {
	querier := v.Q(tx)
	rows, err := querier.Exec(querier.Rebind(`
		DELETE FROM reports WHERE id = ?
	`), reportId)
	if err != nil {
		return errors.Wrapf(err, "failed to delete report %v", reportId)
	}
	affected, err := rows.RowsAffected()
	if err != nil {
		return errors.Wrapf(err, "failed to get rows affected")
	}
	if affected == 0 {
		return fmt.Errorf("report %v was not deleted? does it exist?")
	}
	return nil
}

func (v *VinylDB) GetSkusForReport(tx *postgres.Tx, reportId int64) ([]retailers.SKU, error) {
	querier := v.Q(tx)
	skus := []retailers.SKU{}
	err := querier.Select(&skus, querier.Rebind(`
		SELECT
			rt.name as retailer,
			a.name as artist,
			r.name as name,
    		s.item_url,
    		s.image_url,
    		s.price
		FROM 
			report_skus rs 
			JOIN skus s ON rs.sku_id = s.id
			JOIN artists a ON s.artist_id = a.id
			JOIN releases r ON s.release_id = r.id
			JOIN retailers rt ON s.retailer_id = rt.id
		WHERE 
			rs.report_id = ?
	`), reportId)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get skus for report %v", reportId)
	}
	return skus, nil
}
