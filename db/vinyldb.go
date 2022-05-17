package db

import "github.com/gavinturner/vinylretailers/util/postgres"

// generate an interface from our persistor implementation
//go:generate ifacemaker --sort=true -f "*.go" -s VinylDB -i VinylDS -p db -o vinylds.go
//go:generate goimports -w vinylds.go
//go:generate sed -i -e  /null.\"gopkg\.in\/guregu\/null\.v3\"/d vinylds.go
// Generate a mock for our new interface
//go:generate rm -f vinylds_mock.go
//go:generate moq -out vinylds_mock.go . VinylDS

type VinylDB struct {
	db *postgres.DB
}

func NewDB(db *postgres.DB) VinylDB {
	return VinylDB{
		db: db,
	}
}

func (v *VinylDB) RetrieveArtists() error {
	return nil
}
