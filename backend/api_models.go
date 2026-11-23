package main

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
