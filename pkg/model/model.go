package model

import "fmt"

type LiveStandingData struct {
	ServerName string               `json:"serverName"`
	ServerID   string               `json:"serverId"`
	Drivers    []StandingDriverData `json:"drivers"`
}

type LiveStandingHistoryData struct {
	ServerName  string                                 `json:"serverName"`
	ServerID    string                                 `json:"serverId"`
	DriverNames []string                               `json:"driverNames"`
	DriversData map[string][]StandingHistoryDriverData `json:"driversData"`
}

type LiveSessionInfoData struct {
	ServerName  string      `json:"serverName"`
	ServerID    string      `json:"serverId"`
	SessionInfo SessionInfo `json:"sessionInfo"`
}

type StandingHistoryDriverData struct {
	Position     int     `json:"position"`
	DriverName   string  `json:"driverName"`
	SlotID       int     `json:"slotID"`
	LapTime      float64 `json:"lapTime"`
	SectorTime1  float64 `json:"sectorTime1"`
	SectorTime2  float64 `json:"sectorTime2"`
	TotalLaps    float64 `json:"totalLaps"`
	VehicleName  string  `json:"vehicleName"`
	FinishStatus string  `json:"finishStatus"`
	Pitting      bool    `json:"pitting"`
	CarClass     string  `json:"carClass"`
	TopSpeed     float64 `json:"topSpeed"` // synthetic field
	CarId        string  `json:"carId"`    // synthetic field
}

type StandingDriverData struct {
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
	BestSectorTime3    float64         `json:"bestSectorTime3"` // synthetic field
	TopSpeedPerLap     map[int]float64 `json:"TopSpeedPerLap"`  // synthetic field
	BestLap            int             `json:"BestLap"`         // synthetic field
	BestLapTime        float64         `json:"bestLapTime"`
	LastSectorTime1    float64         `json:"lastSectorTime1"`
	LastSectorTime2    float64         `json:"lastSectorTime2"`
	LastLapTime        float64         `json:"lastLapTime"`
	CurrentSectorTime1 float64         `json:"currentSectorTime1"`
	CurrentSectorTime2 float64         `json:"currentSectorTime2"`
	Pitstops           int             `json:"pitstops"`
	Penalties          int             `json:"penalties"`
	Player             bool            `json:"player"`
	InControl          int             `json:"inControl"`
	Pitting            bool            `json:"pitting"`
	Position           int             `json:"position"`
	CarClass           string          `json:"carClass"`
	TimeBehindNext     float64         `json:"timeBehindNext"`
	LapsBehindNext     float64         `json:"lapsBehindNext"`
	TimeBehindLeader   float64         `json:"timeBehindLeader"`
	LapsBehindLeader   float64         `json:"lapsBehindLeader"`
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

type CarPosition struct {
	Type            int     `json:"type"`
	Z               float64 `json:"z"`
	Y               float64 `json:"y"`
	X               float64 `json:"x"`
	DriverShortName string  `json:"dri,omitempty"`
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

type WindSpeed struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Z        float64 `json:"z"`
	Velocity float64 `json:"velocity"`
}

type RaceCompletion struct {
	LapsCompletion float64 `json:"lapsCompletion"`
}

type SessionInfo struct {
	WebSocketRunning   bool           `json:"wsRunning,omitempty"`
	ReceivingData      bool           `json:"receivingData,omitempty"`
	LiveMapDomain      string         `json:"liveMapDomain,omitempty"`
	LiveMapPath        string         `json:"liveMapPath,omitempty"`
	TrackName          string         `json:"trackName"`
	Session            string         `json:"session"`
	CurrentEventTime   float64        `json:"currentEventTime"`
	EndEventTime       float64        `json:"endEventTime"`
	MaximumLaps        int            `json:"maximumLaps"`
	LapDistance        float64        `json:"lapDistance"`
	NumberOfVehicles   int            `json:"numberOfVehicles"`
	GamePhase          int            `json:"gamePhase"`
	YellowFlagState    string         `json:"yellowFlagState"`
	SectorFlag         []string       `json:"sectorFlag"`
	StartLightFrame    int            `json:"startLightFrame"`
	NumRedLights       int            `json:"numRedLights"`
	InRealtime         bool           `json:"inRealtime"`
	PlayerName         string         `json:"playerName"`
	PlayerFileName     string         `json:"playerFileName"`
	DarkCloud          float64        `json:"darkCloud"`
	Raining            float64        `json:"raining"`
	AmbientTemp        float64        `json:"ambientTemp"`
	TrackTemp          float64        `json:"trackTemp"`
	WindSpeed          WindSpeed      `json:"windSpeed"`
	MinPathWetness     float64        `json:"minPathWetness"`
	AveragePathWetness float64        `json:"averagePathWetness"`
	MaxPathWetness     float64        `json:"maxPathWetness"`
	GameMode           string         `json:"gameMode"`
	PasswordProtected  bool           `json:"passwordProtected"`
	ServerPort         int            `json:"serverPort"`
	MaxPlayers         int            `json:"maxPlayers"`
	ServerName         string         `json:"serverName"`
	StartEventTime     float64        `json:"startEventTime"`
	RaceCompletion     RaceCompletion `json:"raceCompletion"`
}

type ServerStarted struct {
	ServerName  string  `json:"serverName"`
	ServerID    string  `json:"serverId"`
	SessionType string  `json:"sessionType"`
	TrackName   string  `json:"trackName"`
	EventTime   float64 `json:"eventTime"`
}

func (ss ServerStarted) String() string {
	return fmt.Sprintf("  ▸ Servidor: %s\n  ▸ Sesión: %s\n  ▸ Circuito: %s", ss.ServerName, ss.SessionType, ss.TrackName)
}

// Series struct represents the "series" part of the JSON.
type Series struct {
	ShortName   string `json:"shortName"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Disabled    bool   `json:"disabled"`
	Signature   string `json:"signature"`
	Version     string `json:"version"`
}

// Track struct represents the "track" part of the JSON.
type Track struct {
	ID                    string                 `json:"id"`
	ShortName             string                 `json:"shortName"`
	Name                  string                 `json:"name"`
	SceneDesc             string                 `json:"sceneDesc"`
	Year                  string                 `json:"year"`
	Layout                string                 `json:"layout"`
	Description           string                 `json:"description"`
	Length                string                 `json:"length"`
	Type                  string                 `json:"type"`
	Localizations         map[string]interface{} `json:"localizations"`
	CategoryLocalizations map[string]interface{} `json:"categoryLocalizations"`
	PremID                int                    `json:"premId"`
	Owned                 bool                   `json:"owned"`
	Image                 string                 `json:"image"`
	Thumbnail             string                 `json:"thumbnail"`
}

// Car struct represents the "car" part of the JSON.
type Car struct {
	ID                    string                 `json:"id"`
	Name                  string                 `json:"name"`
	Description           string                 `json:"description"`
	BHP                   string                 `json:"bhp"`
	UsedIn                string                 `json:"usedIn"`
	Configuration         string                 `json:"configuration"`
	FullPathTree          string                 `json:"fullPathTree"`
	VehFile               string                 `json:"vehFile"`
	Engine                string                 `json:"engine"`
	Manufacturer          string                 `json:"manufacturer"`
	Localizations         map[string]interface{} `json:"localizations"`
	CategoryLocalizations map[string]interface{} `json:"categoryLocalizations"`
	PremID                int                    `json:"premId"`
	Owned                 bool                   `json:"owned"`
	Image                 string                 `json:"image"`
	Thumbnail             string                 `json:"thumbnail"`
}

// Data struct represents the entire JSON structure.
type SelectedSessionData struct {
	Series Series `json:"series"`
	Track  Track  `json:"track"`
	Car    Car    `json:"car"`
}
