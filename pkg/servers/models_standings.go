package servers

type CarPosition struct {
	Type int     `json:"type"`
	Z    float64 `json:"z"`
	Y    float64 `json:"y"`
	X    float64 `json:"x"`
}

type CarVelocity struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Z        float64 `json:"z"`
	Velocity float64 `json:"velocity"`
}

type CarAcceleration struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Z        float64 `json:"z"`
	Velocity float64 `json:"velocity"`
}

type AttackMode struct {
	TotalCount     int     `json:"totalCount"`
	RemainingCount int     `json:"remainingCount"`
	TimeRemaining  float64 `json:"timeRemaining"`
}

type Standing struct {
	SlotID             int             `json:"slotID"`
	DriverName         string          `json:"driverName"`
	VehicleName        string          `json:"vehicleName"`
	LapsCompleted      int             `json:"lapsCompleted"`
	Sector             string          `json:"sector"`
	FinishStatus       string          `json:"finishStatus"`
	LapDistance        float64         `json:"lapDistance"`
	PathLateral        float64         `json:"pathLateral"`
	TrackEdge          float64         `json:"trackEdge"`
	BestSectorTime1    float64         `json:"bestSectorTime1"`
	BestSectorTime2    float64         `json:"bestSectorTime2"`
	BestLapTime        float64         `json:"bestLapTime"`
	LastSectorTime1    float64         `json:"lastSectorTime1"`
	LastSectorTime2    float64         `json:"lastSectorTime2"`
	LastLapTime        float64         `json:"lastLapTime"`
	CurrentSectorTime1 int             `json:"currentSectorTime1"`
	CurrentSectorTime2 int             `json:"currentSectorTime2"`
	Pitstops           int             `json:"pitstops"`
	Penalties          int             `json:"penalties"`
	Player             bool            `json:"player"`
	InControl          int             `json:"inControl"`
	Pitting            bool            `json:"pitting"`
	Position           int             `json:"position"`
	CarClass           string          `json:"carClass"`
	TimeBehindNext     int             `json:"timeBehindNext"`
	LapsBehindNext     int             `json:"lapsBehindNext"`
	TimeBehindLeader   int             `json:"timeBehindLeader"`
	LapsBehindLeader   int             `json:"lapsBehindLeader"`
	LapStartET         float64         `json:"lapStartET"`
	CarPosition        CarPosition     `json:"carPosition"`
	CarVelocity        CarVelocity     `json:"carVelocity"`
	CarAcceleration    CarAcceleration `json:"carAcceleration"`
	Headlights         bool            `json:"headlights"`
	PitState           string          `json:"pitState"`
	ServerScored       bool            `json:"serverScored"`
	GamePhase          string          `json:"gamePhase"`
	Qualification      int             `json:"qualification"`
	TimeIntoLap        float64         `json:"timeIntoLap"`
	EstimatedLapTime   float64         `json:"estimatedLapTime"`
	PitGroup           string          `json:"pitGroup"`
	Flag               string          `json:"flag"`
	UnderYellow        bool            `json:"underYellow"`
	CountLapFlag       string          `json:"countLapFlag"`
	InGarageStall      bool            `json:"inGarageStall"`
	UpgradePack        string          `json:"upgradePack"`
	PitLapDistance     float64         `json:"pitLapDistance"`
	BestLapSectorTime1 float64         `json:"bestLapSectorTime1"`
	BestLapSectorTime2 float64         `json:"bestLapSectorTime2"`
	SteamID            int             `json:"steamID"`
	VehicleFilename    string          `json:"vehicleFilename"`
	CarID              string          `json:"carId"`
	CarNumber          string          `json:"carNumber"`
	FullTeamName       string          `json:"fullTeamName"`
	HasFocus           bool            `json:"hasFocus"`
	FuelFraction       float64         `json:"fuelFraction"`
	AttackMode         AttackMode      `json:"attackMode"`
	DrsActive          bool            `json:"drsActive"`
	Focus              bool            `json:"focus"`
}
