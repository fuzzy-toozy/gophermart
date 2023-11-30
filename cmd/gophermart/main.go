package main

import (
	"github.com/fuzzy-toozy/gophermart/internal/server"
	"go.uber.org/zap"

	_ "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/lib/pq"
)

func main() {
	logger := zap.Must(zap.NewDevelopment()).Sugar()
	server, err := server.NewServer(logger)

	if err != nil {
		logger.Fatalf("Failed to create server: %v", err)
		return
	}

	err = server.Run()
	if err != nil {
		logger.Infof("Server exit reason: %v", err)
	}
}
