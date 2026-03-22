package exchange

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	exchangepb "github.com/RAF-SI-2025/Banka-3-Backend/gen/exchange"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newTestServer(t *testing.T) (*Server, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	return NewServer(db, gormDB), mock, db
}

func TestConvertMoney(t *testing.T) {
	s, mock, db := newTestServer(t)
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	ctx := context.Background()
	now := time.Now()

	t.Run("Success_EUR_to_USD", func(t *testing.T) {
		mock.ExpectQuery(`SELECT (.+) FROM exchange_rates WHERE currency_code = \$1`).
			WithArgs("EUR").WillReturnRows(sqlmock.NewRows([]string{"c", "r", "u"}).AddRow("EUR", 117.0, now))
		mock.ExpectQuery(`SELECT (.+) FROM exchange_rates WHERE currency_code = \$1`).
			WithArgs("USD").WillReturnRows(sqlmock.NewRows([]string{"c", "r", "u"}).AddRow("USD", 108.0, now))

		resp, err := s.ConvertMoney(ctx, &exchangepb.ConversionRequest{
			FromCurrency: "EUR",
			ToCurrency:   "USD",
			Amount:       100,
		})
		assert.NoError(t, err)
		assert.InDelta(t, 108.333333, resp.ConvertedAmount, 0.0001)
		assert.InDelta(t, 1.083333, resp.ExchangeRate, 0.0001)
	})

	t.Run("Success_RSD_Base", func(t *testing.T) {
		mock.ExpectQuery(`SELECT (.+) FROM exchange_rates WHERE currency_code = \$1`).
			WithArgs("EUR").WillReturnRows(sqlmock.NewRows([]string{"c", "r", "u"}).AddRow("EUR", 117.0, now))

		resp, err := s.ConvertMoney(ctx, &exchangepb.ConversionRequest{
			FromCurrency: "RSD",
			ToCurrency:   "EUR",
			Amount:       1170,
		})
		assert.NoError(t, err)
		assert.InDelta(t, 10.0, resp.ConvertedAmount, 0.0000001)
	})

	t.Run("InvalidAmount", func(t *testing.T) {
		_, err := s.ConvertMoney(ctx, &exchangepb.ConversionRequest{Amount: 0})
		assert.Error(t, err)
	})

	t.Run("CurrencyNotFound", func(t *testing.T) {
		mock.ExpectQuery(`SELECT (.+) FROM exchange_rates WHERE currency_code = \$1`).
			WithArgs("XXX").WillReturnError(sql.ErrNoRows)
		_, err := s.ConvertMoney(ctx, &exchangepb.ConversionRequest{FromCurrency: "XXX", ToCurrency: "USD", Amount: 10})
		assert.Error(t, err)
	})
}

func TestGetExchangeRates(t *testing.T) {
	s, mock, db := newTestServer(t)
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	ctx := context.Background()
	now := time.Now()

	t.Run("DatabaseError", func(t *testing.T) {
		mock.ExpectQuery(`SELECT (.+) FROM exchange_rates`).WillReturnError(fmt.Errorf("db error"))
		_, err := s.GetExchangeRates(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("Success_WithExistingRates", func(t *testing.T) {
		mock.ExpectQuery(`SELECT (.+) FROM exchange_rates`).
			WillReturnRows(sqlmock.NewRows([]string{"c", "r", "u"}).
				AddRow("EUR", 117.0, now).
				AddRow("USD", 108.0, now))

		resp, err := s.GetExchangeRates(ctx, nil)
		assert.NoError(t, err)
		assert.Len(t, resp.Rates, 3)
		assert.Equal(t, now.Unix(), resp.LastUpdated)
	})
}

func TestFetchAndStoreRates_Errors(t *testing.T) {
	s, _, db := newTestServer(t)
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	t.Run("MissingAPIKey", func(t *testing.T) {
		_ = os.Unsetenv("EXCHANGE_RATE_API_KEY")
		err := s.fetchAndStoreRates()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing EXCHANGE_RATE_API_KEY")
	})

	t.Run("InvalidAPIResponse", func(t *testing.T) {
		_ = os.Setenv("EXCHANGE_RATE_API_KEY", "test")
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result":"error"}`))
		}))
		defer ts.Close()

	})
}

func TestUpdateRatesRecord(t *testing.T) {
	s, mock, db := newTestServer(t)
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	t.Run("TransactionFailure", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(fmt.Errorf("tx fail"))
		err := s.UpdateRatesRecord([]Rate{{CurrencyCode: "EUR"}})
		assert.Error(t, err)
	})

	t.Run("ExecFailure", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO exchange_rates").WillReturnError(fmt.Errorf("exec fail"))
		mock.ExpectRollback()
		err := s.UpdateRatesRecord([]Rate{{CurrencyCode: "EUR", RateToRSD: 117}})
		assert.Error(t, err)
	})

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO exchange_rates").
			WithArgs("EUR", 117.0).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		err := s.UpdateRatesRecord([]Rate{{CurrencyCode: "EUR", RateToRSD: 117}})
		assert.NoError(t, err)
	})
}
