package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

	err = downloadFiles(c, resp.Files)
	if err != nil {
		log.Fatalf("Error while downloading files: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error grabbing working directory: %v", err)
	}
	basePath := filepath.Join(wd, "downloads")
	completed, err := ioutil.ReadDir(basePath)
	if err != nil {
		log.Fatalf("Error reading directory: %v", err)
	}

	for _, c := range completed {
		err = importIntoSonarr(c.Name())
		if err != nil {
			log.Fatalf("Error while importing into sonarr: %v", err)
		}
		err = os.RemoveAll(c.Name())
		if err != nil {
			log.Fatalf("Error while importing into sonarr: %v", err)
		}
	}
}

func importIntoSonarr(dir string) error {
	client := http.DefaultClient
	buf := &bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("{\"name\":\"downloadedepisodesscan\", \"path\":\"%s\"}", dir))
	_, err := client.Post("http://localhost:8989", "application/json", buf)
	if err != nil {
		return err
	}
	return nil
}

func downloadFiles(c pb.ServiceDSSyncClient, files []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	basePath := filepath.Join(wd, "downloads")
	for _, f := range files {
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
			return err
		}

		for {
			chunk, err := dc.Recv()
			if err != nil {
				return err
			}

			incHash := server.CalcBlake2b(chunk.Data)
			if bytes.Compare(incHash, chunk.Blake2B) != 0 {
				return errors.New("file hash invalid")
			}

			_, err = fh.Write(chunk.Data)
			if err != nil {
				return err
			}

			if chunk.IsLastChunk {
				break
			}
		}
	}
	return nil
}
