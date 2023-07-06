package main

import (
	"fmt"
	rpc "github.com/disaipe/dev01-rpc-base"
	getFolderSize "github.com/markthree/go-get-folder-size/src"
	"log"
	"os"
	"os/exec"
	"time"
)

type FileStorage struct {
	id      int
	path    string
	appAuth string
}

// Start processing
func (storage *FileStorage) getSize() {
	timeStart := time.Now()
	size := getFolderSize.LooseParallel(storage.path)

	duration := time.Now().Sub(timeStart).Seconds()

	log.Printf("Result: %v (id: %v) - %v (%.2f)", storage.path, storage.id, size, duration)

	if rpc.Config.GetAppUrl() != "" {
		rpcInstance := rpc.Rpc{}
		response := &GetStorageSizeResponse{
			Id:       storage.id,
			Size:     size,
			Duration: duration,
		}

		rpcInstance.SendResult(response, storage.appAuth)
	}
}

// Start processing in other process
func (storage *FileStorage) getSizeProcess() error {
	ex, _ := os.Executable()

	idArg := fmt.Sprintf("-id=%v", storage.id)
	pathArg := fmt.Sprintf("-path=%s", storage.path)
	authArg := fmt.Sprintf("-auth=%s", storage.appAuth)
	appUrl := fmt.Sprintf("-app.url=%s", rpc.Config.GetAppUrl())

	cmd := exec.Command(ex, idArg, pathArg, authArg, appUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
