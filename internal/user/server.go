package user

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"banka-raf/gen/user"
	userpb "banka-raf/gen/user"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	userpb.UnimplementedUserServiceServer
	accessJwtSecret  string
	refreshJwtSecret string
}

func NewServer(accessJwtSecret string, refreshJwtSecret string) *Server {
	return &Server{
		accessJwtSecret:  accessJwtSecret,
		refreshJwtSecret: refreshJwtSecret,
	}
}

func (s *Server) GetEmployeeById(ctx context.Context, req *userpb.GetEmployeeByIdRequest) (*user.EmployeeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func generateRefreshToken() {}

func generateAccessToken(userID uint64, secret string) (string, error) {

	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Minute * 15).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func (s *Server) Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	hasher := sha256.New()
	hasher.Write([]byte(req.Password))
	byteSlice := hasher.Sum(nil)
	encodedHashString := hex.EncodeToString(byteSlice)
	user := GetUserByEmail(req.Email)
	if encodedHashString == user.hashedPassword {
		accessToken, err := generateAccessToken(user.id, s.accessJwtSecret)
		if err != nil {
			return nil, err
		}

		return &userpb.LoginResponse{
			AccessToken: accessToken,
		}, nil
	}

	return &userpb.LoginResponse{
		AccessToken: "",
	}, errors.New("wrong creds")
}
