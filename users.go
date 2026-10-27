package main

import "fmt"

type Role int

const (
	RoleTop Role = iota // 0
	RoleJungle
	RoleMid
	RoleBottom
	RoleSupport
)

var roleNames = map[Role]string{
	RoleTop:     "top",
	RoleJungle:  "jungle",
	RoleMid:     "mid",
	RoleBottom:  "bottom",
	RoleSupport: "support",
}

func (r Role) String() string {
	if r < 0 || int(r) >= len(roleNames) {
		return fmt.Sprintf("Role(%d)", int(r))
	}
	return roleNames[r]
}

type User struct {
	DisplayName        string   `json:"displayName"`
	Username           string   `json:"username"`
	PreferredChampions []string `json:"champions"`
	AssignedRole       Role     `json:"assignedRole"`
}
