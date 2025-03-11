package config

import "os"

var JWTSecret = []byte(os.Getenv("JWT_SECRET")) // Set this in your environment
