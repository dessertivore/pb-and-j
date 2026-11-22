package main

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type Side string

const (
	SideBlue Side = "blue"
	SideRed  Side = "red"
)

type ActionType string

const (
	ActionBan  ActionType = "ban"
	ActionPick ActionType = "pick"
)

const WinRateURL = "https://op.gg/lol/champions/%s/build/%s?target_champion=%s"

type Action struct {
	Side     Side       `json:"side" default:"red"`
	Type     ActionType `json:"type" default:"pick"`
	Champion string     `json:"champion,omitempty" default:"yasuo" doc:"Empty for a skipped ban"`
	Role     string     `json:"role" default:"top" query:"role" enum:"top,jungle,mid,bottom,support" doc:"Lane role. User will have to guess this for opponent."`
}

type Phase struct {
	Type         ActionType
	Sides        string // one letter per action, e.g. "RBBRRB"
	Simultaneous bool   // if true, only the per-side counts matter, not the order
	AllowDupes   bool   // if true, both sides may pick the same champion within this phase
}

var phases = []Phase{
	{Type: ActionBan, Sides: "RRRBBB", Simultaneous: true},
	{Type: ActionPick, Sides: "RBBRRB"},
	{Type: ActionBan, Sides: "BBRR", Simultaneous: true},
	{Type: ActionPick, Sides: "BRRB"},
}

type PickBan struct {
	Team     Team     `json:"team" doc:"team"`
	UserSide Side     `json:"side" default:"red"`
	Actions  []Action `json:"actions" default:"[]" doc:"In action order"`
}

// By returns the champions a side picked or banned, in action order.
func (d PickBan) By(side Side, t ActionType) []string {
	var out []string
	for _, a := range d.Actions {
		if a.Side == side && a.Type == t {
			out = append(out, a.Champion)
		}
	}
	return out
}

// For opponent's selection: compare their win rate percentage against user's top 5 champs (ordered)
func getWinRates(action Action, user User, db *gorm.DB) map[string]float64 {
	winRates := make(map[string]float64)
	if len(Champions) == 0 {
		fmt.Println("Getting champions from DB/internet as not cached")
		_, err := getChamps(context.Background(), db, false)
		if err != nil {
			return map[string]float64{}
		}
	}
	if opponent, ok := Champions[action.Champion]; ok {
		for _, champion := range user.PreferredChampions {
			fullChampData := Champions[champion]
			webpageURL := fmt.Sprintf(WinRateURL, fullChampData.Slug, action.Role, opponent.Slug)
			winRate, err := scrapeWinRates(webpageURL)
			if err != nil {
				fmt.Println("Win rates not obtained. ", err)
				winRates[fullChampData.Name] = 0
			}
			fmt.Printf("Win rate for %s against %s is %f \n", fullChampData.Name, opponent.Name, winRate)
			winRates[fullChampData.Name] = winRate
		}

		return winRates
	}
	return map[string]float64{}
}
