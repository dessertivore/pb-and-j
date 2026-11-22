package main

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  "host=db user=postgres password=this_isnt_secure port=5432 dbname=postgres sslmode=disable TimeZone=Europe/London",
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&DBUser{}, &UserToChampion{}, &Champion{})

	apiEntryPoint(db)
}
