package redis

type ScanRequest struct {
	RetailerID int64  `json:"retailerId"`
	ArtistID   int64  `json:"artistId"`
	ArtistName string `json:"artistId"`
}
