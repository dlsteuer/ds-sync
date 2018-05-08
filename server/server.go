package server

import (
	"bufio"
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/dlsteuer/ds-sync/pb"
	"golang.org/x/crypto/blake2b"
)

type SyncServer struct {
	Watch string
}

func (ss *SyncServer) ListCompleted(ctx context.Context, req *pb.ListCompletedRequest) (*pb.ListCompletedResponse, error) {
	files, err := getFilesFromDir(ss.Watch)
	if err != nil {
		return nil, err
	}
	transformed := []string{}
	for _, f := range files {
		transformed = append(transformed, strings.Replace(f, ss.Watch, "", 1))
	}
	return &pb.ListCompletedResponse{
		Files: transformed,
	}, nil
}

func (ss *SyncServer) DownloadFile(req *pb.DownloadFileRequest, stream pb.ServiceDSSync_DownloadFileServer) error {
	path := filepath.Join(ss.Watch, req.File)
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanBytes)
	chunks := make(chan *pb.DownloadFileResponse)
	go func() {
		buf := &bytes.Buffer{}
		for scanner.Scan() {
			buf.Write(scanner.Bytes())
			if buf.Len() > 1000000 { // 1 MB
				b := buf.Bytes()
				chunks <- &pb.DownloadFileResponse{
					SizeInBytes: int64(len(b)),
					Data:        b,
					IsLastChunk: false,
					Blake2B:     CalcBlake2b(b),
				}
				buf.Reset()
			}
		}

		b := buf.Bytes()
		chunks <- &pb.DownloadFileResponse{
			SizeInBytes: int64(len(b)),
			Data:        b,
			IsLastChunk: true,
			Blake2B:     CalcBlake2b(b),
		}
	}()

	for c := range chunks {
		err := stream.Send(c)
		if err != nil {
			return err
		}
	}

	return nil
}

func CalcBlake2b(b []byte) []byte {
	h, _ := blake2b.New512(nil)
	h.Write(b)
	return []byte(h.Sum(nil))
}

func getFilesFromDir(dir string) ([]string, error) {
	completed, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := []string{}
	for _, c := range completed {
		name := filepath.Join(dir, c.Name())
		if c.IsDir() {
			dirFiles, err := getFilesFromDir(name)
			if err != nil {
				return nil, err
			}
			files = append(files, dirFiles...)
		} else {
			files = append(files, name)
		}
	}
	return files, nil
}
