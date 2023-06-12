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

type FileStorage struct {
	config *Config
	worker *Worker
}

type GetStorageSizeRequest struct {
	Id    int
	Path  string
	Paths []GetStorageSizeRequest
}

type GetStorageSizeResponseRequest struct {
	Id       int
	Size     int64
	Duration float64
}

type GetStorageSizeRequestedResponse struct {
	Status bool
	Data   string
}

func GetStorageSizeQueueJob(req GetStorageSizeRequest, appAuth string, storage *FileStorage) Job {
	return Job{
		Name: req.Path,
		Action: func() error {
			id := req.Id
			path := req.Path
			timeStart := time.Now()
			size := getFolderSize.LooseParallel(path)
			duration := time.Now().Sub(timeStart).Seconds()

			log.Printf("Result: %v (id: %v) - %v (%.2f)", path, id, size, duration)

			var response GetStorageSizeResponseRequest
			response.Id = id
			response.Size = size
			response.Duration = duration

			v, _ := json.Marshal(response)

			appRequest, _ := http.NewRequest("POST", storage.config.appUrl, bytes.NewBuffer(v))
			appRequest.Header.Set("Content-Type", "application/json")
			appRequest.Header.Set("X-APP-AUTH", appAuth)

			client := &http.Client{}
			_, err := client.Do(appRequest)
			return err
		},
	}
}

func (storage *FileStorage) getSize(w http.ResponseWriter, req *http.Request) {
	var body GetStorageSizeRequest

	secret := req.Header.Get("X-SECRET")
	appAuth := req.Header.Get("X-APP-AUTH")

	if secret != storage.config.appSecret {
		http.Error(w, "Wrong secret", http.StatusUnauthorized)
		return
	}

	err := json.NewDecoder(req.Body).Decode(&body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var resultMessage string

	if body.Paths != nil {
		for _, path := range body.Paths {
			if path.Id != 0 && path.Path != "" {
				storage.worker.Queue.AddJob(GetStorageSizeQueueJob(path, appAuth, storage))
			}
		}
	} else {
		if body.Id == 0 {
			resultMessage = "Id is required"
		} else if body.Path == "" {
			resultMessage = "Path is required"
		} else {
			storage.worker.Queue.AddJob(GetStorageSizeQueueJob(body, appAuth, storage))
		}
	}

	var requestAcceptedResponse GetStorageSizeRequestedResponse
	requestAcceptedResponse.Status = true
	requestAcceptedResponse.Data = resultMessage
	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(requestAcceptedResponse)
	if err != nil {
		log.Printf("Failed to send request: %v", err)
	}
}

func (storage *FileStorage) serve(addr string) {
	go func() {
		storage.worker = &Worker{NewQueue("file-storages")}
		storage.worker.DoWork()
	}()

	http.HandleFunc("/get", storage.getSize)
	log.Printf("Listening on %s", addr)
	err := http.ListenAndServe(addr, nil)

	if err != nil {
		log.Fatalf("Cannot start http server: %v", err)
	}
}
