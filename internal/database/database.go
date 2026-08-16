package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func init() {
	db_, err := gorm.Open(sqlite.Open("metrics.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db = db_
}

func Db() *gorm.DB {
	return db
}
