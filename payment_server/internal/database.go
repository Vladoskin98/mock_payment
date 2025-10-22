package internal

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

type Payment struct {
	ID           int
	ProviderName string
	Amount       float64
	PaymentDate  string
	CreatedAt    time.Time
}

type DB struct {
	*sql.DB
}

func NewDB() (*DB, error) {
	connStr := "host=postgres port=5432 user=postgres password=admin dbname=payments_db sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db *DB) AddPayment(providerName string, amount float64, paymentDate string) (int64, error) {
	query := `INSERT INTO payments (provider_name, amount, payment_date) VALUES ($1, $2, $3) RETURNING id`

	var id int64

	err := db.QueryRow(query, providerName, amount, paymentDate).Scan(&id)

	return id, err
}

// Удаление записей, существующих дольше 10 минут
func (db *DB) DeleteExpiredPayments() (int64, error) {
	query := `DELETE FROM payments WHERE created_at < NOW() - INTERVAL '10 minutes'`
	result, err := db.Exec(query)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
