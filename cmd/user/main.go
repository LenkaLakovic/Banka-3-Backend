package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"banka-raf/gen/user"
	internalUser "banka-raf/internal/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	accessJwtSecret, accessSecretSet := os.LookupEnv("ACCESS_JWT_SECRET")
	refreshJwtSecret, refreshSecretSet := os.LookupEnv("REFRESH_JWT_SECRET")
	if accessSecretSet == false || refreshSecretSet == false {
		log.Fatalf("JWT secrets not set, exiting...")
	}

	userService := internalUser.NewServer(accessJwtSecret, refreshJwtSecret)

	srv := grpc.NewServer()
	user.RegisterUserServiceServer(srv, userService)
	reflection.Register(srv)

	log.Printf("user service listening on :%s", port)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
