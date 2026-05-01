package logger

import (
	"fmt"
	"net/url"
	"os"
	"syscall"
	"time"

	"github.com/mattn/go-isatty"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	fileTransport    = "file"
	elasticTransport = "elastic"
)

type transport struct {
	core zapcore.Core
	stop func()
}

func getStdoutTransport(info *loggerInfo) *transport {
	res := &transport{}
	sink := zapcore.BufferedWriteSyncer{
		WS:            zapcore.AddSync(os.Stdout),
		Size:          1024 * 1024,
		FlushInterval: 1 * time.Second,
	}

	var encoder zapcore.Encoder
	if info.cfg.Format != "json" {
		if isatty.IsTerminal(os.Stdout.Fd()) {
			info.encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
		encoder = zapcore.NewConsoleEncoder(info.encCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(info.encCfg)
	}

	res.core = zapcore.NewCore(encoder, &sink, info.lvl)
	res.stop = func() {
		if err := sink.Stop(); err != nil {
			fallbackLogger.Warn("Failed to sync sink", zap.Error(err))
		}
	}

	return res
}

func getFileTransport(info *loggerInfo) (*transport, error) {
	res := &transport{}

	if info.cfg.FilePath == "" {
		return nil, fmt.Errorf("no file path specified")
	}

	u := &url.URL{
		Path: info.cfg.FilePath,
	}

	sink, err := NewLogrotateSink(u, syscall.SIGUSR1)
	if err != nil {
		return nil, fmt.Errorf("open logrotate sink: %w", err)
	}

	res.core = zapcore.NewCore(zapcore.NewJSONEncoder(info.encCfg), sink, info.lvl)
	res.stop = func() {
		if err := sink.Close(); err != nil {
			fallbackLogger.Error("Failed to close sink", zap.Error(err))
		}
	}

	return res, nil
}

func getElasticTransport(info *loggerInfo) (*transport, error) {
	res := &transport{}

	if info.cfg.ElasticConfig == nil {
		return nil, fmt.Errorf("no elastic config specified")
	}

	sink, err := NewElasticSink(fallbackLogger,
		WithFlushInterval(info.cfg.ElasticConfig.FlushInterval),
		WithIndex(info.cfg.ElasticConfig.Index),
		WithUrl(info.cfg.ElasticConfig.Url),
		WithWriteBufferSize(info.cfg.ElasticConfig.WriteBufferSize),
	)
	if err != nil {
		return nil, fmt.Errorf("open elastic sink: %w", err)
	}

	res.core = zapcore.NewCore(zapcore.NewJSONEncoder(info.encCfg), sink, info.lvl)
	res.stop = func() {
		if err := sink.Close(); err != nil {
			fallbackLogger.Error("Failed to close sink", zap.Error(err))
		}
	}
	return res, nil
}
