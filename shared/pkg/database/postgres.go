package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Database struct {
	Pool *pgxpool.Pool
}

func NewDatabase(logger *zap.Logger) (*Database, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://blytz:password@localhost:5432/blytz_mvp?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		logger.Error("Failed to create database pool", zap.Error(err))
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		logger.Error("Failed to ping database", zap.Error(err))
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connected successfully")
	return &Database{Pool: pool}, nil
}

func (db *Database) Close() {
	db.Pool.Close()
	log.Println("Database connection closed")
}

func (db *Database) Health() error {
	return db.Pool.Ping(context.Background())
}
