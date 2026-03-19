// shared testing utils

package bank

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func newTestServer(t *testing.T) (*Server, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}

	return NewServer(db, nil), mock, db
}
