package main

import (
	"errors"
	"flag"
	"log"
	"os"
)

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func main() {
	var serviceFlag = false
	var installFlag = false
	var uninstallFlag = false
	flag.BoolVar(&serviceFlag, "srv", false, "Start as Windows service")
	flag.BoolVar(&installFlag, "srv.install", false, "Install Windows service")
	flag.BoolVar(&uninstallFlag, "srv.uninstall", false, "Uninstall Windows service")

	addr := flag.String("http.addr", ":8090", "Listening network address")
	appUrl := flag.String("app.url", "", "Application hook URL")
	secret := flag.String("app.secret", "", "Application secret")

	flag.Parse()

	config := &Config{
		addr:      *addr,
		appUrl:    *appUrl,
		appSecret: *secret,
	}

	if isFlagPassed("srv.install") {
		srv := getService(config)
		err := srv.Install()

		if err != nil {
			log.Fatalf("Cannot install service: %v\n", err)
		}

		log.Println("Service installed")
		os.Exit(0)
	}

	if isFlagPassed("srv.uninstall") {
		srv := getService(config)
		err := srv.Uninstall()

		if err != nil {
			log.Fatalf("Cannot uninstall service: %v\n", err)
		}

		log.Println("Service uninstalled")
		os.Exit(0)
	}

	if *appUrl == "" {
		panic(errors.New("application hook URL is required"))
	}

	if isFlagPassed("srv") {
		srv := getService(config)
		err := srv.Run()

		if err != nil {
			log.Fatalf("Cannot start the service: %v\n", err)
		}
	} else {
		folder := &Folder{
			config: config,
		}
		folder.serve(*addr)
	}
}
