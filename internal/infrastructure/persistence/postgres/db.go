package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.elastic.co/apm/module/apmsql/v2"
	_ "go.elastic.co/apm/module/apmsql/v2/pq"
)

func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := apmsql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}
	return db, nil
}
