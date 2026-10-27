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

type CreateUserInput struct {
	Body User
}

type GetUserOutput struct {
	Body []User
}

func apiEntryPoint() {
	// Create a new router & API
	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("PB&JPI", "1.0.0"))

	users := make(map[string]User) // TODO add these to a db

	// Register GET /champions handler
	huma.Get(api, "/champions", func(ctx context.Context, _ *struct {
	}) (*ChampsOutput, error) {
		resp := &ChampsOutput{}
		champs, err := getChampsString(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Champs = champs
		return resp, nil
	})

	// Register GET /users handler
	huma.Get(api, "/users", func(ctx context.Context, _ *struct {
	}) (*GetUserOutput, error) {
		var userList []User
		for _, v := range users {
			userList = append(userList, v)
		}
		return &GetUserOutput{Body: userList}, nil
	})

	// Register POST /user handler
	huma.Post(api, "/user", func(ctx context.Context, user *CreateUserInput) (*CreateUserInput, error) {
		users[user.Body.DisplayName] = user.Body
		fmt.Println("Successfully added " + user.Body.DisplayName)
		return &CreateUserInput{Body: user.Body}, nil
	})


	// Start the server!
	http.ListenAndServe("127.0.0.1:8888", router)
}
