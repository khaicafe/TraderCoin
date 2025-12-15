package services

import (
	"database/sql"

	"github.com/redis/go-redis/v9"
)

type Services struct {
	DB    *sql.DB
	Redis *redis.Client
}
