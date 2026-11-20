package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open("public.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&DBUser{}, &UserToChampions{})
	db.AutoMigrate(&Champion{})
	apiEntryPoint(db)
}
