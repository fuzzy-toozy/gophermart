package main

import (
	"fmt"
	"os"

	"github.com/fuzzy-toozy/gophermart/internal/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	_ "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/lib/pq"
)

func logInit(d bool, logFile string) *zap.SugaredLogger {

	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)

	if err != nil {
		panic(err)
	}

	pe := zap.NewProductionEncoderConfig()

	fileEncoder := zapcore.NewJSONEncoder(pe)

	pe.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(pe)

	level := zap.InfoLevel
	if d {
		level = zap.DebugLevel
	}

	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, zapcore.AddSync(f), level),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level),
	)

	l := zap.New(core)

	return l.Sugar()
}

func main() {
	server, err := server.NewServer()
	if err != nil {
		fmt.Printf("Failed to create server: %v", err)
		return
	}

	server.Run()
}
