package repositories

import (
	"context"
	"database/sql"
)

type Repository struct {
	DB      *sql.DB
	Context context.Context
}

func NewRepository(conn *sql.DB, ctx context.Context) Repository {
	return Repository{
		DB:      conn,
		Context: ctx,
	}
}
