package db

import (
	"fmt"
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

func (v *VinylDB) AddNewBatch(tx *postgres.Tx, numRequiredSearches int, userArtists map[int64][]WatchedArtist) (batchId int64, err error) {
	querier := v.Q(tx)
	err = querier.Get(&batchId, querier.Rebind(`
		INSERT INTO batches (req_searches) VALUES (?) RETURNING id
	`), numRequiredSearches)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to add new batch")
	}
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
