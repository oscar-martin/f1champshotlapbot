package servers

type StandingsHistory map[string][]DriverInfo

type DriverInfo struct {
	Position     int     `json:"position"`
	DriverName   string  `json:"driverName"`
	SectorTime2  float64 `json:"sectorTime2"`
	LapTime      float64 `json:"lapTime"`
	SectorTime1  float64 `json:"sectorTime1"`
	SlotID       int     `json:"slotID"`
	Pitting      bool    `json:"pitting"`
	CarClass     string  `json:"carClass"`
	FinishStatus string  `json:"finishStatus"`
	VehicleName  string  `json:"vehicleName"`
	TotalLaps    int     `json:"totalLaps"`
}
