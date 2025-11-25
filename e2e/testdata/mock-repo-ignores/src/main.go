package main

import (
	"fmt"
	"os"
)

func main() {
	// This should be reported as missing (not in ignores)
	apiKey := os.Getenv("API_KEY")
	
	// These should be ignored (in ignores.missing)
	customKey := os.Getenv("CUSTOM_API_KEY")
	externalToken := os.Getenv("EXTERNAL_SERVICE_TOKEN")
	
	fmt.Println(apiKey, customKey, externalToken)
}

