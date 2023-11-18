package servers

type DriversSession struct {
	ServerName string          `json:"serverName"`
	ServerID   string          `json:"serverId"`
	Drivers    []DriverSession `json:"drivers"`
}

type DriverSession struct {
	Driver           string  `json:"driver"`
	S1               float64 `json:"s1"`
	S2               float64 `json:"s2"`
	S3               float64 `json:"s3"`
	Time             float64 `json:"time"`
	CarType          string  `json:"carType"`
	CarClass         string  `json:"carClass"`
	Team             string  `json:"team"`
	Compound         string  `json:"compound"`
	Lapcount         int     `json:"lapcount"`
	Lapcountcomplete int     `json:"lapcountcomplete"`
	S1InBestLap      float64 `json:"s1InBestLap"`
	S2InBestLap      float64 `json:"s2InBestLap"`
	S3InBestLap      float64 `json:"s3InBestLap"`
	BestLap          float64 `json:"bestLap"`
	BestS1           float64 `json:"bestS1"`
	BestS2           float64 `json:"bestS2"`
	BestS3           float64 `json:"bestS3"`
	OptimumLap       float64 `json:"optimumLap"`
	MaxSpeed         float64 `json:"maxSpeed"`
}
