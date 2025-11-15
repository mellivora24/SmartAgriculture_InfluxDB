package database

import (
	"backend/internal/shared"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PostgresDB struct {
	DB     *gorm.DB
	Config *shared.DatabaseConfig
}

func NewPostgresDB(config *shared.DatabaseConfig) (*PostgresDB, error) {
	dsn := config.GetDSN()

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	log.Printf("PostgreSQL connected successfully at %s:%d", config.Host, config.Port)

	return &PostgresDB{
		DB:     db,
		Config: config,
	}, nil
}

func (p *PostgresDB) Close() error {
	sqlDB, err := p.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close PostgreSQL connection: %w", err)
	}

	log.Println("PostgreSQL connection closed")
	return nil
}

func (p *PostgresDB) GetDB() *gorm.DB {
	return p.DB
}

func (p *PostgresDB) HealthCheck() error {
	sqlDB, err := p.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

func (p *PostgresDB) GetStats() map[string]interface{} {
	sqlDB, err := p.DB.DB()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}
