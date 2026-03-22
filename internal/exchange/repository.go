package exchange

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrRateNotFound = errors.New("exchange rate not found")

func scanRate(scanner interface {
	Scan(dest ...any) error
}) (*Rate, error) {
	var r Rate
	err := scanner.Scan(
		&r.CurrencyCode,
		&r.RateToRSD,
		&r.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Server) GetRatesRecord() ([]Rate, error) {
	rows, err := s.database.Query(`SELECT currency_code, rate_to_rsd, updated_at FROM exchange_rates`)
	if err != nil {
		return nil, fmt.Errorf("querying rates: %w", err)
	}

	defer func(rows *sql.Rows) { _ = rows.Close() }(rows)

	var rates []Rate
	for rows.Next() {
		r, err := scanRate(rows)
		if err != nil {
			return nil, err
		}
		rates = append(rates, *r)
	}
	return rates, nil
}

func (s *Server) GetRateByCodeRecord(code string) (*Rate, error) {
	if code == "RSD" {
		return &Rate{CurrencyCode: "RSD", RateToRSD: 1.0, UpdatedAt: time.Now()}, nil
	}

	row := s.database.QueryRow(`SELECT currency_code, rate_to_rsd, updated_at FROM exchange_rates WHERE currency_code = $1`, code)
	r, err := scanRate(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRateNotFound
		}
		return nil, err
	}
	return r, nil
}

func (s *Server) UpdateRatesRecord(rates []Rate) error {
	tx, err := s.database.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	for _, r := range rates {
		_, err := tx.Exec(`
            INSERT INTO exchange_rates (currency_code, rate_to_rsd, updated_at)
            VALUES ($1, $2, NOW())
            ON CONFLICT (currency_code)
            DO UPDATE SET rate_to_rsd = EXCLUDED.rate_to_rsd, updated_at = NOW()
        `, r.CurrencyCode, r.RateToRSD)
		if err != nil {
			return fmt.Errorf("failed to update %s: %w", r.CurrencyCode, err)
		}
	}
	return tx.Commit()
}
