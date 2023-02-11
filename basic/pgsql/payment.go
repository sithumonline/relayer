package pgsql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/fiatjaf/relayer/storage"
	"github.com/fiatjaf/relayer/storage/postgresql"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
)

type Payment struct {
	PubKey       string
	CreatedAt    int64
	ExpirationAt int64
	TxHex        string
}

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
        VALUES ($1, $2, $3, $4)
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

func (b *BasicPostgresBackend) CheckPayment(pubKey string) (bool, error) {
	rows, err := b.DB.Query(`
		SELECT * FROM payment
		WHERE created_at = (
			SELECT MAX (created_at)
			FROM payment
		) AND pubkey = $1
    `, pubKey)
	if err != nil && err != sql.ErrNoRows {
		return false, fmt.Errorf("failed to fetch payment: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var payment Payment
		err := rows.Scan(&payment.PubKey, &payment.CreatedAt, &payment.ExpirationAt, &payment.TxHex)
		if err != nil {
			return false, fmt.Errorf("failed to scan payment row: %w", err)
		}
		if payment.ExpirationAt > time.Now().Unix() {
			return true, nil
		} else {
			return false, nil
		}
	}

	return false, nil
}
