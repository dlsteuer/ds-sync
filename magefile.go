// +build mage

package main

import (
	"fmt"
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var Default = Build

func BuildServer() error {
	fmt.Println("Building Server...")
	return sh.Run("go", "build", "-o", ".build/ds-sync-server", "./cmd/server")
}

func BuildClient() error {
	fmt.Println("Building Client...")
	return sh.Run("go", "build", "-o", ".build/ds-sync-client", "./cmd/client")
}

func Proto() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	return sh.Run("docker", "run", "-it", "--rm", "-v", fmt.Sprintf("%s/pb:/build/pb", wd), "sendwithus/protoc",
		"-I", "/build/pb", "--gogo_out=Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,plugins=grpc:/build/pb", "/build/pb/server.proto")
}

func Build() {
	mg.Deps(BuildServer, BuildClient)
}

func Clean() {
	fmt.Println("Cleaning...")
	os.RemoveAll(".build/")
}
