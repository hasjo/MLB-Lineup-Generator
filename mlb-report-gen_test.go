package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func BuildTestServer(expectedURL string, expectedArgs map[string]string, returnData []byte, t *testing.T) *httptest.Server {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != expectedURL {
					t.Error("Expected to hit the proper endpoint, got", r.URL.Path)
				}
				for key, value := range expectedArgs {
					queryValue := r.URL.Query().Get(key)
					if queryValue != value {
						t.Errorf("Expected query args [key: value] [%s: %s], got [%s, %s]",
							key, value,
							key, queryValue)
					}
				}
				w.WriteHeader(http.StatusOK)
				w.Write(returnData)
			},
		),
	)
	return server
}

func CompareStringOutput(in1 string, in2 string) (bool, string, string) {
	onesplits := strings.Split(in1, "\n")
	twosplits := strings.Split(in2, "\n")
	for ind, item := range onesplits {
		if item != twosplits[ind] {
			return false, item, twosplits[ind]
		}
	}
	return true, "", ""
}

func TestGetURLBody(t *testing.T) {
}

func TestFindGameLinks(t *testing.T) {
	// Set up Server Args
	expectedURL := "/api/v1/schedule"
	expectedArgs := make(map[string]string)
	expectedArgs["sportId"] = "1"
	testData, err := os.ReadFile("schedule.json")
	if err != nil {
		t.Error("Hey, can't read test data, look into that please...")
	}

	// Spin up httptest server
	server := BuildTestServer(expectedURL, expectedArgs, testData, t)
	defer server.Close()

	// Define expected data
	expectedData := make(map[string]GameLink)
	expectedData["2025-05-23-Baltimore-OriolesatBoston-Red-Sox-777815.pdf"] = GameLink{
		Matchup:     "Baltimore Orioles @ Boston Red Sox",
		FileMatchup: "2025-05-23-Baltimore-OriolesatBoston-Red-Sox-777815.pdf",
		Link:        fmt.Sprintf("%s/api/v1.1/game/777815/feed/live", server.URL),
		PK:          777815,
	}
	// This day had an Orioles doubleheader
	expectedData["2025-05-23-Baltimore-OriolesatBoston-Red-Sox-777809.pdf"] = GameLink{
		Matchup:     "Baltimore Orioles @ Boston Red Sox",
		FileMatchup: "2025-05-23-Baltimore-OriolesatBoston-Red-Sox-777809.pdf",
		Link:        fmt.Sprintf("%s/api/v1.1/game/777809/feed/live", server.URL),
		PK:          777809,
	}

	// Set up config
	var config ConfigData
	config.WatchTeams = []string{"Baltimore Orioles"}

	// Run the function
	data := FindGameLinks(config, server.URL)

	// Check the output
	for _, item := range data {
		checkData, ok := expectedData[item.FileMatchup]
		if ok {
			if item != checkData {
				t.Errorf("Expected GameLink Mismatch:\n %+v\n %+v", checkData, item)
			}
		}
	}
}

func TestFindMaxBattingOrderLength(t *testing.T) {
	//Build a test object
	testBattingOrder := [9]BatOrderInfo{
		{Position: "C", Name: "Player Longname"},
		{Position: "1B", Name: "Player Longnamer"},
		{Position: "2B", Name: "Player"},
		{Position: "SS", Name: "Player Shortname"},
		{Position: "3B", Name: "Longname"},
		{Position: "LF", Name: "Player Middlename Longname"},
		{Position: "CF", Name: "Player Long"},
		{Position: "RF", Name: "Play Name"},
		{Position: "DH", Name: "Hitting Longname"},
	}
	// run the function
	foundValue := findMaxBattingOrderLength(testBattingOrder)
	// compare the output
	if foundValue != 31 {
		t.Errorf("Max length incorrect: %d", foundValue)
	}
}

func TestFindMaxBenchLength(t *testing.T) {
	// Build a test object
	testBench := []BenchInfo{
		{Name: "guy1"},
		{Name: "guy10"},
		{Name: "guy1000"},
		{Name: "guy100"},
	}
	// run the function
	maxBench := findMaxBenchLength(testBench)

	if maxBench != 7 {
		t.Errorf("Max length incorrect: %d", maxBench)
	}
}

func TestFindMaxPitcherLength(t *testing.T) {
	// Build a test object
	testPitchers := BullpenList{
		Bullpen: []BullpenInfo{
			{Name: "This Pitcher", Number: "12", Handed: "R", OK: true},
			{Name: "That Pitcher", Number: "12", Handed: "R", OK: true},
			{Name: "Another Pitcher", Number: "12", Handed: "R", OK: true},
		},
		TeamName: "TestTeam",
		OK:       true,
	}

	// run the function
	maxPitcher := findMaxPitcherLength(testPitchers)

	if maxPitcher != 30 {
		t.Errorf("Max length incorrect: %d", maxPitcher)
	}
}

func TestPrettyPrintStartingOrderReceipt(t *testing.T) {
	// Build test data
	testData := StartingList{
		BattingOrder: [9]BatOrderInfo{
			{Position: "C", Name: "Catcher"},
			{Position: "1B", Name: "First Base"},
			{Position: "2B", Name: "Second Base"},
			{Position: "SS", Name: "Shortstop"},
			{Position: "3B", Name: "Third Base"},
			{Position: "LF", Name: "Left Field"},
			{Position: "CF", Name: "Center Field"},
			{Position: "RF", Name: "Right Field"},
			{Position: "DH", Name: "Designated Hitter"},
		},
		Pitcher: BullpenInfo{Name: "Pitcher", Number: "12", Handed: "R", OK: true},
	}

	// Expected Output
	ExpectedData := " C - Catcher\n" +
		"1B - First Base\n" +
		"2B - Second Base\n" +
		"SS - Shortstop\n" +
		"3B - Third Base\n" +
		"LF - Left Field\n" +
		"CF - Center Field\n" +
		"RF - Right Field\n" +
		"DH - Designated Hitter\n" +
		" P -  R - 12 - Pitcher\n"

	// run the function
	outputdata := PrettyPrintStartingOrderReceipt(testData)
	if outputdata != ExpectedData {
		t.Errorf("Return data does not match the expected data:\n %s\n %s", ExpectedData, outputdata)
	}
}

func TestPrettyPrintStartingOrder(t *testing.T) {
	// Setup input data
	team := StartingList{
		BattingOrder: [9]BatOrderInfo{
			{Position: "C", Name: "Catcher"},
			{Position: "1B", Name: "First Base"},
			{Position: "2B", Name: "Second Base"},
			{Position: "SS", Name: "Shortstop"},
			{Position: "3B", Name: "Third Base"},
			{Position: "LF", Name: "Left Field"},
			{Position: "CF", Name: "Center Field"},
			{Position: "RF", Name: "Right Field"},
			{Position: "DH", Name: "Designated Hitter"},
		},
		Pitcher: BullpenInfo{Name: "Pitcher", Number: "12", Handed: "R", OK: true},
	}
	maxlength := 22

	expectedOutput := " C - Catcher           |  C - Catcher           \n" +
		"1B - First Base        | 1B - First Base        \n" +
		"2B - Second Base       | 2B - Second Base       \n" +
		"SS - Shortstop         | SS - Shortstop         \n" +
		"3B - Third Base        | 3B - Third Base        \n" +
		"LF - Left Field        | LF - Left Field        \n" +
		"CF - Center Field      | CF - Center Field      \n" +
		"RF - Right Field       | RF - Right Field       \n" +
		"DH - Designated Hitter | DH - Designated Hitter \n" +
		" P -  R - 12 - Pitcher |  P -  R - 12 - Pitcher\n"

	// Run thing
	testOutput := PrettyPrintStartingOrder(team, maxlength, team, maxlength)

	if expectedOutput != testOutput {
		OK, expected, test := CompareStringOutput(expectedOutput, testOutput)
		if OK != false {
			t.Errorf("Mismatched Lines: \n %s\n %s\nExpected and Test Output do not match:\n %s\n%s", expected, test, expectedOutput, testOutput)
		}
		t.Error("failed still")
	}
}

func TestPrettyPrintBullpenReceipt(t *testing.T) {
	// Setup Input Data
	team := StartingList{
		Bullpen: BullpenList{
			TeamName: "Team",
			Bullpen: []BullpenInfo{
				{Name: "Pitcher 1", Number: "12", Handed: "R", OK: true},
				{Name: "Pitcher 2", Number: "22", Handed: "L", OK: true},
				{Name: "Pitcher 3", Number: "2", Handed: "R", OK: true},
				{Name: "Pitcher 4", Number: "17", Handed: "L", OK: true},
				{Name: "Pitcher 5", Number: "6", Handed: "R", OK: true},
			},
			OK: true,
		},
	}

	// Define expectations
	expectedOutput := "---BULLPEN\n" +
		" R - 12 - Pitcher 1\n" +
		" L - 22 - Pitcher 2\n" +
		" R -  2 - Pitcher 3\n" +
		" L - 17 - Pitcher 4\n" +
		" R -  6 - Pitcher 5\n"

	// Run the stuff
	testOutput := PrettyPrintBullpenReceipt(team)

	// Compare

	if testOutput != expectedOutput {
		t.Errorf("Mismatched output: \n%s\n%s\n", expectedOutput, testOutput)
	}
}

func TestPrettyPrintBullpen(t *testing.T) {
	// Setup input data
	team := StartingList{
		Bullpen: BullpenList{
			TeamName: "Team",
			Bullpen: []BullpenInfo{
				{Name: "Pitcher 1", Number: "12", Handed: "R", OK: true},
				{Name: "Pitcher 2", Number: "22", Handed: "L", OK: true},
				{Name: "Pitcher 3", Number: "2", Handed: "R", OK: true},
				{Name: "Pitcher 4", Number: "17", Handed: "L", OK: true},
				{Name: "Pitcher 5", Number: "6", Handed: "R", OK: true},
			},
			OK: true,
		},
	}

	// Set up expected output
	maxlength := 19
	expectedOutput := "---BULLPEN          |---BULLPEN          \n" +
		" R - 12 - Pitcher 1 |  R - 12 - Pitcher 1\n" +
		" L - 22 - Pitcher 2 |  L - 22 - Pitcher 2\n" +
		" R -  2 - Pitcher 3 |  R -  2 - Pitcher 3\n" +
		" L - 17 - Pitcher 4 |  L - 17 - Pitcher 4\n" +
		" R -  6 - Pitcher 5 |  R -  6 - Pitcher 5\n"

	testOutput := PrettyPrintBullpen(team, maxlength, team, maxlength)

	if testOutput != expectedOutput {
		t.Errorf("Mismatched output: \n%s\n%s\n", expectedOutput, testOutput)
	}
}
