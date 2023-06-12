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
	Id    int
	Path  string
	Paths []GetFolderSizeRequest
}

type GetFolderSizeResponseRequest struct {
	Id       int
	Size     int64
	Duration float64
}

type GetFolderSizeRequestedResponse struct {
	Status bool
	Data   string
}

func GetFolderSizeQueueJob(req GetFolderSizeRequest, appAuth string, folder *Folder) Job {
	return Job{
		Name: req.Path,
		Action: func() error {
			id := req.Id
			path := req.Path
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
			appRequest.Header.Set("X-APP-AUTH", appAuth)

			client := &http.Client{}
			_, err := client.Do(appRequest)
			return err
		},
	}
}

func (folder *Folder) getSize(w http.ResponseWriter, req *http.Request) {
	var body GetFolderSizeRequest

	secret := req.Header.Get("X-SECRET")
	appAuth := req.Header.Get("X-APP-AUTH")

	if secret != folder.config.appSecret {
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
				folder.worker.Queue.AddJob(GetFolderSizeQueueJob(path, appAuth, folder))
			}
		}
	} else {
		if body.Id == 0 {
			resultMessage = "Id is required"
		} else if body.Path == "" {
			resultMessage = "Path is required"
		} else {
			folder.worker.Queue.AddJob(GetFolderSizeQueueJob(body, appAuth, folder))
		}
	}

	var requestAcceptedResponse GetFolderSizeRequestedResponse
	requestAcceptedResponse.Status = true
	requestAcceptedResponse.Data = resultMessage
	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(requestAcceptedResponse)
	if err != nil {
		log.Printf("Failed to send request: %v", err)
	}
}

func (folder *Folder) serve(addr string) {
	go func() {
		folder.worker = &Worker{NewQueue("file-storages")}
		folder.worker.DoWork()
	}()

	http.HandleFunc("/get", folder.getSize)
	log.Printf("Listening on %s", addr)
	err := http.ListenAndServe(addr, nil)

	if err != nil {
		log.Fatalf("Cannot start http server: %v", err)
	}
}
