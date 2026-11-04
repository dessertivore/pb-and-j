package main

import (
	"context"
	"fmt"
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
	Side     Side       `json:"side"`
	Type     ActionType `json:"type"`
	Champion string     `json:"champion,omitempty" doc:"Empty for a skipped ban"`
	Role     string     `json:"role" query:"role" enum:"top,jungle,mid,bottom,support" doc:"Lane role. User will have to guess this for opponent."`
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
	ID       string `json:"id"`
	User     User   `json:"user"`
	UserSide Side
	Actions  []Action `json:"actions" doc:"In action order"`
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
func getWinRates(action Action, user User) map[string]float64 {
	winRates := make(map[string]float64)
	_, err := getChamps(context.Background())
	if err != nil {
		return map[string]float64{}
	}
	if opponentSlug, ok := Champions[action.Champion]; ok {
		for _, champion := range user.PreferredChampions {
			webpageURL := fmt.Sprintf(WinRateURL, champion.Slug, action.Role, opponentSlug)
			winRate := scrapeWinRates(webpageURL)
			fmt.Printf("Win rate for %s against %s is %f", champion.Name, opponentSlug, winRate)
			winRates[champion.Name] = winRate
		}

		return map[string]float64{}
	}
	return winRates
}
