package main

import (
	"fmt"
	"os"
)

func main() {
	apiKey := os.Getenv("API_KEY")
	dbUrl := os.Getenv("DATABASE_URL")
	fmt.Println(apiKey, dbUrl)
}

