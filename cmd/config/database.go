package config

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func ConnectDatabase() *gorm.DB {
	db, err := gorm.Open(mysql.Open("grpc:{superSecretPassword!123}@tcp(127.0.0.1:3306)/grpcbookauthor"))

	if err != nil {
		log.Fatalf("Database connection failed %v", err.Error())
	}

	return db
}
