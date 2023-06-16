package main

import (
	"fmt"
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

func (storage *FileStorage) getSize() {
	timeStart := time.Now()
	size := getFolderSize.LooseParallel(storage.path)

	duration := time.Now().Sub(timeStart).Seconds()

	log.Printf("Result: %v (id: %v) - %v (%.2f)", storage.path, storage.id, size, duration)

	if AppConfig.appUrl != "" {
		rpc := Rpc{}
		err := rpc.sendResult(storage.id, size, duration, storage.appAuth)

		if err != nil {
			log.Printf("Failed to send results: %v", err)
		} else {
			log.Printf("Results sent to %s", AppConfig.appUrl)
		}
	}
}

func (storage *FileStorage) getSizeProcess() {
	ex, _ := os.Executable()

	idArg := fmt.Sprintf("-id=%v", storage.id)
	pathArg := fmt.Sprintf("-path=%s", storage.path)
	authArg := fmt.Sprintf("-auth=%s", storage.appAuth)
	appUrl := fmt.Sprintf("-app.url=%s", AppConfig.appUrl)

	cmd := exec.Command(ex, idArg, pathArg, authArg, appUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		return
	}
}
