package logger

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"sync/atomic"
	"unsafe"

	"go.uber.org/zap"
)

type sink struct {
	path     string
	notifier chan os.Signal
	file     unsafe.Pointer
}

func NewLogrotateSink(u *url.URL, sig ...os.Signal) (zap.Sink, error) {
	notifier := make(chan os.Signal, 1)
	signal.Notify(notifier, sig...)

	if u.User != nil {
		return nil, fmt.Errorf("user and password not allowed with logrotate file URLs: got %v", u)
	}
	if u.Fragment != "" {
		return nil, fmt.Errorf("fragments not allowed with logrotate file URLs: got %v", u)
	}
	if u.Port() != "" {
		return nil, fmt.Errorf("ports not allowed with logrotate file URLs: got %v", u)
	}
	if hn := u.Hostname(); hn != "" && hn != "localhost" {
		return nil, fmt.Errorf("logrotate file URLs must leave host empty or use localhost: got %v", u)
	}

	sink := &sink{
		path:     u.Path,
		notifier: notifier,
	}
	if err := sink.reopen(); err != nil {
		return nil, err
	}
	go sink.listenToSignal()
	return sink, nil
}

func (m *sink) listenToSignal() {
	for {
		_, ok := <-m.notifier
		if !ok {
			return
		}
		if err := m.reopen(); err != nil {
			fallbackLogger.Errorf("%s", err)
		}
	}
}

func (m *sink) reopen() error {
	file, err := os.OpenFile(m.path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file on %s: %w", m.path, err)
	}
	old := (*os.File)(m.file)
	atomic.StorePointer(&m.file, unsafe.Pointer(file))
	if old != nil {
		if err := old.Close(); err != nil {
			return fmt.Errorf("failed to close old file: %w", err)
		}
	}
	return nil
}

func (m *sink) getFile() *os.File {
	return (*os.File)(atomic.LoadPointer(&m.file))
}

func (m *sink) Close() error {
	signal.Stop(m.notifier)
	close(m.notifier)
	return m.getFile().Close()
}

func (m *sink) Write(p []byte) (n int, err error) {
	return m.getFile().Write(p)
}

func (m *sink) Sync() error {
	return m.getFile().Sync()
}
