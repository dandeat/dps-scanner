package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	JWTKey    = GetEnv("JWT_KEY", "MKPMobile123@")
	SECRETKey = GetEnv("SECRET_KEY", "qFpdW1A9udvi8PPh")

	// Swicth Debug Mode
	DebugMode = GetEnv("DEBUG_MODE", "true")
)

func GetEnv(key string, value ...string) string {
	if err := godotenv.Load(".env"); err != nil {
		panic("Error Load file .env not found")
	}

	if os.Getenv(key) != "" {
		log.Println("GetEnv: ", key, " = ", os.Getenv(key))
		return os.Getenv(key)
	} else {
		if len(value) > 0 {
			log.Println("GetEnv: Default ", key, " = ", value[0])
			return value[0]
		}

		log.Println("GetEnv: Not Found ", key)
		return ""
	}
}
