package main

import (
	"encoding/json"
	"flag"
	"io"
	"runtime/debug"

	rpc "github.com/disaipe/dev01-rpc-base"
)

type GetStorageSizeRequest struct {
	rpc.Response

	Id    int
	Path  string
	Paths []GetStorageSizeRequest
}

type GetStorageSizeResponse struct {
	rpc.ResultResponse

	Id       int
	Size     int64
	Duration float64
}

func GetStorageSizeQueueJob(storage *FileStorage) rpc.Job {
	return rpc.Job{
		Name: storage.path,
		Action: func() error {
			return storage.getSizeProcess()
		},
	}
}

var storageId int
var storagePath string
var storageAuth string

func init() {
	flag.IntVar(&storageId, "id", 0, "Storage id")
	flag.StringVar(&storagePath, "path", "", "Path to get size")
	flag.StringVar(&storageAuth, "auth", "", "Storage result request auth key")
}

var rpcAction = rpc.ActionFunction(func(rpcServer *rpc.Rpc, body io.ReadCloser, appAuth string) (rpc.Response, error) {
	var storageRequest GetStorageSizeRequest

	err := json.NewDecoder(body).Decode(&storageRequest)

	if err != nil {
		return nil, err
	}

	var resultStatus = true
	var resultMessage string

	if storageRequest.Paths != nil {
		go func() {
			for _, path := range storageRequest.Paths {
				if path.Id != 0 && path.Path != "" {
					storage := &FileStorage{
						id:      path.Id,
						path:    path.Path,
						appAuth: appAuth,
					}

					rpcServer.AddJob(GetStorageSizeQueueJob(storage))
				}
			}
		}()
	} else {
		if storageRequest.Id == 0 {
			resultMessage = "Id is required"
			resultStatus = false
		} else if storageRequest.Path == "" {
			resultMessage = "Path is required"
			resultStatus = false
		} else {
			go func() {
				storage := &FileStorage{
					id:      storageRequest.Id,
					path:    storageRequest.Path,
					appAuth: appAuth,
				}

				rpcServer.AddJob(GetStorageSizeQueueJob(storage))
			}()
		}
	}

	requestAcceptedResponse := &rpc.ActionResponse{
		Status: resultStatus,
		Data:   resultMessage,
	}

	return requestAcceptedResponse, nil
})

func main() {
	flag.Parse()

	rpc.Config.SetServiceSettings(
		"dev01-fs-daemon",
		"Dev01 file storages monitor daemon",
		"The part of the Dev01 platform",
	)

	rpc.Config.SetAction("/get", &rpcAction)

	if rpc.Config.Serving() {
		serve()
	} else if storagePath != "" {
		debug.SetMaxThreads(1000000)

		storage := FileStorage{
			id:      storageId,
			path:    storagePath,
			appAuth: storageAuth,
		}
		storage.getSize()
	}
}

func serve() {
	rpcServer := &rpc.Rpc{}

	rpcServer.Run()
}
