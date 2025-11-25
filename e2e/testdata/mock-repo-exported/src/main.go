package main

import (
	"fmt"
	"os"
)

func main() {
	// This is in .env file
	apiKey := os.Getenv("API_KEY")
	
	// This is only in exported environment (not in .env)
	ciToken := os.Getenv("CI_TOKEN")
	
	// This is missing (not in .env or exported)
	missingVar := os.Getenv("MISSING_VAR")
	
	fmt.Println(apiKey, ciToken, missingVar)
}

