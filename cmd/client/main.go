package main

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/dlsteuer/ds-sync/pb"
	"github.com/dlsteuer/ds-sync/server"
	"google.golang.org/grpc"
)

func main() {
	g, err := grpc.Dial(":8079", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	c := pb.NewServiceDSSyncClient(g)
	resp, err := c.ListCompleted(context.Background(), &pb.ListCompletedRequest{})
	if err != nil {
		panic(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	basePath := filepath.Join(wd, "downloads")
	for _, f := range resp.Files {
		dc, err := c.DownloadFile(context.Background(), &pb.DownloadFileRequest{
			File: f,
		})
		if err != nil {
			panic(err)
		}

		file := filepath.Join(basePath, f)
		dir := filepath.Dir(file)
		os.MkdirAll(dir, os.ModePerm)
		log.Printf("downloading file: %s", file)
		fh, err := os.Create(file)
		if err != nil {
			panic(err)
		}

		for {
			chunk, err := dc.Recv()
			if err != nil {
				panic(err)
			}

			incHash := server.CalcBlake2b(chunk.Data)
			if bytes.Compare(incHash, chunk.Blake2B) != 0 {
				panic(errors.New("file hash invalid"))
			}

			_, err = fh.Write(chunk.Data)
			if err != nil {
				panic(err)
			}

			if chunk.IsLastChunk {
				break
			}
		}
	}

}
