package services

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Services struct {
	DB    *gorm.DB
	Redis *redis.Client
}
