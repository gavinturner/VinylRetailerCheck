package db

import (
	"fmt"
	"github.com/gavinturner/vinylretailers/util/postgres"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type SKU struct {
	ID         int64     `db:"id" json:"id"`
	Name       string    `json:"name"`
	ReleaseID  int64     `db:"release_id" json:"releaseId"`
	RetailerID int64     `db:"retailer_id" json:"retailerId"`
	ArtistID   int64     `db:"artist_id" json:"artistId"`
	ItemUrl    string    `db:"item_url" json:"itemUrl"`
	ImageUrl   string    `db:"image_url" json:"imageUrl"`
	Price      string    `db:"price" json:"price"`
	CreatedAt  time.Time `db:"created_at" json:"createdAt"`
}

func (v *VinylDB) GetCurrentSKUForRelease(tx *postgres.Tx, releaseID int64, retailerID int64) (*SKU, error) {
	querier := v.Q(tx)
	skus := []SKU{}
	err := querier.Select(&skus, querier.Rebind(`
		SELECT id, retailer_id,  release_id, artist_id, item_url, image_url, price, created_at
		FROM skus
		WHERE retailer_id = ? AND release_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`), retailerID, releaseID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve most recent sku for release")
	}
	if len(skus) == 0 {
		return nil, nil
	}
	return &skus[0], nil
}

func (v *VinylDB) GetAllSKUs(tx *postgres.Tx) ([]SKU, error) {
	querier := v.Q(tx)
	skus := []SKU{}
	err := querier.Select(&skus, querier.Rebind(`
		SELECT id, retailer_id,  release_id, artist_id, item_url, image_url, price, created_at
		FROM skus
	`))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve all skus")
	}
	if len(skus) == 0 {
		return nil, nil
	}
	return skus, nil
}

func (v *VinylDB) UpdateSKU(tx *postgres.Tx, sku *SKU) error {
	querier := v.Q(tx)
	if sku == nil {
		return errors.New("nil sku")
	}
	_, err := querier.Exec(querier.Rebind(`
		UPDATE skus SET item_url=?, image_url=?, price=?
		WHERE id=?
	`), sku.ItemUrl, sku.ImageUrl, sku.Price, sku.ID)
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve all skus")
	}
	return nil
}

func (v *VinylDB) UpsertSKU(tx *postgres.Tx, sku *SKU) (same bool, err error) {
	if sku == nil {
		return false, fmt.Errorf("supplied sku is nil")
	}
	querier := v.Q(tx)

	// first get the current price to see if we should insert a new record
	sku.Price = strings.ToLower(strings.TrimSpace(sku.Price))
	existingSku, err := v.GetCurrentSKUForRelease(tx, sku.ReleaseID, sku.RetailerID)
	if err != nil {
		return false, errors.Wrapf(err, "failed to retrieve existing sku")
	}
	if existingSku != nil && existingSku.Price == sku.Price {
		*sku = *existingSku
		return true, nil
	}
	var id int64
	err = querier.Get(&id, querier.Rebind(`
		INSERT INTO skus (retailer_id, release_id, artist_id, item_url, image_url, price) 
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id
	`), sku.RetailerID, sku.ReleaseID, sku.ArtistID, sku.ItemUrl, sku.ImageUrl, sku.Price)
	if err != nil {
		return false, errors.Wrapf(err, "failed to upsert release")
	}
	sku.ID = id
	return false, nil
}
