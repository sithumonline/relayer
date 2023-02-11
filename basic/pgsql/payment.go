package pgsql

import (
	"time"

	"github.com/fiatjaf/relayer/storage"
	"github.com/fiatjaf/relayer/storage/postgresql"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
)

type BasicPostgresBackend struct {
	*postgresql.PostgresBackend
}

func (b *BasicPostgresBackend) PaymentInit() error {
	db, err := sqlx.Connect("postgres", b.DatabaseURL)
	if err != nil {
		return err
	}

	// sqlx default is 0 (unlimited), while postgresql by default accepts up to 100 connections
	db.SetMaxOpenConns(80)

	db.Mapper = reflectx.NewMapperFunc("json", sqlx.NameMapper)
	b.DB = db

	_, err = b.DB.Exec(`
CREATE TABLE IF NOT EXISTS payment (
  pubkey text NOT NULL,
  created_at integer NOT NULL,
  expiration_at integer NOT NULL,
  tx_hash text UNIQUE NOT NULL
);
`)

	return err
}

func (b *BasicPostgresBackend) SavePayment(pubKey string, txHash string) error {
	timeNow := time.Now()
	timeInMonth := 30 * 24 * time.Hour
	res, err := b.DB.Exec(`
        INSERT INTO payment (pubkey, created_at, expiration_at, tx_hash)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO NOTHING
    `, pubKey, timeNow.Unix(), timeNow.Add(timeInMonth).Unix(), txHash)
	if err != nil {
		return err
	}

	nr, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if nr == 0 {
		return storage.ErrDupEvent
	}

	return nil
}
