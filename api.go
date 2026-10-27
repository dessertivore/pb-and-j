package main

import (
	"context"
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

func apiEntryPoint() {
	// Create a new router & API
	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("My API", "1.0.0"))

	// Register GET /greeting/{name} handler.
	huma.Get(api, "/champions", func(ctx context.Context, _ *struct {
	}) (*ChampsOutput, error) {
		resp := &ChampsOutput{}
		champs, err := getChamps(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		names := make([]string, len(champs))
		for i, c := range champs {
			names[i] = c.Name
		}
		resp.Body.Champs = names
		return resp, nil
	})

	// Start the server!
	http.ListenAndServe("127.0.0.1:8888", router)
}
