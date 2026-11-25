package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
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
	Side     Side               `json:"side" default:"red"`
	Type     ActionType         `json:"type" default:"pick"`
	Champion string             `json:"champion,omitempty" default:"yasuo" doc:"Empty for a skipped ban"`
	Role     string             `json:"role" default:"top" query:"role" enum:"top,jungle,mid,bottom,support" doc:"Lane role. User will have to guess this for opponent."`
	WinRates map[string]float64 `json:"winRates,omitempty"`
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

func scrapeWinRates(winRateURL string) (float64, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	req, err := http.NewRequest("GET", winRateURL, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	// Parse the HTML response.
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return 0, err
	}

	// Find the Win rate element and extract its percentage.
	var winRate float64
	var found bool

	doc.Find("em").EachWithBreak(func(i int, s *goquery.Selection) bool {

		if strings.TrimSpace(s.Text()) == "Win rate" {

			value := strings.TrimSuffix(
				strings.TrimSpace(s.Parent().Find("b").First().Text()),
				"%",
			)

			winRate, err = strconv.ParseFloat(value, 64)
			found = err == nil

			return false
		}

		return true
	})

	if !found {
		if err != nil {
			return 0, err
		}
		return 0, fmt.Errorf("win rate not found")
	}

	return winRate, nil
}

func postActionFunc(action *PostActionInput, game *PickBan, availableChamps *[]string, db *gorm.DB) (*PostActionOutput, error) {
	var winRates map[string]float64
	if action.Body.Action.Side != game.UserSide {
		if action.Body.Action.Type == "pick" {
			relevantUser := game.Team.selectTeamMate(action.Body.Action.Role)

			if relevantUser.DisplayName != "" { // not empty User obj
				fmt.Println("Comparing win rates against champs for", relevantUser.DisplayName)
				winRates = getWinRates(action.Body.Action, relevantUser, db)
				action.Body.Action.WinRates = winRates
			} else {
				fmt.Println("No relevant users in this role to compare champs for")
			}
			fmt.Println("Successfully added action " + action.Body.Action.Type)
		}
	}
	game.Actions = append(game.Actions, action.Body.Action)
	for i, champ := range *availableChamps {
		if strings.ToLower(champ) == action.Body.Action.Champion {
			*availableChamps = append((*availableChamps)[:i], (*availableChamps)[i+1:]...)
			break
		}
	} // can't reuse this champ in this game, remove from champ list
	return &PostActionOutput{Body: winRates}, nil
}
