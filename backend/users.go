package main

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type User struct {
	DisplayName        string     `json:"displayName" default:"default_user"`
	Username           string     `json:"username" default:"dessertivore"`
	PreferredChampions []string   `json:"preferredChampions" default:"[\"dr. mundo\", \"jinx\", \"lux\"]"`
	AssignedRole       string     `json:"assignedRole" default:"mid" query:"role" enum:"top,jungle,mid,bottom,support" doc:"Lane role"`
	ChampsValidated    bool       `json:"champsValidated" default:"false" required:"false"`
	ValidatedChamps    []Champion `json:"validatedChamps" default:"[]" required:"false"`
}

type DBUser struct {
	gorm.Model
	DisplayName  string
	Username     string
	AssignedRole string
}

type UserToChampions struct {
	gorm.Model
	Username string
	Champion string
}

func DefaultUser() *User {
	champs := []string{"dr. mundo", "jinx", "lux"}
	return &User{DisplayName: "default_user", Username: "dessertivore", PreferredChampions: champs, AssignedRole: "top", ChampsValidated: true, ValidatedChamps: ValidateChamps(&champs)}
}

func ValidateChamps(champs *[]string) []Champion {
	validatedChamps := make([]Champion, 0, 5)

	for _, champ := range *champs {
		champ = strings.ToLower(champ)
		if validChamp, ok := Champions[champ]; ok {
			validatedChamps = append(validatedChamps, validChamp)
		} else {
			fmt.Println("Champion", champ, "does not exist")
		}
	}
	return validatedChamps
}
