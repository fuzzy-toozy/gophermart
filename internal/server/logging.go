package server

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type AppLogger struct {
	Logger  *zap.SugaredLogger
	LogFile *os.File
}

func LogTimeFormat(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	encodeTimeLayout(t, "2006/01/02 - 15:04:05", enc)
}

func encodeTimeLayout(t time.Time, layout string, enc zapcore.PrimitiveArrayEncoder) {
	type appendTimeEncoder interface {
		AppendTimeLayout(time.Time, string)
	}

	if enc, ok := enc.(appendTimeEncoder); ok {
		enc.AppendTimeLayout(t, layout)
		return
	}

	enc.AppendString(t.Format(layout))
}

type PrefixWriter struct {
	prefix   []byte
	writer   io.Writer
	wrBuffer bytes.Buffer // buffer to cache the slice to not allocate memory each call
}

func NewPrefixWriter(prefix string, w io.Writer) *PrefixWriter {
	return &PrefixWriter{
		prefix: []byte(prefix + " "),
		writer: w,
	}
}

func (w *PrefixWriter) Write(p []byte) (n int, err error) {

	w.wrBuffer.Reset()
	w.wrBuffer.Write(w.prefix)
	w.wrBuffer.Write(p)
	return w.writer.Write(w.wrBuffer.Bytes())
}

func getLogFile(fileName string) (*os.File, error) {
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file '%v': %w", fileName, err)
	}

	return f, nil
}

func LogInit(logLevel string, prefix string, fileName string) (AppLogger, error) {
	pe := zap.NewProductionEncoderConfig()

	pe.EncodeTime = LogTimeFormat
	pe.EncodeLevel = zapcore.CapitalColorLevelEncoder

	encoder := zapcore.NewConsoleEncoder(pe)

	level, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		return AppLogger{}, err
	}

	logFile, err := getLogFile(fileName)
	if err != nil {
		return AppLogger{}, err
	}

	fileWriter := NewPrefixWriter(prefix, logFile)
	consoleWriter := NewPrefixWriter(prefix, os.Stdout)

	core := zapcore.NewTee(
		// Add custom writers to zapcore
		zapcore.NewCore(encoder, zapcore.AddSync(fileWriter), level),
		zapcore.NewCore(encoder, zapcore.AddSync(consoleWriter), level),
	)

	l := zap.New(core)

	a := AppLogger{
		LogFile: logFile,
		Logger:  l.Sugar(),
	}

	return a, nil
}
