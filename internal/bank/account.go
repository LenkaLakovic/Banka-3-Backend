package bank

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	bankpb "github.com/RAF-SI-2025/Banka-3-Backend/gen/bank"
)

func (s *Server) UpdateAccountName(_ context.Context, req *bankpb.UpdateAccountNameRequest) (*bankpb.UpdateAccountNameResponse, error) {
	accountNumber := strings.TrimSpace(req.AccountNumber)
	name := strings.TrimSpace(req.Name)

	if accountNumber == "" {
		return nil, status.Error(codes.InvalidArgument, "account number is required")
	}
	if name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	account, err := s.GetAccountByNumberRecord(accountNumber)
	if err != nil {
		return nil, status.Error(codes.NotFound, "account not found")
	}

	exists, err := s.AccountNameExistsForOwner(account.Owner, name, accountNumber)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to check account name")
	}
	if exists {
		return nil, status.Error(codes.InvalidArgument, "name is already used by another account belonging to the customer")
	}

	if err := s.UpdateAccountNameRecord(accountNumber, name); err != nil {
		if err.Error() == "account not found" {
			return nil, status.Error(codes.NotFound, "account not found")
		}
		return nil, status.Error(codes.Internal, "failed to update account name")
	}

	return &bankpb.UpdateAccountNameResponse{}, nil
}

func (s *Server) UpdateAccountLimits(_ context.Context, req *bankpb.UpdateAccountLimitsRequest) (*bankpb.UpdateAccountLimitsResponse, error) {
	accountNumber := strings.TrimSpace(req.AccountNumber)
	if accountNumber == "" {
		return nil, status.Error(codes.InvalidArgument, "account number is required")
	}

	if req.DailyLimit == nil && req.MonthlyLimit == nil {
		return nil, status.Error(codes.InvalidArgument, "at least one limit must be provided")
	}

	_, err := s.GetAccountByNumberRecord(accountNumber)
	if err != nil {
		return nil, status.Error(codes.NotFound, "account not found")
	}

	if req.DailyLimit != nil && *req.DailyLimit < 0 {
		return nil, status.Error(codes.InvalidArgument, "daily_limit must be non-negative")
	}

	if req.MonthlyLimit != nil && *req.MonthlyLimit < 0 {
		return nil, status.Error(codes.InvalidArgument, "monthly_limit must be non-negative")
	}

	if err := s.UpdateAccountLimitsRecord(accountNumber, req.DailyLimit, req.MonthlyLimit); err != nil {
		if err.Error() == "account not found" {
			return nil, status.Error(codes.NotFound, "account not found")
		}
		return nil, status.Error(codes.Internal, "failed to update account limits")
	}

	return &bankpb.UpdateAccountLimitsResponse{}, nil
}
