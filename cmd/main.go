package main

import (
	"log"
	"net"

	"grpc-bookauthor/cmd/config"
	"grpc-bookauthor/cmd/services"
	pb "grpc-bookauthor/proto"

	"google.golang.org/grpc"
)

const (
	PORT = ":3000"
)

func main() {
	listen, err := net.Listen("tcp", PORT)

	if err != nil {
		log.Fatalf("Failed to listen %v", err.Error())
	}

	db := config.ConnectDatabase()

	grpcServer := grpc.NewServer()
	bookService := services.BookService{DB: db}
	pb.RegisterBookServiceServer(grpcServer, &bookService)

	log.Printf("Server started at %v", listen.Addr())
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("Failed to serve %v", err.Error())
	}
}
