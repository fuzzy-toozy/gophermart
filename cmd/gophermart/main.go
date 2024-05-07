package main

import (
	"log"

	"github.com/fuzzy-toozy/gophermart/internal/config"
	"github.com/fuzzy-toozy/gophermart/internal/server"
)

func main() {
	c, err := config.BuildConfig()
	if err != nil {
		log.Fatalf("Failed to build app config: %v", err)
	}

	appLog, err := server.LogInit(c.LogLevel, c.LogPrefix, c.LogFile)
	if err != nil {
		log.Fatalf("Failed to create app logger: %v", err)
	}

	defer appLog.Logger.Sync()

	server, err := server.NewServer(c, appLog)
	if err != nil {
		appLog.Logger.Errorf("Failed to create server: %v", err)
		return
	}

	server.Run()
}
