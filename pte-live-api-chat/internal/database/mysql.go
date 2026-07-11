package database

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
	"pte_live_api_chat/pkg/setting"
)

func NewMySQL() (*gorm.DB, error) {
	writeDSN := setting.MySQL.WriteDSN
	if writeDSN == "" {
		writeDSN = setting.MySQL.DSN
	}
	if writeDSN == "" {
		return nil, fmt.Errorf("mysql write dsn is required")
	}

	db, err := gorm.Open(mysql.Open(writeDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}

	if setting.MySQL.HasReadReplica() {
		replicas := make([]gorm.Dialector, 0, len(setting.MySQL.ReadDSNs))
		for _, dsn := range setting.MySQL.ReadDSNs {
			if dsn != "" && dsn != writeDSN {
				replicas = append(replicas, mysql.Open(dsn))
			}
		}
		if len(replicas) > 0 {
			if err := db.Use(dbresolver.Register(dbresolver.Config{
				Sources:  []gorm.Dialector{mysql.Open(writeDSN)},
				Replicas: replicas,
				Policy:   dbresolver.RandomPolicy{},
			})); err != nil {
				return nil, fmt.Errorf("mysql dbresolver: %w", err)
			}
		}
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(setting.MySQL.MaxOpenConns)
	sqlDB.SetMaxIdleConns(setting.MySQL.MaxIdleConns)
	if setting.MySQL.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(setting.MySQL.ConnMaxLifetime) * time.Second)
	}
	return db, nil
}
