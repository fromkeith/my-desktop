package globals

import (
	_ "github.com/joho/godotenv/autoload"
)

func init() {
	Db() // connect to postgress
}
