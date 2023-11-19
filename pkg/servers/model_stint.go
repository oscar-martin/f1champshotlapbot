package servers

type StintData struct {
	ServerName string                 `json:"serverName"`
	ServerID   string                 `json:"serverId"`
	Drivers    map[string]DriverStint `json:"drivers"`
}

type LapTime struct {
	LapTime  float64 `json:"lapTime"`
	S1       float64 `json:"s1"`
	S2       float64 `json:"s2"`
	S3       float64 `json:"s3"`
	MaxSpeed float64 `json:"maxSpeed"`
	Diff     float64 `json:"diff"`
}

type DriverStint struct {
	Driver     string    `json:"driver"`
	Laps       []LapTime `json:"laps"`
	BestLap    LapTime   `json:"bestLap"`
	OptimumLap LapTime   `json:"optimumLap"`
	CarType    string    `json:"carType"`
	CarClass   string    `json:"carClass"`
	Team       string    `json:"team"`
}
