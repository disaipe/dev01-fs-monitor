package main

import (
	"flag"
	"log"
	"os"
	"runtime/debug"
	"strings"
)

type Config struct {
	addr      string
	appUrl    string
	appSecret string
}

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

var AppConfig *Config

func main() {
	var serviceFlag = false
	var installFlag = false
	var uninstallFlag = false
	var helpFlag = false
	var serve = false
	flag.BoolVar(&serve, "serve", false, "Start HTTP server")
	flag.BoolVar(&serviceFlag, "srv", false, "Start as Windows service")
	flag.BoolVar(&installFlag, "srv.install", false, "Install Windows service")
	flag.BoolVar(&uninstallFlag, "srv.uninstall", false, "Uninstall Windows service")
	flag.BoolVar(&helpFlag, "help", false, "Usage help")

	addr := flag.String("http.addr", ":8090", "Listening network address")
	appUrl := flag.String("app.url", "", "Application hook URL")
	secret := flag.String("app.secret", "", "Application secret")

	storageId := flag.Int("id", 0, "Storage id")
	storagePath := flag.String("path", "", "Path to get size")
	storageAuth := flag.String("auth", "", "Storage result request auth key")

	flag.Parse()

	if isFlagPassed("help") {
		flag.PrintDefaults()
		os.Exit(0)
	}

	AppConfig = &Config{
		addr:      *addr,
		appUrl:    *appUrl,
		appSecret: *secret,
	}

	if isService() {
		if *appUrl == "" {
			flag.PrintDefaults()
			log.Fatal("application hook URL is required")
		}

		runService()
	} else if serve {
		if *appUrl == "" {
			flag.PrintDefaults()
			log.Fatal("application hook URL is required")
		}

		rpc := &Rpc{}
		rpc.serve(*addr)
	} else if *storagePath != "" {
		debug.SetMaxThreads(1000000)

		storage := FileStorage{
			id:      *storageId,
			path:    *storagePath,
			appAuth: *storageAuth,
		}
		storage.getSize()
	}
}
