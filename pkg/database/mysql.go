package database

import (
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Options struct {
	DSN             string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int
}

func NewMySQL(opt Options) (*gorm.DB, error) {

	db, err := gorm.Open(mysql.Open(opt.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(opt.MaxIdleConns)
	sqlDB.SetMaxOpenConns(opt.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(opt.ConnMaxLifetime) * time.Second)

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
