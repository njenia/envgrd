package config

import "os"

// This file is in an ignored folder, so its env vars should be ignored
func LoadConfig() {
	_ = os.Getenv("DATABASE_URL")
	_ = os.Getenv("SECRET_KEY")
}

