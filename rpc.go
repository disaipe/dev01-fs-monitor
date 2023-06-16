package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

type Rpc struct {
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

func GetStorageSizeQueueJob(storage *FileStorage) Job {
	return Job{
		Name: storage.path,
		Action: func() error {
			storage.getSizeProcess()
			// storage.getSize()
			return nil
		},
	}
}

func (rpc *Rpc) serve(addr string) {
	go func() {
		rpc.worker = &Worker{NewQueue("file-storages")}
		rpc.worker.DoWork()
	}()

	http.HandleFunc("/get", rpc.getSizeRequest)
	log.Printf("Listening on %s", addr)
	err := http.ListenAndServe(addr, nil)

	if err != nil {
		log.Fatalf("Cannot start http server: %v", err)
	}
}

func (rpc *Rpc) getSizeRequest(w http.ResponseWriter, req *http.Request) {
	var body GetStorageSizeRequest

	secret := req.Header.Get("X-SECRET")
	appAuth := req.Header.Get("X-APP-AUTH")

	if secret != AppConfig.appSecret {
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
				storage := &FileStorage{
					id:      path.Id,
					path:    path.Path,
					appAuth: appAuth,
				}

				rpc.worker.Queue.AddJob(GetStorageSizeQueueJob(storage))
			}
		}
	} else {
		if body.Id == 0 {
			resultMessage = "Id is required"
		} else if body.Path == "" {
			resultMessage = "Path is required"
		} else {
			storage := &FileStorage{
				id:      body.Id,
				path:    body.Path,
				appAuth: appAuth,
			}

			rpc.worker.Queue.AddJob(GetStorageSizeQueueJob(storage))
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

func (rpc *Rpc) sendResult(id int, size int64, duration float64, appAuth string) error {
	var response GetStorageSizeResponseRequest
	response.Id = id
	response.Size = size
	response.Duration = duration

	v, _ := json.Marshal(response)

	appRequest, _ := http.NewRequest("POST", AppConfig.appUrl, bytes.NewBuffer(v))
	appRequest.Header.Set("Content-Type", "application/json")
	appRequest.Header.Set("X-APP-AUTH", appAuth)

	client := &http.Client{}
	_, err := client.Do(appRequest)
	if err != nil {
		return err
	}

	return nil
}
