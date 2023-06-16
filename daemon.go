package main

import (
	"bufio"
	"fmt"
	"github.com/kardianos/service"
	"log"
	"os"
	"strings"
)

type Daemon struct{}

func (p Daemon) Start(s service.Service) error {
	rpc := &Rpc{}

	go func() {
		rpc.serve(AppConfig.addr)
	}()

	return nil
}

func (p Daemon) Stop(s service.Service) error {
	return nil
}

func runService() {
	if isFlagPassed("srv.install") {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Application hook URL (required): ")
		appUrl, _ := reader.ReadString('\n')
		appUrl = strings.Replace(appUrl, "\n", "", -1)

		fmt.Print("Application secret: ")
		appSecret, _ := reader.ReadString('\n')
		appSecret = strings.Replace(appSecret, "\n", "", -1)

		srv := getService([]string{
			"-srv",
			fmt.Sprintf("-app.url=%v", appUrl),
			fmt.Sprintf("-app.secret=%v", appSecret),
		})

		err := srv.Install()

		if err != nil {
			log.Fatalf("Cannot install service: %v\n", err)
		}

		log.Println("Service installed")
		os.Exit(0)
	}

	if isFlagPassed("srv.uninstall") {
		srv := getService([]string{})
		err := srv.Uninstall()

		if err != nil {
			log.Fatalf("Cannot uninstall service: %v\n", err)
		}

		log.Println("Service uninstalled")
		os.Exit(0)
	}

	srv := getService([]string{})
	err := srv.Run()

	if err != nil {
		log.Fatalf("Cannot start the service: %v\n", err)
	}

	os.Exit(0)
}

func getService(args []string) service.Service {
	serviceConfig := &service.Config{
		Name:        "dev01-fs-daemon",
		DisplayName: "Dev01 file storages monitor daemon",
		Description: "The part of the Dev01 platform",
		Arguments:   args,
	}

	prg := &Daemon{}

	srv, err := service.New(prg, serviceConfig)

	if err != nil {
		log.Fatalf("Cannot create the service: %v\n", err)
	}

	return srv
}
