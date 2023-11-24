// main.go

package main

import (
	"log"
	"net"
	"url_shortener/api/pb"
	"url_shortener/pkg/db"
	"url_shortener/pkg/services"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	redisClient := db.CreateClient(0) // Assuming DB 0 is used for URL shortener data
	defer redisClient.Close()

	urlShortenerServer := services.NewURLShortenerServer(redisClient)

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)

	pb.RegisterURLShortenerServer(grpcServer, urlShortenerServer)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Println("gRPC server is running on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
