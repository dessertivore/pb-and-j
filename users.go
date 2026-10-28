package main

type User struct {
	DisplayName        string   `json:"displayName"`
	Username           string   `json:"username"`
	PreferredChampions []string `json:"champions"`
	AssignedRole       string   `json:"assignedRole" query:"role" enum:"top,jungle,mid,bottom,support" doc:"Lane role"`
}
