package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Queries struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Queries {
	return &Queries{db: db}
}

func NewDb(databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database URL: %w", err)
	}

	// Set connection pool options
	config.MaxConns = 10
	config.MinConns = 2

	dbPool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return dbPool, nil
}


// TODO: Install GORM and implement CRUD operations
// Link: https://github.com/go-gorm/gorm
// func CreateSensorReading(queries *Queries, sensor_reading models.SensorReading) error {
// 	queries.db.E
// }