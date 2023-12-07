package database

import (
	"time"

	"go.uber.org/zap"
)

type DBConfig struct {
	ConnURI    string
	DriverName string
	Timeout    time.Duration
}

func (c *DBConfig) Print(logger *zap.SugaredLogger) {
	logger.Infof("Conn string: %v", c.ConnURI)
	logger.Infof("Driver name: %v", c.DriverName)
	logger.Infof("Ping timeout: %v", c.Timeout)
}
