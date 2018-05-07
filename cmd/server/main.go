package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/dlsteuer/ds-sync/pb"
	"github.com/dlsteuer/ds-sync/server"
	"google.golang.org/grpc"
)

func main() {
	watch := flag.String("watch", "", "completion folder")
	flag.Parse()

	grpcServer := grpc.NewServer()
	pb.RegisterServiceDSSyncServer(grpcServer, &server.SyncServer{
		Watch: *watch,
	})
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", "8079"))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("listening on port 8079")
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Println("failed to serve: %v", err)
	}
}
