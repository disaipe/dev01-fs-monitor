package main

import (
	"bytes"
	"encoding/json"
	getFolderSize "github.com/markthree/go-get-folder-size/src"
	"log"
	"net/http"
	"time"
)

type Config struct {
	addr      string
	appUrl    string
	appSecret string
}

type Folder struct {
	config *Config
	worker *Worker
}

type GetFolderSizeRequest struct {
	Id   int
	Path string
}

type GetFolderSizeResponseRequest struct {
	Id       int
	Size     int64
	Duration float64
}

type GetFolderSizeRequestedResponse struct {
	Status bool
}

func (folder *Folder) getSize(w http.ResponseWriter, req *http.Request) {
	var body GetFolderSizeRequest

	secret := req.Header.Get("X-SECRET")

	if secret != folder.config.appSecret {
		http.Error(w, "Wrong secret", http.StatusUnauthorized)
		return
	}

	err := json.NewDecoder(req.Body).Decode(&body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if body.Id == 0 {
		http.Error(w, "Id is required", http.StatusBadRequest)
		return
	}

	folder.worker.Queue.AddJob(Job{
		Name: body.Path,
		Action: func() error {
			id := body.Id
			path := body.Path
			timeStart := time.Now()
			size := getFolderSize.LooseParallel(path)
			duration := time.Now().Sub(timeStart).Seconds()

			log.Printf("Result: %v (id: %v) - %v (%.2f)", path, id, size, duration)

			var response GetFolderSizeResponseRequest
			response.Id = id
			response.Size = size
			response.Duration = duration

			v, _ := json.Marshal(response)

			appRequest, _ := http.NewRequest("POST", folder.config.appUrl, bytes.NewBuffer(v))
			appRequest.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			_, err := client.Do(appRequest)
			return err
		},
	})

	var requestAcceptedResponse GetFolderSizeRequestedResponse
	requestAcceptedResponse.Status = true
	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(requestAcceptedResponse)
	if err != nil {
		return
	}
}

func (folder *Folder) serve(addr string) {
	folder.worker = &Worker{NewQueue("file-storages")}

	go func() {
		folder.worker.DoWork()
	}()

	http.HandleFunc("/get", folder.getSize)
	log.Printf("Listening on %s", addr)
	err := http.ListenAndServe(addr, nil)

	if err != nil {
		log.Fatalf("Cannot start http server: %v", err)
	}
}
