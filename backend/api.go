package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type ChampsOutput struct {
	Body struct {
		Champs []string `json:"champs" example:"[\"Dr Mundo\"]" doc:"List of available champs"`
	}
}
type GetChampionsInput struct {
	All bool `query:"all" nullable:"true"`
}
type GetUsersInput struct {
	Username string `query:"username" example:"dessertivore"`
}

type CreateUserInput struct {
	Body User
}

type GetUsersOutput struct {
	Body []User
}

type SetSideInput struct {
	Side Side `query:"side"`
}

type GetGameOutput struct {
	Body PickBan
}
type PostActionInput struct {
	Body struct {
		Action Action `json:"action"`
	}
}

type PostActionOutput struct {
	Body map[string]float64 `json:"win_rates"`
}

func apiEntryPoint(db *gorm.DB) {
	// Create a new router & API
	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("PB&JPI", "1.0.0"))
	champs, err := getChamps(context.Background(), db, false)
	if err != nil {
		log.Fatal(err.Error())
	}
	availableChamps := champs
	game := new(PickBan)
	game.UserSide = "blue" // default to blue side
	game.Team = CreateTeamFromUsers(GetUsersFromDB(*db, []string{}, ""))
	fmt.Println("Champions scraped or obtained from db, default user set up for testing")
	// Register GET /champions handler
	huma.Get(api, "/champions", func(ctx context.Context, in *GetChampionsInput) (*ChampsOutput, error) {
		resp := &ChampsOutput{}
		if in.All {
			resp.Body.Champs = champs
		} else {
			resp.Body.Champs = availableChamps
		}
		return resp, nil
	})

	// Register GET /users handler
	huma.Get(api, "/user", func(ctx context.Context, in *GetUsersInput) (*GetUsersOutput, error) {
		var userSearch []string
		if in.Username == "" {
			userSearch = []string{}
		} else {
			userSearch = []string{in.Username}
		}
		userList := GetUsersFromDB(*db, userSearch, "")
		return &GetUsersOutput{Body: userList}, nil
	})

	// Register POST /user handler
	huma.Post(api, "/user", func(ctx context.Context, user *CreateUserInput) (*CreateUserInput, error) {
		validatedChamps := ValidateChampsAgainstCache(&user.Body.PreferredChampions)

		result := db.Save(&DBUser{DisplayName: user.Body.DisplayName,
			Username:     user.Body.Username,
			AssignedRole: user.Body.AssignedRole})

		if result.Error != nil {
			fmt.Println(result.Error)
			return nil, result.Error
		}

		var champOrms []UserToChampion
		for _, validatedChamp := range validatedChamps {
			champOrms = append(champOrms, UserToChampion{Username: user.Body.Username, ChampionName: strings.ToLower(validatedChamp.Name)})
		}
		result = db.Save(&champOrms)
		if result.Error != nil {
			fmt.Println(result.Error)
			return nil, result.Error
		}

		fmt.Println("Successfully added or updated " + user.Body.DisplayName)
		return &CreateUserInput{Body: user.Body}, nil
	})

	// Register POST /action handler
	huma.Post(api, "/action", func(ctx context.Context, action *PostActionInput) (*PostActionOutput, error) {
		var winRates map[string]float64
		if action.Body.Action.Side != game.UserSide {
			if action.Body.Action.Type == "pick" {

				relevantUser := GetUsersFromDB(*db, []string{}, action.Body.Action.Role)
				fmt.Println("Comparing win rates against champs for", relevantUser[0].DisplayName)
				winRates = getWinRates(action.Body.Action, relevantUser[0], db)
				action.Body.Action.WinRates = winRates
				fmt.Println("Successfully added action " + action.Body.Action.Type)
			}
		}
		game.Actions = append(game.Actions, action.Body.Action)
		for i, champ := range availableChamps {
			if strings.ToLower(champ) == action.Body.Action.Champion {
				availableChamps = append(availableChamps[:i], availableChamps[i+1:]...)
				break
			}
		} // can't reuse this champ in this game, remove from champ list
		return &PostActionOutput{Body: winRates}, nil

	})

	// Register POST /side handler
	huma.Post(api, "/side", func(ctx context.Context, in *SetSideInput) (*struct{}, error) {
		game.UserSide = in.Side
		return nil, nil
	})

	// Register GET /game handler
	huma.Get(api, "/game", func(ctx context.Context, in *struct{}) (*GetGameOutput, error) {
		return &GetGameOutput{Body: *game}, nil
	})

	// Register POST /reset handler - resets game but keeps user
	huma.Post(api, "/reset", func(ctx context.Context, in *struct{}) (*struct{}, error) {
		game = new(PickBan)
		game.Team = CreateTeamFromUsers(GetUsersFromDB(*db, []string{}, ""))
		champs, err = getChamps(ctx, db, true)
		return nil, nil
	})

	// Start the server!
	fmt.Println("Successfully started server")

	http.ListenAndServe(":8888", router)
}
