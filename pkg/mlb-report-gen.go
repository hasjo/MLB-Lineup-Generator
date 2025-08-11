package pkg

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kirsle/configdir"
	"golang.org/x/exp/slices"
)

const BaseLinksURL = "https://statsapi.mlb.com"

// This is a var so it can be mocked over in testing
var GetURLBody = func(targetURL string) ([]byte, bool) {
	resp, err := http.Get(targetURL)
	if err != nil {
		log.Printf("Can't retrieve %s\n", targetURL)
		return []byte{}, false
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("There was a problem reading the body from %s\n", targetURL)
		return []byte{}, false
	}
	return body, true

}

func FindGameLinks(config ConfigData, baseURL string) []GameLink {
	var returnLinks []GameLink
	MonitoredTeams := config.WatchTeams
	url := fmt.Sprintf("%s/api/v1/schedule?sportId=1", baseURL)
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Can't retrieve schedule endpoint, retrying in a moment...")
		return []GameLink{}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	var ScheduleResponse Schedule
	err = json.Unmarshal(body, &ScheduleResponse)
	if err != nil {
		log.Fatal("Failed to unmarshal schedule information = ", err)
	}
	for _, game := range ScheduleResponse.Dates[0].Games {
		awayMatchup := strings.ReplaceAll(game.Teams.Away.Team.Name, " ", "-")
		homeMatchup := strings.ReplaceAll(game.Teams.Home.Team.Name, " ", "-")
		filename := fmt.Sprintf(
			"%s-%s-%d.pdf",
			ScheduleResponse.Dates[0].Date,
			fmt.Sprintf("%s-at-%s", awayMatchup, homeMatchup),
			game.GamePk,
		)
		filepath := fmt.Sprintf("%s%s", config.ReportPath, filename)
		away := game.Teams.Away.Team.Name
		home := game.Teams.Home.Team.Name
		if slices.Contains(MonitoredTeams, away) || slices.Contains(MonitoredTeams, home) {
			matchup := fmt.Sprintf("%s @ %s", away, home)
			link := fmt.Sprintf("%s%s", baseURL, game.Link)
			addLink := GameLink{
				Matchup:     matchup,
				FileMatchup: filename,
				Link:        link,
				PK:          game.GamePk,
			}
			if _, err := os.Stat(filepath); errors.Is(err, os.ErrNotExist) {
				returnLinks = append(returnLinks, addLink)
			}
		}
	}
	return returnLinks
}

func findMaxBattingOrderLength(bo [9]BatOrderInfo) int {
	var maxLength int
	for _, player := range bo {
		playerString := fmt.Sprintf("%2s - %2s - %s",
			player.Position,
			player.JerseyNumber,
			player.Name)
		if len(playerString) > maxLength {
			maxLength = len(playerString)
		}
	}
	return maxLength

}

func findMaxBenchLength(bench []BenchInfo) int {
	var maxLength int
	for _, pitcher := range bench {
		if len(pitcher.Name) > maxLength {
			maxLength = len(pitcher.Name)
		}
	}
	return maxLength

}

func findMaxPitcherLength(bullpen BullpenList) int {
	var maxLength int
	for _, pitcher := range bullpen.Bullpen {
		playerString := fmt.Sprintf("%2s - %2s - %2s - %s",
			"P",
			pitcher.Handed,
			pitcher.Number,
			pitcher.Name)
		if len(playerString) > maxLength {
			maxLength = len(playerString)
		}
	}
	return maxLength
}

func PrettyPrintStartingOrderReceipt(team StartingList) string {
	var returnString string
	for _, player := range team.BattingOrder {
		returnString += fmt.Sprintf("%2s - %2s - %s\n",
			player.Position,
			player.JerseyNumber,
			player.Name)
	}
	returnString += fmt.Sprintf("%2s - %2s - %2s - %s\n",
		"P",
		team.Pitcher.Handed,
		team.Pitcher.Number,
		team.Pitcher.Name,
	)
	return returnString
}

func PrettyPrintStartingOrder(awayTeam StartingList, awayMaxName int, homeTeam StartingList, homeMaxName int) string {
	var returnString string
	for ind, player := range awayTeam.BattingOrder {
		homePlayer := homeTeam.BattingOrder[ind]
		returnString += fmt.Sprintf("%2s - %2s - %-*s|",
			player.Position,
			player.JerseyNumber,
			awayMaxName-9,
			player.Name)
		returnString += fmt.Sprintf(" %2s - %2s - %-*s\n",
			homePlayer.Position,
			homePlayer.JerseyNumber,
			homeMaxName-9,
			homePlayer.Name)
	}
	returnString += fmt.Sprintf(
		"%2s - %2s - %2s - %-*s | %2s - %2s - %2s - %-*s\n",
		"P",
		awayTeam.Pitcher.Handed,
		awayTeam.Pitcher.Number,
		awayMaxName-15,
		awayTeam.Pitcher.Name,
		"P",
		homeTeam.Pitcher.Handed,
		homeTeam.Pitcher.Number,
		homeMaxName-15,
		homeTeam.Pitcher.Name,
	)
	return returnString
}

func PrettyPrintBullpenReceipt(team StartingList) string {
	var returnString string
	returnString += "---BULLPEN\n"
	for _, pitcher := range team.Bullpen.Bullpen {
		returnString += fmt.Sprintf("%2s - %2s - %s\n",
			pitcher.Handed,
			pitcher.Number,
			pitcher.Name,
		)
	}
	return returnString
}

func PrettyPrintBullpen(awayTeam StartingList, awayMaxName int, homeTeam StartingList, homeMaxName int) string {
	var returnString string
	returnString += fmt.Sprintf("%-*s|%-*s\n",
		awayMaxName+1,
		"---BULLPEN",
		homeMaxName+1,
		"---BULLPEN",
	)
	awaySize := len(awayTeam.Bullpen.Bullpen)
	homeSize := len(homeTeam.Bullpen.Bullpen)
	biggerSize := awaySize
	if homeSize > biggerSize {
		biggerSize = homeSize
	}
	for ind := range biggerSize {
		if ind >= awaySize {
			returnString += fmt.Sprintf(
				"%-*s | %2s - %2s - %-*s\n",
				awayMaxName,
				"",
				homeTeam.Bullpen.Bullpen[ind].Handed,
				homeTeam.Bullpen.Bullpen[ind].Number,
				homeMaxName-10,
				homeTeam.Bullpen.Bullpen[ind].Name,
			)

		} else if ind >= homeSize {
			returnString += fmt.Sprintf(
				"%2s - %2s - %-*s | %-*s\n",
				awayTeam.Bullpen.Bullpen[ind].Handed,
				awayTeam.Bullpen.Bullpen[ind].Number,
				awayMaxName-10,
				awayTeam.Bullpen.Bullpen[ind].Name,
				homeMaxName,
				"",
			)
		} else {
			returnString += fmt.Sprintf(
				"%2s - %2s - %-*s | %2s - %2s - %-*s\n",
				awayTeam.Bullpen.Bullpen[ind].Handed,
				awayTeam.Bullpen.Bullpen[ind].Number,
				awayMaxName-10,
				awayTeam.Bullpen.Bullpen[ind].Name,
				homeTeam.Bullpen.Bullpen[ind].Handed,
				homeTeam.Bullpen.Bullpen[ind].Number,
				homeMaxName-10,
				homeTeam.Bullpen.Bullpen[ind].Name,
			)
		}
	}
	return returnString
}

func PrettyPrintBenchReceipt(team StartingList) string {
	var returnString string
	returnString += "---BENCH\n"
	for _, player := range team.Bench.Bench {
		returnString += fmt.Sprintf("%s - %s\n", player.Number, player.Name)
	}
	return returnString
}

func PrettyPrintBench(awayTeam StartingList, awayMaxName int, homeTeam StartingList, homeMaxName int) string {
	var returnString string
	returnString += fmt.Sprintf("%-*s|%-*s\n",
		awayMaxName+1,
		"---BENCH",
		homeMaxName+1,
		"---BENCH",
	)
	awaySize := len(awayTeam.Bench.Bench)
	homeSize := len(homeTeam.Bench.Bench)
	biggerSize := awaySize
	if homeSize > biggerSize {
		biggerSize = homeSize
	}
	for ind := range biggerSize {
		if ind >= awaySize {
			returnString += fmt.Sprintf(
				"%-*s | %2s - %-*s\n",
				awayMaxName-5,
				"",
				homeTeam.Bench.Bench[ind].Number,
				homeMaxName,
				homeTeam.Bench.Bench[ind].Name,
			)

		} else if ind >= homeSize {
			returnString += fmt.Sprintf(
				"%2s - %-*s | %-*s\n",
				awayTeam.Bench.Bench[ind].Number,
				awayMaxName-5,
				awayTeam.Bench.Bench[ind].Name,
				homeMaxName,
				"",
			)
		} else {
			returnString += fmt.Sprintf(
				"%2s - %-*s | %2s - %-*s\n",
				awayTeam.Bench.Bench[ind].Number,
				awayMaxName-5,
				awayTeam.Bench.Bench[ind].Name,
				homeTeam.Bench.Bench[ind].Number,
				homeMaxName,
				homeTeam.Bench.Bench[ind].Name,
			)
		}
	}
	return returnString
}

func PrettyPrintTeams(awayTeam StartingList, homeTeam StartingList, live LiveGame) string {
	var returnString string
	awayName := fmt.Sprintf("----- %s -----", awayTeam.TeamName)
	homeName := fmt.Sprintf("----- %s -----", homeTeam.TeamName)
	returnString += fmt.Sprintf("%s %s @ %s %s\n%s - %s\n%s - %s - %s\n\n",
		awayTeam.TeamName,
		fmt.Sprintf("%d-%d",
			live.GameData.Teams.Away.Record.Wins,
			live.GameData.Teams.Away.Record.Losses,
		),
		homeTeam.TeamName,
		fmt.Sprintf("%d-%d",
			live.GameData.Teams.Home.Record.Wins,
			live.GameData.Teams.Home.Record.Losses,
		),
		live.GameData.Venue.Name,
		fmt.Sprintf("%s, %s", live.GameData.Venue.Location.City, live.GameData.Venue.Location.StateAbbrev),
		live.GameData.Datetime.OfficialDate,
		live.GameData.Datetime.Time,
		fmt.Sprintf("%sf, %s", live.GameData.Weather.Temp, live.GameData.Weather.Condition),
	)
	awayMaxName := len(awayName)
	homeMaxName := len(homeName)
	awayBOMaxLength := findMaxBattingOrderLength(awayTeam.BattingOrder)
	homeBOMaxLength := findMaxBattingOrderLength(homeTeam.BattingOrder)
	if awayBOMaxLength > awayMaxName {
		awayMaxName = awayBOMaxLength
	}
	if homeBOMaxLength > homeMaxName {
		homeMaxName = homeBOMaxLength
	}
	awayStartingPitcherMax := findMaxPitcherLength(
		BullpenList{
			TeamName: awayTeam.TeamName,
			Bullpen:  []BullpenInfo{awayTeam.Pitcher},
		})
	if awayStartingPitcherMax > awayMaxName {
		awayMaxName = awayStartingPitcherMax
	}
	homeStartingPitcherMax := findMaxPitcherLength(
		BullpenList{
			TeamName: homeTeam.TeamName,
			Bullpen:  []BullpenInfo{homeTeam.Pitcher},
		})
	if homeStartingPitcherMax > homeMaxName {
		homeMaxName = homeStartingPitcherMax
	}
	awayPitcherMax := findMaxPitcherLength(awayTeam.Bullpen)
	if awayPitcherMax > awayMaxName {
		awayMaxName = awayPitcherMax
	}
	homePitcherMax := findMaxPitcherLength(homeTeam.Bullpen)
	if homePitcherMax > homeMaxName {
		homeMaxName = homePitcherMax
	}
	awayBenchMax := findMaxBenchLength(awayTeam.Bench.Bench)
	if awayBenchMax > awayMaxName {
		awayMaxName = awayBenchMax
	}
	homeBenchMax := findMaxBenchLength(homeTeam.Bench.Bench)
	if homeBenchMax > homeMaxName {
		homeMaxName = homeBenchMax
	}
	returnString += fmt.Sprintf("%*s | %*s\n",
		awayMaxName,
		awayName,
		homeMaxName,
		homeName)
	returnString += PrettyPrintStartingOrder(
		awayTeam, awayMaxName,
		homeTeam, homeMaxName)
	returnString += PrettyPrintBullpen(
		awayTeam, awayMaxName,
		homeTeam, homeMaxName)
	returnString += PrettyPrintBench(
		awayTeam, awayMaxName,
		homeTeam, homeMaxName)
	returnString += "\n"
	return returnString
}

func PrettyPrintTeamsReceipt(awayTeam StartingList, homeTeam StartingList, live LiveGame) string {
	var returnString string
	awayName := fmt.Sprintf("----- %s -----\n", awayTeam.TeamName)
	homeName := fmt.Sprintf("\n----- %s -----\n", homeTeam.TeamName)
	returnString += fmt.Sprintf("%s - %s\n@\n%s - %s\n%s\n%s\n%s - %s\n%s\n\n",
		awayTeam.TeamName,
		fmt.Sprintf("%d-%d",
			live.GameData.Teams.Away.Record.Wins,
			live.GameData.Teams.Away.Record.Losses,
		),
		homeTeam.TeamName,
		fmt.Sprintf("%d-%d",
			live.GameData.Teams.Home.Record.Wins,
			live.GameData.Teams.Home.Record.Losses,
		),
		live.GameData.Venue.Name,
		fmt.Sprintf("%s, %s", live.GameData.Venue.Location.City, live.GameData.Venue.Location.StateAbbrev),
		live.GameData.Datetime.OfficialDate,
		live.GameData.Datetime.Time,
		fmt.Sprintf("%sf, %s", live.GameData.Weather.Temp, live.GameData.Weather.Condition),
	)
	returnString += awayName
	returnString += PrettyPrintStartingOrderReceipt(awayTeam)
	returnString += PrettyPrintBullpenReceipt(awayTeam)
	returnString += PrettyPrintBenchReceipt(awayTeam)
	returnString += homeName
	returnString += PrettyPrintStartingOrderReceipt(homeTeam)
	returnString += PrettyPrintBullpenReceipt(homeTeam)
	returnString += PrettyPrintBenchReceipt(homeTeam)
	return returnString
}
func GetPitcherInformation(infoURL string) BullpenInfo {
	baseURL := "https://statsapi.mlb.com"
	URL := baseURL + infoURL
	resp, err := http.Get(URL)
	if err != nil {
		log.Println("Unable to get Pitcher Info... Will try again later. Error:", err)
		return BullpenInfo{OK: false}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	var PitcherResponse PeopleInfo
	err = json.Unmarshal(body, &PitcherResponse)
	PitcherData := PitcherResponse.People[0]
	return BullpenInfo{
		Name:   PitcherData.FullName,
		Number: PitcherData.PrimaryNumber,
		Handed: PitcherData.PitchHand.Code,
		OK:     true,
	}

}

func GenerateStartingList(inTeam LiveDataTeam) StartingList {
	var returnList StartingList
	returnList.TeamName = inTeam.Team.Name
	Order := inTeam.BattingOrder
	for ind, playerId := range Order {
		IDString := fmt.Sprintf("ID%d", playerId)
		PlayerItem := inTeam.Players[IDString]
		returnList.BattingOrder[ind] = BatOrderInfo{
			Position:     PlayerItem.Position.Abbreviation,
			Name:         PlayerItem.Person.FullName,
			JerseyNumber: PlayerItem.JerseyNumber,
		}
	}
	StartingPitcher := inTeam.Pitchers[0]
	IDString := fmt.Sprintf("ID%d", StartingPitcher)
	PitcherItem := inTeam.Players[IDString]
	returnList.Pitcher = GetPitcherInformation(PitcherItem.Person.Link)
	if returnList.Pitcher.OK == false {
		returnList.OK = false
	} else {
		returnList.OK = true
	}
	return returnList
}

func GenerateBullpen(inTeam LiveDataTeam) BullpenList {
	var returnList BullpenList
	returnList.OK = true
	returnList.TeamName = inTeam.Team.Name
	Order := inTeam.Bullpen
	for ind, playerId := range Order {
		IDString := fmt.Sprintf("ID%d", playerId)
		PlayerItem := inTeam.Players[IDString]
		returnList.Bullpen = append(returnList.Bullpen, GetPitcherInformation(PlayerItem.Person.Link))
		if returnList.Bullpen[ind].OK == false {
			returnList.OK = false
		}
	}
	return returnList
}

func GenerateBench(inTeam LiveDataTeam) BenchList {
	var returnList BenchList
	returnList.TeamName = inTeam.Team.Name
	Order := inTeam.Bench
	for _, playerId := range Order {
		IDString := fmt.Sprintf("ID%d", playerId)
		PlayerItem := inTeam.Players[IDString]
		addItem := BenchInfo{
			Name:   PlayerItem.Person.FullName,
			Number: PlayerItem.JerseyNumber,
		}
		returnList.Bench = append(returnList.Bench, addItem)
	}
	return returnList
}

func GenerateUmpires(officialSlice []LiveDataOfficial) Officials {
	var retOfficials Officials
	for _, official := range officialSlice {
		switch official.OfficialType {
		case "Home Plate":
			retOfficials.Home = official.Official.FullName
		case "First Base":
			retOfficials.First = official.Official.FullName
		case "Second Base":
			retOfficials.Second = official.Official.FullName
		case "Third Base":
			retOfficials.Third = official.Official.FullName
		}
	}
	return retOfficials
}

func PrettyPrintOfficials(officials Officials) string {
	var returnString string
	returnString += "---OFFICIALS\n"
	returnString += fmt.Sprintf("HOME   - %s\n", officials.Home)
	returnString += fmt.Sprintf("FIRST  - %s\n", officials.First)
	returnString += fmt.Sprintf("SECOND - %s\n", officials.Second)
	returnString += fmt.Sprintf("THIRD  - %s\n", officials.Third)
	return returnString
}

func GenerateStandings() StandingsData {
	var returnData StandingsData
	getURL := "https://bdfed.stitch.mlbinfra.com/bdfed/transform-mlb-standings?" +
		"=&=&splitPcts=false&" +
		"numberPcts=false" +
		"&standingsView=division" +
		"&sortTemplate=3" +
		"&season=2025" +
		"&leagueIds=103" +
		"&leagueIds=104" +
		"&standingsTypes=regularSeason" +
		"&contextTeamId=" +
		"&teamId=" +
		"&hydrateAlias=noSchedule" +
		"&favoriteTeams=sortSports=1"
	resp, err := http.Get(getURL)
	if err != nil {
		log.Println("Failed to get standings... error:", err)
		returnData.OK = false
		return returnData
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	var StandingsResponse Standings
	err = json.Unmarshal(body, &StandingsResponse)
	var DivisionIDs = make(map[int]string)
	for _, league := range StandingsResponse.Structure.Sports[0].Leagues {
		for _, division := range league.Divisions {
			DivisionIDs[division.Id] = division.Abbreviation
		}
	}
	for _, record := range StandingsResponse.Records {
		for ind, teamRecord := range record.TeamRecords {
			switch DivisionIDs[record.Division] {
			case "ALC":
				returnData.ALCentral.standings[ind] = teamRecord
			case "ALW":
				returnData.ALWest.standings[ind] = teamRecord
			case "ALE":
				returnData.ALEast.standings[ind] = teamRecord
			case "NLC":
				returnData.NLCentral.standings[ind] = teamRecord
			case "NLW":
				returnData.NLWest.standings[ind] = teamRecord
			case "NLE":
				returnData.NLEast.standings[ind] = teamRecord
			}
		}
	}
	returnData.OK = true
	return returnData
}

func PrettyPrintStandings(standings StandingsData) string {
	OutLines := "- AL West - | - AL Central - | - AL East -\n" +
		" Team | GB  |  Team  |  GB   | Team  | GB\n"
	for ind := range standings.ALCentral.standings {
		OutLines += fmt.Sprintf("%4s  |%4s |%5s   |%5s  |%4s   |%4s\n",
			standings.ALWest.standings[ind].Abbreviation,
			standings.ALWest.standings[ind].DivisionGamesBack,
			standings.ALCentral.standings[ind].Abbreviation,
			standings.ALCentral.standings[ind].DivisionGamesBack,
			standings.ALEast.standings[ind].Abbreviation,
			standings.ALEast.standings[ind].DivisionGamesBack,
		)
	}
	OutLines += "- NL West - | - NL Central - | - NL East -\n" +
		" Team | GB  |  Team  |  GB   | Team  | GB\n"
	for ind := range standings.NLCentral.standings {
		OutLines += fmt.Sprintf("%4s  |%4s |%5s   |%5s  |%4s   |%4s\n",
			standings.NLWest.standings[ind].Abbreviation,
			standings.NLWest.standings[ind].DivisionGamesBack,
			standings.NLCentral.standings[ind].Abbreviation,
			standings.NLCentral.standings[ind].DivisionGamesBack,
			standings.NLEast.standings[ind].Abbreviation,
			standings.NLEast.standings[ind].DivisionGamesBack,
		)
	}
	return OutLines
}

func GeneratePreGameReport(InLink GameLink, debug bool) ReportData {
	var ReturnReportReceipt string
	var ReturnReportPage string
	var Message string
	getURL := InLink.Link
	resp, err := http.Get(getURL)
	if err != nil {
		log.Println("Unable to get game info... error:", err)
		return ReportData{OK: false}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	var LiveGameResponse LiveGame
	err = json.Unmarshal(body, &LiveGameResponse)
	if err != nil {
		log.Fatal("CANT UNMARSHAL = ", err)
	}
	var filename string
	filename = InLink.FileMatchup
	isLive := LiveGameResponse.GameData.Status.AbstractGameState == "Live"
	if isLive {
		awayTeam := GenerateStartingList(
			LiveGameResponse.LiveData.Boxscore.Teams.Away)
		awayTeam.Bullpen = GenerateBullpen(
			LiveGameResponse.LiveData.Boxscore.Teams.Away)
		awayTeam.Bench = GenerateBench(
			LiveGameResponse.LiveData.Boxscore.Teams.Away)
		if awayTeam.OK == false || awayTeam.Bullpen.OK == false {
			return ReportData{OK: false}
		}
		homeTeam := GenerateStartingList(
			LiveGameResponse.LiveData.Boxscore.Teams.Home)
		homeTeam.Bullpen = GenerateBullpen(
			LiveGameResponse.LiveData.Boxscore.Teams.Home)
		homeTeam.Bench = GenerateBench(
			LiveGameResponse.LiveData.Boxscore.Teams.Home)
		if homeTeam.OK == false || homeTeam.Bullpen.OK == false {
			return ReportData{OK: false}
		}
		ReturnReportReceipt += PrettyPrintTeamsReceipt(awayTeam, homeTeam, LiveGameResponse)
		ReturnReportPage += PrettyPrintTeams(awayTeam, homeTeam, LiveGameResponse)
		officials := GenerateUmpires(LiveGameResponse.LiveData.Boxscore.Officials)
		ReturnReportReceipt += PrettyPrintOfficials(officials)
		ReturnReportPage += PrettyPrintOfficials(officials)
		if debug == true {
			ReturnReportReceipt += getURL
			ReturnReportPage += getURL
		}
	} else {
		Message += fmt.Sprintf("%s - Starts: %s - %s%s\n",
			InLink.Matchup,
			LiveGameResponse.GameData.Datetime.OfficialDate,
			LiveGameResponse.GameData.Datetime.Time,
			LiveGameResponse.GameData.Datetime.Ampm,
		)
		if debug == true {
			Message += getURL
		}
	}
	return ReportData{
		ReceiptData: ReturnReportReceipt,
		PageData:    ReturnReportPage,
		Message:     Message,
		Filename:    filename,
		Live:        isLive,
		OK:          true,
	}
}

func GenerateFullReport(config ConfigData, debug bool) []ReportData {
	var returnData []ReportData
	teams := config.WatchTeams
	if len(teams) == 0 {
		teams = []string{"Minnesota Twins"}
	}
	foundLinks := FindGameLinks(config, BaseLinksURL)
	for _, link := range foundLinks {
		newReport := GeneratePreGameReport(link, debug)
		if newReport.OK {
			returnData = append(returnData, newReport)
		}
	}
	if len(returnData) > 0 {
		standings := GenerateStandings()
		if standings.OK == false {
			return []ReportData{}
		}
		prettyStandings := PrettyPrintStandings(standings)
		for ind := range returnData {
			returnData[ind].ReceiptData += "\n" + prettyStandings
			returnData[ind].PageData += "\n" + prettyStandings
		}
	}
	return returnData
}

func GetOrHandleConfiguration() ConfigData {
	var returnData ConfigData
	configPath := configdir.LocalConfig("mlb-report-gen")
	err := configdir.MakePath(configPath)
	if err != nil {
		log.Fatal("Failed to make the config directory", configPath, "exiting...")
	}
	configFile := filepath.Join(configPath, "config.json")

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Create the new config file.
		returnData = ConfigData{
			WatchTeams:  AllTeams,
			ReportPath:  "",
			ReceiptPath: filepath.Join("", "receipt"),
			PagePath:    filepath.Join("", "page"),
		}
		fh, err := os.Create(configFile)
		if err != nil {
			log.Fatal(err)
		}
		defer fh.Close()

		encoder := json.NewEncoder(fh)
		encoder.SetIndent("", "    ")
		encoder.Encode(&returnData)
		log.Println("Created new log file at:", configFile)
	} else {
		// Load the existing file.
		fh, err := os.Open(configFile)
		if err != nil {
			log.Fatal(err)
		}
		defer fh.Close()
		log.Println("Read log file from:", configFile)

		decoder := json.NewDecoder(fh)
		decoder.Decode(&returnData)
		returnData.ReceiptPath = filepath.Join(returnData.ReportPath, "receipt")
		returnData.PagePath = filepath.Join(returnData.ReportPath, "page")
	}
	CreateReportDirs(returnData)

	return returnData
}

func RunLookup(debug bool, config ConfigData, watchlist map[string]bool) {
	data := GenerateFullReport(config, debug)
	for _, report := range data {
		datapath := filepath.Join(config.ReportPath, report.Filename)
		receiptpath := filepath.Join(config.ReceiptPath, report.Filename)
		pagepath := filepath.Join(config.PagePath, report.Filename)
		if report.Live == true {
			if _, err := os.Stat(receiptpath); errors.Is(err, os.ErrNotExist) {
				fmt.Printf("\n Writing %s\n", receiptpath)
				GenerateReportPDFReceipt(report.ReceiptData, receiptpath, config)
			}
			if _, err := os.Stat(pagepath); errors.Is(err, os.ErrNotExist) {
				fmt.Printf("\n Writing %s\n", pagepath)
				GenerateReportPDF(report.PageData, pagepath, config)
			}
			delete(watchlist, datapath)
		} else {
			_, ok := watchlist[datapath]
			if ok {
				fmt.Print(".")
			} else {
				fmt.Printf("\n Found %s - monitoring...\n",
					strings.SplitN(report.Message, "\n", 2)[0])
				watchlist[datapath] = true
			}
		}
	}
}

func RunLocal() {
	//Setup the config dir
	debugPtr := flag.Bool("debug", false, "Enable debug output")
	flag.Parse()
	debug := *debugPtr
	config := GetOrHandleConfiguration()
	var watchList map[string]bool
	watchList = make(map[string]bool)
	go RunLookup(debug, config, watchList)
	for range time.Tick(time.Second * 60) {
		go RunLookup(debug, config, watchList)
	}
}
