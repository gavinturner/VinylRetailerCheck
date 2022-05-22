package redis

type ScanRequest struct {
	BatchID      int64  `json:"batchId"`
	RetailerID   int64  `json:"retailerId"`
	RetailerName string `json:"retailerName"`
	ArtistID     int64  `json:"artistId"`
	ArtistName   string `json:"artistName"`
}
