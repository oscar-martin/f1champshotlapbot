package servers

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
	Online             bool           `json:"online,omitempty"`
	ServerID           string         `json:"serverID,omitempty"`
	TrackName          string         `json:"trackName"`
	Session            string         `json:"session"`
	CurrentEventTime   int            `json:"currentEventTime"`
	EndEventTime       int            `json:"endEventTime"`
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
	DarkCloud          int            `json:"darkCloud"`
	Raining            int            `json:"raining"`
	AmbientTemp        float64        `json:"ambientTemp"`
	TrackTemp          float64        `json:"trackTemp"`
	WindSpeed          WindSpeed      `json:"windSpeed"`
	MinPathWetness     int            `json:"minPathWetness"`
	AveragePathWetness int            `json:"averagePathWetness"`
	MaxPathWetness     int            `json:"maxPathWetness"`
	GameMode           string         `json:"gameMode"`
	PasswordProtected  bool           `json:"passwordProtected"`
	ServerPort         int            `json:"serverPort"`
	MaxPlayers         int            `json:"maxPlayers"`
	ServerName         string         `json:"serverName"`
	StartEventTime     int            `json:"startEventTime"`
	RaceCompletion     RaceCompletion `json:"raceCompletion"`
	// added by me
	TimeToEnd string `json:"timeToEnd"`
}
