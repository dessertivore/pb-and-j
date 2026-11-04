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
	"github.com/google/uuid"
)

type ChampsOutput struct {
	Body struct {
		Champs []string `json:"champs" example:"[\"Dr Mundo\"]" doc:"List of available champs"`
	}
}

type GetUsersInput struct {
	Username string `query:"username" example:"dessertivore"`
}

type GetGamesInput struct {
	ID string `query:"id" example:"test-id"`
}
type CreateUserInput struct {
	Body UserInput
}

type GetUsersOutput struct {
	Body []User
}

type CreateGameInput struct {
	ID   string `query:"id"`
	User string `query:"username"`
	Side Side   `query:"side"`
}

type CreateGameOutput struct {
	Body struct {
		ID string `json:"id"`
	}
}

type GetGameOutput struct {
	Body []PickBan
}
type PostActionInput struct {
	Body struct {
		Action    Action `json:"action"`
		PickBanID string `json:"pick_ban_id"`
	}
}

type PostActionOutput struct {
	Body map[string]float64 `json:"win_rates"`
}

func apiEntryPoint() {
	// Create a new router & API
	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("PB&JPI", "1.0.0"))

	users := make(map[string]User, 5)     // TODO add these to a db
	games := make(map[string]PickBan, 10) // TODO add these to a db

	// Register GET /champions handler
	huma.Get(api, "/champions", func(ctx context.Context, _ *struct {
	}) (*ChampsOutput, error) {
		champs, err := getChamps(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		resp := &ChampsOutput{}
		resp.Body.Champs = champs
		return resp, nil
	})

	// Register GET /users handler
	huma.Get(api, "/user", func(ctx context.Context, in *GetUsersInput) (*GetUsersOutput, error) {
		userList := make([]User, 0, len(users))
		if in.Username != "" {
			userList = append(userList, users[in.Username])
		} else {
			for _, v := range users {
				userList = append(userList, v)
			}
		}
		return &GetUsersOutput{Body: userList}, nil
	})

	// Register POST /user handler
	huma.Post(api, "/user", func(ctx context.Context, user *CreateUserInput) (*CreateUserInput, error) {
		validatedChamps := make([]Champion, 0, 5)
		for _, champ := range user.Body.PreferredChampions {
			champ = strings.ToLower(champ)
			if champ, ok := Champions[champ]; ok {
				validatedChamps = append(validatedChamps, champ)
			} else {
				return nil, huma.Error400BadRequest(fmt.Sprintf("Champion %s does not exist", champ))
			}
		}
		users[user.Body.DisplayName] = User{
			DisplayName:        user.Body.DisplayName,
			Username:           user.Body.Username,
			PreferredChampions: validatedChamps,
			AssignedRole:       user.Body.AssignedRole,
		}
		fmt.Println("Successfully added " + user.Body.DisplayName)
		return &CreateUserInput{Body: user.Body}, nil
	})

	// Register POST /action handler
	huma.Post(api, "/action", func(ctx context.Context, action *PostActionInput) (*PostActionOutput, error) {
		game, ok := games[action.Body.PickBanID]
		if !ok {
			return nil, huma.Error404NotFound("game not found")
		}
		game.Actions = append(game.Actions, action.Body.Action)
		if action.Body.Action.Side != game.UserSide {
			winRates := getWinRates(action.Body.Action, game.User)
			fmt.Println("Successfully added action to game ID " + action.Body.PickBanID)
			return &PostActionOutput{Body: winRates}, nil
		}

		return nil, nil

	})

	// Register POST /game handler
	huma.Post(api, "/game", func(ctx context.Context, in *CreateGameInput) (*CreateGameOutput, error) {
		id := in.ID
		if id != "" {
			if _, exists := games[id]; exists {
				return nil, huma.Error409Conflict("game ID already exists")
			}
		} else {
			id = uuid.New().String()
		}
		if user, ok := users[in.User]; ok {

			games[id] = PickBan{ID: id, User: user, UserSide: in.Side}
			fmt.Println("Successfully added game ID " + id)
			out := &CreateGameOutput{}
			out.Body.ID = id
			return out, nil
		}
		return nil, huma.Error404NotFound("game or user not found")
	})

	// Register GET /game handler
	huma.Get(api, "/game", func(ctx context.Context, in *GetGamesInput) (*GetGameOutput, error) {
		gamesList := make([]PickBan, 0, len(games))
		if in.ID != "" {
			fmt.Println("Getting game ID " + in.ID)
			if val, ok := games[in.ID]; ok {
				gamesList = append(gamesList, val)
			} else {
				return nil, huma.Error400BadRequest("game ID does not exists")

			}
		} else {
			fmt.Println("Getting all games...")
			for _, v := range games {
				gamesList = append(gamesList, v)
			}
		}
		return &GetGameOutput{Body: gamesList}, nil
	})

	// Start the server!
	http.ListenAndServe("127.0.0.1:8888", router)
}
