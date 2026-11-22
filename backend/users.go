package main

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type User struct {
	DisplayName        string   `json:"displayName" default:"default_user"`
	Username           string   `json:"username" default:"dessertivore"`
	PreferredChampions []string `json:"preferredChampions" default:"[\"dr. mundo\", \"jinx\", \"lux\"]"`
	AssignedRole       string   `json:"assignedRole" default:"mid" query:"role" enum:"top,jungle,mid,bottom,support" doc:"Lane role"`
}

type Team struct {
	Top     User `json:"top"`
	Jungle  User `json:"jungle"`
	Mid     User `json:"mid"`
	Bottom  User `json:"bottom"`
	Support User `json:"support"`
}

type DBUser struct {
	DisplayName  string
	Username     string `gorm:"primaryKey"`
	AssignedRole string
}

func (DBUser) TableName() string {
	return "users"
}

type UserToChampion struct {
	Username     string `gorm:"not null;primaryKey;foreignKey:Username;references:Username;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ChampionName string `gorm:"not null;primaryKey"`
}

func DefaultUser() *User {
	champs := []string{"dr. mundo", "jinx", "lux"}
	return &User{DisplayName: "default_user", Username: "dessertivore", PreferredChampions: champs, AssignedRole: "top"}
}

func ValidateChampsAgainstCache(champs *[]string) []Champion {
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

func DefaultTeam() Team {
	return Team{
		Top:     *DefaultUser(),
		Jungle:  *DefaultUser(),
		Mid:     *DefaultUser(),
		Bottom:  *DefaultUser(),
		Support: *DefaultUser(),
	}
}
func GetUsersFromDB(db gorm.DB, usernames []string, role string) []User {
	// Get users from database, filtering by usernames if provided
	var users []DBUser
	var usersWithChamps []User
	var result *gorm.DB
	if len(usernames) > 0 {
		result = db.Find(&users, usernames)
	} else if role != "" {
		result = db.Where("assigned_role = ?", role).Find(&users)
	} else {
		result = db.Find(&users)
	}
	if result.Error != nil {
		fmt.Println(result.Error)
	}
	for _, user := range users {
		var userChampions []string

		result := db.
			Model(&UserToChampion{}).
			Where("username = ?", user.Username).
			Pluck("champion_name", &userChampions)
		if result.Error != nil {
			fmt.Println(result.Error)
		}

		userWithChamps := User{
			Username:           user.Username,
			DisplayName:        user.DisplayName,
			AssignedRole:       user.AssignedRole,
			PreferredChampions: userChampions,
		}
		usersWithChamps = append(usersWithChamps, userWithChamps)

	}

	return usersWithChamps

}

func CreateTeamFromUsers(users []User) Team {
	team := DefaultTeam()
	for _, user := range users {
		switch user.AssignedRole {
		case "top":
			team.Top = user
		case "mid":
			team.Mid = user
		case "bottom":
			team.Bottom = user
		case "jungle":
			team.Jungle = user
		case "support":
			team.Support = user
		}

	}
	return team
}
