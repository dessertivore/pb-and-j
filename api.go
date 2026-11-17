package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

type ChampsOutput struct {
	Body struct {
		Champs []string `json:"champs" example:"[\"Dr Mundo\"]" doc:"List of available champs"`
	}
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

func apiEntryPoint() {
	// Create a new router & API
	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("PB&JPI", "1.0.0"))

	champs, err := getChamps(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	users := make(map[string]*User, 5) // TODO add these to a db if having multiple
	users["dessertivore"] = DefaultUser()
	game := new(PickBan)
	game.User = *users["dessertivore"]
	// Register GET /champions handler
	huma.Get(api, "/champions", func(ctx context.Context, _ *struct {
	}) (*ChampsOutput, error) {
		resp := &ChampsOutput{}
		resp.Body.Champs = champs
		return resp, nil
	})

	// Register GET /users handler
	huma.Get(api, "/user", func(ctx context.Context, in *GetUsersInput) (*GetUsersOutput, error) {
		userList := make([]User, 0, len(users))
		if in.Username != "" {
			userList = append(userList, *users[in.Username])
		} else {
			for _, v := range users {
				userList = append(userList, *v)
			}
		}
		return &GetUsersOutput{Body: userList}, nil
	})

	// Register POST /user handler
	huma.Post(api, "/user", func(ctx context.Context, user *CreateUserInput) (*CreateUserInput, error) {
		validatedChamps := ValidateChamps(&user.Body.PreferredChampions)

		newUser := User{
			DisplayName:        user.Body.DisplayName,
			Username:           user.Body.Username,
			ChampsValidated:    true,
			ValidatedChamps:    validatedChamps,
			PreferredChampions: user.Body.PreferredChampions,
			AssignedRole:       user.Body.AssignedRole,
		}
		users[user.Body.DisplayName] = &newUser
		game.User = newUser
		fmt.Println("Successfully added " + user.Body.DisplayName)
		return &CreateUserInput{Body: user.Body}, nil
	})

	// Register POST /action handler
	huma.Post(api, "/action", func(ctx context.Context, action *PostActionInput) (*PostActionOutput, error) {
		game.Actions = append(game.Actions, action.Body.Action)
		if action.Body.Action.Side != game.UserSide {
			if action.Body.Action.Type == "pick" {
				winRates := getWinRates(action.Body.Action, game.User)
				fmt.Println("Successfully added action " + action.Body.Action.Type)
				return &PostActionOutput{Body: winRates}, nil
			} else {
				return &PostActionOutput{Body: make(map[string]float64)}, nil
			}
		}

		return nil, nil

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

	// Start the server!
	http.ListenAndServe("127.0.0.1:8888", router)
}
