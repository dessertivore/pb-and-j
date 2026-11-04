package main

type User struct {
	DisplayName        string     `json:"displayName"`
	Username           string     `json:"username"`
	PreferredChampions []Champion `json:"champions"`
	AssignedRole       string     `json:"assignedRole" query:"role" enum:"top,jungle,mid,bottom,support" doc:"Lane role"`
}
type UserInput struct {
	DisplayName        string   `json:"displayName"`
	Username           string   `json:"username"`
	PreferredChampions []string `json:"preferredChampions"`
	AssignedRole       string   `json:"assignedRole" query:"role" enum:"top,jungle,mid,bottom,support" doc:"Lane role"`
}
