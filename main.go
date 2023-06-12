package main

import (
	"flag"
	"log"
	"os"
	"strings"
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

func isService() bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "srv" || strings.HasPrefix(f.Name, "srv.") {
			found = true
		}
	})
	return found
}

func main() {
	var serviceFlag = false
	var installFlag = false
	var uninstallFlag = false
	var helpFlag = false
	flag.BoolVar(&serviceFlag, "srv", false, "Start as Windows service")
	flag.BoolVar(&installFlag, "srv.install", false, "Install Windows service")
	flag.BoolVar(&uninstallFlag, "srv.uninstall", false, "Uninstall Windows service")
	flag.BoolVar(&helpFlag, "help", false, "Usage help")

	addr := flag.String("http.addr", ":8090", "Listening network address")
	appUrl := flag.String("app.url", "", "Application hook URL")
	secret := flag.String("app.secret", "", "Application secret")

	flag.Parse()

	if isFlagPassed("help") {
		flag.PrintDefaults()
		os.Exit(0)
	}

	config := &Config{
		addr:      *addr,
		appUrl:    *appUrl,
		appSecret: *secret,
	}

	if isService() {
		runService(config)
	} else {
		if *appUrl == "" {
			flag.PrintDefaults()
			log.Fatal("application hook URL is required")
		}

		folder := &FileStorage{
			config: config,
		}
		folder.serve(*addr)
	}
}
