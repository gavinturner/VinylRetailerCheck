package redis

type ScanRequest struct {
	RetailerID   int64  `json:"retailerId"`
	RetailerName string `json:"retailerName"`
	ArtistID     int64  `json:"artistId"`
	ArtistName   string `json:"artistName"`
}
