package main

import (
	"encoding/json"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"os"
	"strconv"
	"strings"
)

type BattingPlayer struct {
	ID                int `gorm:"primaryKey; not null"`
	Name              string
	BattingStyle      string
	Role              string
	Team              string
	Matches           int
	InningsBatted     int
	RunsScored        int
	BallsFaced        int
	DotBallsPlayed    int
	FoursHit          int
	SixesHit          int
	NotOuts           int
	Fifties           int
	Centuries         int
	HighestScore      string
	BattingStrikeRate *float64
	BattingAverage    *float64
}

type BowlingPlayer struct {
	ID                int `gorm:"primaryKey; not null"`
	Name              string
	BowlingStyle      string
	Role              string
	Team              string
	Matches           int
	InningsBowled     int
	RunsGiven         int
	BallsBowled       int
	Overs             *float64
	DotBallsBowled    int
	FoursGiven        int
	SixesGiven        int
	Wickets           int
	FourWickets       int
	FiveWickets       int
	TenWickets        int
	Maidens           int
	HighestWickets    string
	BowlingStrikeRate *float64
	BowlingAverage    *float64
	EconomyRate       *float64
}

type Match struct {
	ID                int `gorm:"primaryKey; not null"`
	TeamAName         string
	TeamBName         string
	DateTime          string `json:"MatchDateTime"`
	GroundName        string
	City              string
	TossInfo          string
	InningsOneSummary string
	InningsTwoSummary string
	WinTeamName       string
	ManOfTheMatchName string
	Result            string `json:"MatchResult"`
}

type MatchData struct {
	Matches []Match `json:"CompetitionDeatails"`
}

type Team struct {
	ID         int `gorm:"primaryKey; not null"`
	Name       string
	FullName   string
	Matches    int
	Wins       int
	Loss       int
	Points     int
	NetRunRate float64
	Image      string
}

type BatStats struct {
	Batsmen []Batsman `json:"CompetitionPlayerStats"`
}

type Batsman struct {
	ID             string
	PlayerName     string
	BattingStyle   string
	BowlingStyle   string
	PlayerRole     string
	TeamName       string
	Matches        int
	Innings        int
	Runs           int
	Balls          int
	DotBalls       int
	BdryFours      int
	BdrySixes      int
	NotOuts        int
	Fifties        int
	Centuries      int
	HighestScore   string
	StrikeRate     string
	BattingAverage string
}

type BowlStats struct {
	Bowlers []Bowler `json:"CompetitionPlayerStats"`
}

type Bowler struct {
	ID             string
	PlayerName     string
	BattingStyle   string
	BowlingStyle   string
	PlayerRole     string
	TeamName       string
	Matches        int
	Innings        int
	Runs           int
	Balls          int
	Overs          string
	DotBalls       int
	BdryFours      int
	BdrySixes      int
	Wickets        int
	FourWickets    int
	FiveWickets    int
	TenWickets     int
	Maidens        int
	BestWickets    string
	StrikeRate     float64
	BowlingAverage string
	EconomyRate    string
}

func initTeams() []Team {
	return []Team{
		{
			Name:       "SEA",
			FullName:   "Seattle Orcas",
			Matches:    5,
			Wins:       4,
			Loss:       1,
			Points:     8,
			NetRunRate: 0.725,
			Image:      "https://splcms.blob.core.windows.net/mlc/team_logos/sea.png",
		},
		{
			Name:       "TSK",
			FullName:   "Texas Super Kings",
			Matches:    5,
			Wins:       3,
			Loss:       2,
			Points:     6,
			NetRunRate: 0.57,
			Image:      "https://splcms.blob.core.windows.net/mlc/team_logos/tex.png",
		},
		{
			Name:       "WSH",
			FullName:   "Washington Freedom",
			Matches:    5,
			Wins:       3,
			Loss:       2,
			Points:     6,
			NetRunRate: 0.097,
			Image:      "https://splcms.blob.core.windows.net/mlc/team_logos/wsh.png",
		},
		{
			Name:       "MI NY",
			FullName:   "Mumbai Indians New York",
			Matches:    5,
			Wins:       2,
			Loss:       3,
			Points:     4,
			NetRunRate: 1.004,
			Image:      "https://splcms.blob.core.windows.net/mlc/team_logos/ny.png",
		},
		{
			Name:       "SF",
			FullName:   "San Francisco Unicorns",
			Matches:    5,
			Wins:       2,
			Loss:       3,
			Points:     8,
			NetRunRate: -0.303,
			Image:      "https://splcms.blob.core.windows.net/mlc/team_logos/sf.png",
		},
		{
			Name:       "LAKR",
			FullName:   "Los Angeles Knight Riders",
			Matches:    5,
			Wins:       1,
			Loss:       4,
			Points:     2,
			NetRunRate: -2.028,
			Image:      "https://splcms.blob.core.windows.net/mlc/team_logos/la.png",
		},
	}
}

func toTitleCase(str string) string {
	str = strings.ToLower(str)
	words := strings.Fields(str)
	for i, word := range words {
		words[i] = strings.ToUpper(word[:1]) + word[1:]
	}
	return strings.Join(words, " ")
}

func main() {
	batData, err := os.ReadFile("./raw/batsman.json")
	if err != nil {
		log.Fatal(err)
	}
	var bs BatStats
	err = json.Unmarshal(batData, &bs)
	if err != nil {
		log.Fatal(err)
	}

	bowlData, err := os.ReadFile("./raw/bowlers.json")
	if err != nil {
		log.Fatal(err)
	}
	var bos BowlStats
	err = json.Unmarshal(bowlData, &bos)
	if err != nil {
		log.Fatal(err)
	}

	battingPlayers := make([]BattingPlayer, 0)
	for _, b := range bs.Batsmen {
		sr, err := strconv.ParseFloat(b.StrikeRate, 64)
		if err != nil {
			log.Fatal(err)
		}
		strikeRate := &sr

		var average *float64
		if b.BattingAverage == "NA" {
			average = nil
		} else {
			av, err := strconv.ParseFloat(b.BattingAverage, 64)
			if err != nil {
				log.Fatal(err)
			}
			average = &av
		}

		p := BattingPlayer{
			Name:              toTitleCase(b.PlayerName),
			BattingStyle:      b.BattingStyle,
			Role:              b.PlayerRole,
			Team:              b.TeamName,
			Matches:           b.Matches,
			InningsBatted:     b.Innings,
			RunsScored:        b.Runs,
			BallsFaced:        b.Balls,
			DotBallsPlayed:    b.DotBalls,
			FoursHit:          b.BdryFours,
			SixesHit:          b.BdrySixes,
			NotOuts:           b.NotOuts,
			Fifties:           b.Fifties,
			Centuries:         b.Centuries,
			HighestScore:      b.HighestScore,
			BattingStrikeRate: strikeRate,
			BattingAverage:    average,
		}
		battingPlayers = append(battingPlayers, p)
	}

	bowlingPlayers := make([]BowlingPlayer, 0)
	for _, b := range bos.Bowlers {
		var average *float64
		if b.BowlingAverage == "NA" {
			average = nil
		} else {
			av, err := strconv.ParseFloat(b.BowlingAverage, 64)
			if err != nil {
				log.Fatal(err)
			}
			average = &av
		}

		ov, err := strconv.ParseFloat(b.Overs, 64)
		if err != nil {
			log.Fatal(err)
		}
		overs := &ov

		er, err := strconv.ParseFloat(b.EconomyRate, 64)
		if err != nil {
			log.Fatal(err)
		}
		economyRate := &er

		p := BowlingPlayer{
			Name:              toTitleCase(b.PlayerName),
			BowlingStyle:      b.BowlingStyle,
			Role:              b.PlayerRole,
			Team:              b.TeamName,
			Matches:           b.Matches,
			InningsBowled:     b.Innings,
			RunsGiven:         b.Runs,
			BallsBowled:       b.Balls,
			DotBallsBowled:    b.DotBalls,
			FoursGiven:        b.BdryFours,
			SixesGiven:        b.BdrySixes,
			Overs:             overs,
			Wickets:           b.Wickets,
			FourWickets:       b.FourWickets,
			FiveWickets:       b.FiveWickets,
			TenWickets:        b.TenWickets,
			Maidens:           b.Maidens,
			HighestWickets:    b.BestWickets,
			BowlingStrikeRate: &b.StrikeRate,
			BowlingAverage:    average,
			EconomyRate:       economyRate,
		}
		bowlingPlayers = append(bowlingPlayers, p)
	}

	// fmt.Println(battingPlayers[21:22])
	// fmt.Println(bowlingPlayers[21:22])

	teams := initTeams()
	// fmt.Println(teams[1])

	matchData, err := os.ReadFile("./raw/matches.json")
	if err != nil {
		log.Fatal(err)
	}
	var md MatchData
	err = json.Unmarshal(matchData, &md)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(len(md.Matches))

	// Open a connection to the SQLite database using GORM
	err = os.Remove("./prepared/player_stats.db")
	if err != nil {
		log.Println("No existing db file found.")
	}
	db, err := gorm.Open(sqlite.Open("./prepared/player_stats.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Automatically create the 'batting players' table based on the BattingPlayer struct
	err = db.AutoMigrate(&BattingPlayer{})
	if err != nil {
		log.Fatal(err)
	}
	// Insert data into the 'batting players' table
	for _, player := range battingPlayers {
		result := db.Create(&player)
		if result.Error != nil {
			log.Println("Error inserting:", result.Error)
		}
	}

	// Automatically create the 'bowling players' table based on the BowlingPlayer struct
	err = db.AutoMigrate(&BowlingPlayer{})
	if err != nil {
		log.Fatal(err)
	}
	// Insert data into the 'bowling players' table
	for _, player := range bowlingPlayers {
		result := db.Create(&player)
		if result.Error != nil {
			log.Println("Error inserting:", result.Error)
		}
	}

	// Automatically create the 'teams' table based on the Team struct
	err = db.AutoMigrate(&Team{})
	if err != nil {
		log.Fatal(err)
	}
	// Insert data into the 'teams' table
	for _, team := range teams {
		result := db.Create(&team)
		if result.Error != nil {
			log.Println("Error inserting:", result.Error)
		}
	}

	// Automatically create the 'matches' table based on the Match struct
	err = db.AutoMigrate(&Match{})
	if err != nil {
		log.Fatal(err)
	}
	// Insert data into the 'matches' table
	for _, match := range md.Matches {
		match.City = toTitleCase(match.City)
		split := strings.Split(match.GroundName, ",")
		match.GroundName = toTitleCase(split[0])
		result := db.Create(&match)
		if result.Error != nil {
			log.Println("Error inserting:", result.Error)
		}
	}

	fmt.Println("Data inserted successfully.")

	// Specify the name of the player you're looking for
	targetPlayerName := "Daniel Sams"
	// Querying the database for a player with a specific name
	var player BattingPlayer
	result := db.Where("name = ?", targetPlayerName).First(&player)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			fmt.Println("Player not found.")
		} else {
			log.Println("Error querying:", result.Error)
		}
		return
	}
	fmt.Println(player)

	// Specify the name of the team you're looking for
	targetTeamName := "Mumbai Indians New York"
	// Querying the database for a team with a specific name
	var team Team
	result = db.Where("full_name = ?", targetTeamName).First(&team)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			fmt.Println("Team not found.")
		} else {
			log.Println("Error querying:", result.Error)
		}
		return
	}
	fmt.Println(team)

	// Specify the name of the  winning team in the match you're looking for
	targetWinningTeamName := "MI NY"
	// Querying the database for a team with a specific name
	var match Match
	result = db.Where("win_team_name = ?", targetWinningTeamName).First(&match)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			fmt.Println("Match not found.")
		} else {
			log.Println("Error querying:", result.Error)
		}
		return
	}
	fmt.Println(match)

	// var res BattingPlayer
	// var res BowlingPlayer
	var res Match
	// var res Team

	// rawQuery := "SELECT win_team_name, result FROM matches WHERE team_a_name = 'SEA' and team_b_name = 'MI NY' AND city = 'Dallas'"
	// rawQuery := "SELECT * FROM batting_players WHERE name = 'Nicholas Pooran'"
	// rawQuery := "SELECT * FROM players WHERE id < 2"
	// rawQuery := "SELECT name, batting_strike_rate FROM batting_players ORDER BY batting_strike_rate ASC LIMIT 1"
	// rawQuery := "SELECT name, batting_strike_rate FROM players ORDER BY runs_scored DESC LIMIT 1"
	// rawQuery := "SELECT name, bowling_average FROM bowling_players ORDER BY bowling_average ASC LIMIT 1"
	rawQuery := "SELECT team_a_name, team_b_name FROM matches WHERE ground_name = 'Church Street Park'"
	// rawQuery := "SELECT team_a_name, team_b_name FROM matches WHERE id = 1"
	// rawQuery := "SELECT team_a_name, team_b_name FROM matches WHERE city = 'Morrisville'"
	// rawQuery := "SELECT image FROM teams WHERE name = (SELECT win_team_name FROM matches WHERE id = (SELECT MIN(id) FROM matches WHERE city = 'Morrisville'))"

	result = db.Raw(rawQuery).Scan(&res)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			fmt.Println("Match not found.")
		} else {
			log.Println("Error querying:", result.Error)
		}
		return
	}
	fmt.Println(res)
}
