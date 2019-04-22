package main

import (
	"log"
	"staply/storage/app"
)

func main() {
	router := app.NewRouter()
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("error: %v\n", err)
	}
}
