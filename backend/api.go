package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

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
		return UpsertUser(user, db)
	})

	// Register POST /action handler
	huma.Post(api, "/action", func(ctx context.Context, action *PostActionInput) (*PostActionOutput, error) {
		return postActionfunc(action, game, &availableChamps, db)

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
