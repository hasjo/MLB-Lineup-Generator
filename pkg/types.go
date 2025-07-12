package pkg

import ()

type ConfigData struct {
	WatchTeams  []string
	ReportPath  string
	ReceiptPath string
	PagePath    string
}

type GameLink struct {
	Matchup     string
	FileMatchup string
	Link        string
	PK          int
}

type Schedule struct {
	TotalGamesInProgress int
	Dates                []Date
	Copyright            string
	TotalItems           int
	TotalEvents          int
	TotalGames           int
}

type Date struct {
	Date   string
	Games  []Game
	Events []string
}

type Game struct {
	Content      GameContent
	GameDate     string
	GamePk       int
	Link         string
	OfficialDate string
	Status       GameStatus
	Teams        GameTeams
}

type WeatherData struct {
	Condition string
	Temp      string
}

type VenueData struct {
	Name     string
	Location VenueLocation
}

type VenueLocation struct {
	City        string
	StateAbbrev string
}

type GameContent struct {
	Link string
}

type GameStatus struct {
	AbstractGameState string
	StatusCode        string
}

type GameTeams struct {
	Away GameTeamInfo
	Home GameTeamInfo
}

type GameTeamInfo struct {
	Team Team
}

type Team struct {
	Id     int
	Link   string
	Name   string
	Record GameTeamRecord
}

type GameTeamRecord struct {
	Wins   int
	Losses int
}

type LiveGame struct {
	GameData LiveGameData
	LiveData LiveLiveData
}

type LiveGameData struct {
	Status   LiveGameDataStatus
	Datetime LiveGameDataDatetime
	Venue    VenueData
	Weather  WeatherData
	Teams    LiveGameTeams
}

type LiveGameTeams struct {
	Away Team
	Home Team
}

type LiveGameDataDatetime struct {
	Time         string
	Ampm         string
	OfficialDate string
}

type LiveGameDataStatus struct {
	AbstractGameState string
}

type LiveLiveData struct {
	Boxscore LiveDataBoxscore
}

type LiveDataBoxscore struct {
	Teams     LiveDataTeams
	Officials []LiveDataOfficial
}

type LiveDataTeams struct {
	Away LiveDataTeam
	Home LiveDataTeam
}

type LiveDataTeam struct {
	Team         LiveDataTeamInfo
	Players      map[string]LiveDataTeamPlayers
	BattingOrder []int
	Bullpen      []int
	Bench        []int
	Pitchers     []int
}

type LiveDataTeamInfo struct {
	Name string
}

type LiveDataTeamPlayers struct {
	Person       LiveDataPersonInfo
	JerseyNumber string
	Position     LiveDataPlayerPosition
	BattingOrder string
}

type LiveDataPersonInfo struct {
	Id       int
	FullName string
	Link     string
}

type PeopleInfo struct {
	People []PlayerData
}

type PlayerData struct {
	FullName      string
	PrimaryNumber string
	PitchHand     PlayerPitchInfo
}

type PlayerPitchInfo struct {
	Code string
}

type LiveDataPlayerPosition struct {
	Name         string
	Abbreviation string
}

type LiveDataOfficial struct {
	Official     LiveDataOfficialInfo
	OfficialType string
}

type LiveDataOfficialInfo struct {
	Id       int
	FullName string
}

type StartingList struct {
	TeamName     string
	BattingOrder [9]BatOrderInfo
	Bullpen      BullpenList
	Bench        BenchList
	Pitcher      BullpenInfo
	OK           bool
}

type BullpenList struct {
	TeamName string
	Bullpen  []BullpenInfo
	OK       bool
}

type BullpenInfo struct {
	Name   string
	Number string
	Handed string
	OK     bool
}

type BenchList struct {
	TeamName string
	Bench    []BenchInfo
}

type BenchInfo struct {
	Name string
}

type BatOrderInfo struct {
	Position string
	Name     string
}

type Standings struct {
	Structure StandingsStructure
	Records   []StandingsRecord
}

type StandingsStructure struct {
	Sports []StandingsSport
}

type StandingsSport struct {
	Leagues []StandingsLeague
}

type StandingsLeague struct {
	Divisions []StandingsDivision
}

type StandingsDivision struct {
	Id           int
	NameShort    string
	Abbreviation string
}

type StandingsRecord struct {
	StandingsType string
	Division      int
	TeamRecords   []StandingsTeam
}

type StandingsTeam struct {
	Abbreviation      string
	DivisionGamesBack string
}

type StandingsData struct {
	ALWest    DivisionStandings
	ALCentral DivisionStandings
	ALEast    DivisionStandings
	NLWest    DivisionStandings
	NLCentral DivisionStandings
	NLEast    DivisionStandings
	OK        bool
}

type DivisionStandings struct {
	standings [5]StandingsTeam
}

type ReportData struct {
	ReceiptData string
	PageData    string
	Message     string
	Filename    string
	Live        bool
	OK          bool
}

type Officials struct {
	Home   string
	First  string
	Second string
	Third  string
}
