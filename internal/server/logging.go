package server

import (
	"bytes"
	"io"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	LogLevelDebug   = "debug"
	LogLevelRelease = "release"
)

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

func LogInit(d bool, prefix string, f *os.File) *zap.SugaredLogger {
	pe := zap.NewProductionEncoderConfig()

	// Create our custom writers
	fileWriter := NewPrefixWriter(prefix, f)
	consoleWriter := NewPrefixWriter(prefix, os.Stdout)

	pe.EncodeTime = LogTimeFormat
	pe.EncodeLevel = zapcore.CapitalColorLevelEncoder

	encoder := zapcore.NewConsoleEncoder(pe)

	level := zap.InfoLevel
	if d {
		level = zap.DebugLevel
	}

	core := zapcore.NewTee(
		// Add custom writers to zapcore
		zapcore.NewCore(encoder, zapcore.AddSync(fileWriter), level),
		zapcore.NewCore(encoder, zapcore.AddSync(consoleWriter), level),
	)

	l := zap.New(core)

	return l.Sugar()
}
