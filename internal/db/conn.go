package db

import (
	"fmt"
	"github.com/RacoonMediaServer/rms-bot-server/internal/model"
	"github.com/RacoonMediaServer/rms-packages/pkg/configuration"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	conn *gorm.DB
}

func Connect(config configuration.Database) (*Database, error) {
	db, err := gorm.Open(postgres.Open(config.GetConnectionString()))
	if err != nil {
		return nil, err
	}
	if err = db.AutoMigrate(&model.Linkage{}); err != nil {
		return nil, fmt.Errorf("update database failed: %s", err)
	}
	return &Database{conn: db}, nil
}
