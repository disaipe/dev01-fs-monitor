package main

import (
	"github.com/kardianos/service"
	"log"
)

const serviceName = "dev01-fs-daemon"
const serviceDescription = "dev01 filesystem watcher daemon"

type WindowsService struct {
	config *Config
}

func (p WindowsService) Start(s service.Service) error {
	folder := &Folder{
		config: p.config,
	}

	go func() {
		folder.serve(p.config.addr)
	}()

	return nil
}

func (p WindowsService) Stop(s service.Service) error {
	return nil
}

func getService(config *Config) service.Service {
	serviceConfig := &service.Config{
		Name:        serviceName,
		DisplayName: serviceName,
		Description: serviceDescription,
	}

	prg := &WindowsService{
		config: config,
	}

	srv, err := service.New(prg, serviceConfig)

	if err != nil {
		log.Fatalf("Cannot create the service: %v\n", err)
	}

	return srv
}
