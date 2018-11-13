package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Toggly/core/internal/pkg/api"
	"github.com/Toggly/core/internal/server"
	"github.com/Toggly/core/internal/server/rest"

	"github.com/Toggly/core/internal/pkg/cache"
	"github.com/Toggly/core/internal/pkg/storage"
	"github.com/Toggly/core/internal/pkg/storage/mongo"
	flags "github.com/jessevdk/go-flags"
)

var revision = "development" //revision assigned on build

// Opts describes application command line arguments
type Opts struct {
	Toggly struct {
		Port     int    `short:"p" long:"port" env:"API_PORT" default:"8080" description:"port"`
		BasePath string `long:"base-path" env:"API_BASE_PATH" default:"/api" description:"Base API Path"`
		Debug    bool   `long:"debug" description:"Run in DEBUG mode"`
		Store    struct {
			Mongo struct {
				URL string `long:"url" env:"URL" description:"mongo connection url"`
			} `group:"mongo" namespace:"mongo" env-namespace:"MONGO"`
		} `group:"store" namespace:"store" env-namespace:"STORE"`
		// Cache struct {
		// 	Disabled bool `long:"disable" description:"Disable cache" env:"DISABLE"`
		// 	Redis    struct {
		// 		URL string `long:"url" env:"URL" description:"redis connection url"`
		// 	} `group:"redis" namespace:"redis" env-namespace:"REDIS"`
		// } `group:"cache" namespace:"cache" env-namespace:"CACHE"`
	} `group:"toggly" env-namespace:"TOGGLY"`
}

func main() {

	fmt.Println(`
::::::::::: ::::::::   ::::::::   ::::::::  :::     :::   ::: 
    :+:    :+:    :+: :+:    :+: :+:    :+: :+:     :+:   :+: 
    +:+    +:+    +:+ +:+        +:+        +:+      +:+ +:+  
    +#+    +#+    +:+ :#:        :#:        +#+       +#++:   
    +#+    +#+    +#+ +#+   +#+# +#+   +#+# +#+        +#+    
    #+#    #+#    #+# #+#    #+# #+#    #+# #+#        #+#    
    ###     ########   ########   ########  ########## ###    
	`)
	fmt.Println(centeredText("-= Core API Server =-", 63))
	fmt.Println(centeredText(fmt.Sprintf("ver: %s", revision), 63))
	fmt.Print("--------------------------------------------------------------\n\n")

	var apiCache cache.DataCache
	var dataStorage storage.DataStorage
	var err error

	var opts Opts
	if _, err = flags.NewParser(&opts, flags.Default).ParseArgs(os.Args[1:]); err != nil {
		os.Exit(0)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() { // catch signal and invoke graceful termination
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Print("[WARN] interrupt signal")
		cancel()
	}()

	// if opts.Toggly.Cache.Disabled {
	// 	log.Print("[WARN] CACHE DISABLED")
	// } else {
	// 	if apiCache, err = cache.NewHashMapCache(); err != nil {
	// 		log.Fatalf("Can't connect to cache service: %v", err)
	// 	}
	// }

	if dataStorage, err = mongo.NewMongoStorage(opts.Toggly.Store.Mongo.URL); err != nil {
		log.Fatalf("Can't connect to storage: %v", err)
	}

	server := &server.Application{
		Router: &rest.APIRouter{
			Version:  revision,
			Cache:    apiCache,
			Engine:   &api.Engine{Storage: &dataStorage},
			BasePath: opts.Toggly.BasePath,
			Port:     opts.Toggly.Port,
			IsDebug:  opts.Toggly.Debug,
		},
	}

	if err != nil {
		log.Fatalf("[ERROR] failed to setup application, %+v", err)
	}

	log.Print("[INFO] API server started")

	server.Run(ctx)
	log.Print("[INFO] application terminated")
	log.Print("[INFO] Bye!")
}

func centeredText(txt string, width int) string {
	return strings.Repeat(" ", (width-len(txt))/2) + txt
}
