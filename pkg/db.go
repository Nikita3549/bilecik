// Package db provides database connectivity
package db

import (
	"bilecik/internal/configs"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func NewDB(conf *configs.Config) *DB {
	db, err := gorm.Open(postgres.Open(conf.DBConfig.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	return &DB{
		DB: db,
	}
}

func (d *DB) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
