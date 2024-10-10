package data

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func OpenDB(url string) (*gorm.DB, error) {
	conn, err := sql.Open("pgx", url)
	if err != nil {
		return nil, err
	}

	return gorm.Open(postgres.New(postgres.Config{
		Conn: conn,
	}), &gorm.Config{TranslateError: true, NowFunc: func() time.Time {
		return time.Now().UTC()
	}})
}
