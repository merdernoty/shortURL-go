package postgres

import (
	"context"
	"example.com/internal/storage"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.postgres.New"

	config, err := pgxpool.ParseConfig(storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s - %w", op, err)
	}
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = time.Minute * 30

	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("%s - %w", op, err)
	}

	err = db.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("%s - %w", op, err)
	}

	_, err = db.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS url (
            id SERIAL PRIMARY KEY,
            alias TEXT NOT NULL UNIQUE,
            url TEXT NOT NULL
        );
        
        CREATE INDEX IF NOT EXISTS idx_alias ON url (alias);
    `)
	if err != nil {
		return nil, fmt.Errorf("%s - %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) StoreURL(ctx context.Context, urlToSave, alias string) (int64, error) {
	const op = "storage.postgres.StoreURL"
	var id int64
	err := s.db.QueryRow(ctx, `INSERT INTO url (url, alias) VALUES ($1, $2) RETURNING id`, urlToSave, alias).Scan(&id)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" {
				return 0, fmt.Errorf("%s - unique constraint violated: %w", op, err)
			}
		}
		return 0, fmt.Errorf("%s - %w", op, err)
	}
	return id, nil
}

func (s *Storage) GetURL(ctx context.Context, alias string) (string, error) {
	const op = "storage.postgres.GetURL"
	var url string
	err := s.db.QueryRow(ctx, `SELECT url FROM url WHERE alias = $1`, alias).Scan(&url)
	if err != nil {
		if pgx.ErrNoRows == err {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s - %w", op, err)
	}
	return url, nil
}

func (s *Storage) DeleteURL(ctx context.Context, alias string) error {
	const op = "storage.postgres.DeleteURL"

	res, err := s.db.Exec(ctx, `DELETE FROM url WHERE alias = $1`, alias)
	if err != nil {
		return fmt.Errorf("%s - %w", op, err)
	}
	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return storage.ErrURLNotFound
	}

	return nil
}

func (s *Storage) Close() {
	s.db.Close()
}
