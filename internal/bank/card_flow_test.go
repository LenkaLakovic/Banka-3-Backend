package bank

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	bankpb "github.com/RAF-SI-2025/Banka-3-Backend/gen/bank"
	notificationpb "github.com/RAF-SI-2025/Banka-3-Backend/gen/notification"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type testNotificationServer struct {
	notificationpb.UnimplementedNotificationServiceServer
}

func (s *testNotificationServer) SendCardCreatedEmail(_ context.Context, _ *notificationpb.CardCreatedMailRequest) (*notificationpb.SuccessResponse, error) {
	return &notificationpb.SuccessResponse{Successful: true}, nil
}

func (s *testNotificationServer) SendCardConfirmationEmail(_ context.Context, _ *notificationpb.CardConfirmationMailRequest) (*notificationpb.SuccessResponse, error) {
	return &notificationpb.SuccessResponse{Successful: true}, nil
}

func TestCreateCardSuccess(t *testing.T) {
	server, mock, db := newTestServer(t)
	defer db.Close()

	mock.ExpectQuery("INSERT INTO cards").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "number", "type", "brand", "creation_date",
			"valid_until", "account_number", "cvv", "card_limit", "status",
		}).AddRow(
			int64(1), "1234567890123456", "debit", "visa",
			time.Now(), time.Now().AddDate(5, 0, 0),
			"123456789", "123", 0, "active",
		))

	resp, err := server.CreateCard(context.Background(), &bankpb.CreateCardRequest{
		CardType:      "debit",
		CardBrand:     "visa",
		AccountNumber: "123456789",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CardNumber == "" {
		t.Fatalf("expected card number")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestRequestCardSuccess(t *testing.T) {
	server, mock, db := newTestServer(t)
	defer db.Close()

	notificationServer := &testNotificationServer{}
	addr, stop := startNotificationTestServer(t, notificationServer)
	defer stop()

	t.Setenv("NOTIFICATION_GRPC_ADDR", addr)

	ctx := metadata.NewIncomingContext(context.Background(),
		metadata.Pairs("user-email", "test@mail.com"),
	)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, number, name, owner, balance, currency, active, owner_type, account_type,
		       maintainance_cost, daily_limit, monthly_limit, daily_expenditure, monthly_expenditure,
		       created_by, created_at, valid_until
		FROM accounts WHERE number = $1`)).
		WithArgs("123456789").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "number", "name", "owner", "balance", "currency", "active",
			"owner_type", "account_type", "maintainance_cost",
			"daily_limit", "monthly_limit", "daily_expenditure", "monthly_expenditure",
			"created_by", "created_at", "valid_until",
		}).AddRow(
			1, "123456789", "acc", 1, 0, "RSD", true,
			"person", "checking", 0,
			0, 0, 0, 0,
			1, time.Now(), time.Now(),
		))

	mock.ExpectQuery("SELECT EXISTS").
		WillReturnRows(sqlmockBoolRow(false))

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("123456789", Deactivated).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery("INSERT INTO card_requests").
		WillReturnRows(sqlmockCardRequestRows().AddRow(
			1, "123456789", "debit", "visa", "token",
			time.Now().Add(24*time.Hour), false, "test@mail.com",
		))

	resp, err := server.RequestCard(ctx, &bankpb.RequestCardRequest{
		AccountNumber: "123456789",
		CardType:      "debit",
		CardBrand:     "visa",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Accepted {
		t.Fatalf("expected accepted true")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestRequestCardLimitReached(t *testing.T) {
	server, mock, db := newTestServer(t)
	defer db.Close()

	ctx := metadata.NewIncomingContext(context.Background(),
		metadata.Pairs("user-email", "test@mail.com"),
	)

	mock.ExpectQuery("FROM accounts WHERE number").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "number", "name", "owner", "balance", "currency", "active",
			"owner_type", "account_type", "maintainance_cost",
			"daily_limit", "monthly_limit", "daily_expenditure", "monthly_expenditure",
			"created_by", "created_at", "valid_until",
		}).AddRow(
			1, "123456789", "acc", 1, 0, "RSD", true,
			"person", "checking", 0,
			0, 0, 0, 0,
			1, time.Now(), time.Now(),
		))

	mock.ExpectQuery("SELECT EXISTS").
		WillReturnRows(sqlmockBoolRow(false))

	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	_, err := server.RequestCard(ctx, &bankpb.RequestCardRequest{
		AccountNumber: "123456789",
	})

	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("expected FailedPrecondition, got %v", status.Code(err))
	}
}

func TestConfirmCardSuccess(t *testing.T) {
	server, mock, db := newTestServer(t)
	defer db.Close()

	notificationServer := &testNotificationServer{}
	addr, stop := startNotificationTestServer(t, notificationServer)
	defer stop()

	t.Setenv("NOTIFICATION_GRPC_ADDR", addr)

	mock.ExpectQuery("SELECT id, account_number").
		WithArgs("token123").
		WillReturnRows(sqlmockCardRequestRows().AddRow(
			1, "123456789", "debit", "visa", "token123",
			time.Now().Add(time.Hour), false, "test@mail.com",
		))

	mock.ExpectQuery("INSERT INTO cards").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "number", "type", "brand", "creation_date",
			"valid_until", "account_number", "cvv", "card_limit", "status",
		}).AddRow(
			int64(1), "123", "debit", "visa",
			time.Now(), time.Now().AddDate(5, 0, 0),
			"123456789", "123", 0, "active",
		))

	mock.ExpectExec("UPDATE card_requests SET complete").
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err := server.ConfirmCard(context.Background(), &bankpb.ConfirmCardRequest{
		Token: "token123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfirmCardExpired(t *testing.T) {
	server, mock, db := newTestServer(t)
	defer db.Close()

	mock.ExpectQuery("SELECT id, account_number").
		WillReturnRows(sqlmockCardRequestRows().AddRow(
			1, "123456789", "debit", "visa", "token123",
			time.Now().Add(-time.Hour), false, "test@mail.com",
		))

	_, err := server.ConfirmCard(context.Background(), &bankpb.ConfirmCardRequest{
		Token: "token123",
	})

	if status.Code(err) != codes.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", status.Code(err))
	}
}

func TestBlockCardSuccess(t *testing.T) {
	server, mock, db := newTestServer(t)
	defer db.Close()

	mock.ExpectExec("UPDATE cards SET status").
		WithArgs(Blocked, int64(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := server.BlockCard(context.Background(), &bankpb.BlockCardRequest{
		CardId: 1,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success true")
	}
}

func TestBlockCardNotFound(t *testing.T) {
	server, mock, db := newTestServer(t)
	defer db.Close()

	mock.ExpectExec("UPDATE cards SET status").
		WithArgs(Blocked, int64(999)).
		WillReturnResult(sqlmock.NewResult(1, 0))

	_, err := server.BlockCard(context.Background(), &bankpb.BlockCardRequest{
		CardId: 999,
	})

	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", status.Code(err))
	}
}
