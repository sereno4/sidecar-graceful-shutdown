package main

import (
	"context"
	"testing"
	"time"

	"sidecar-graceful-shutdown/pkg/config"
	"sidecar-graceful-shutdown/pkg/shared/graceperiod"
)

func TestSidecar_BatchDone(t *testing.T) {

	cfg := &config.Config{
		GracePeriodSeconds: 120,
	}

	s := &Sidecar{
		cfg: cfg,
		graceMgr: graceperiod.New(
			120,
			5*time.Second,
		),
	}

	if s.isBatchDone() {
		t.Error("batch deveria estar false inicialmente")
	}

	s.setBatchDone(true)

	if !s.isBatchDone() {
		t.Error("batch deveria estar true após set")
	}
}

func TestSidecar_FlushBuffer(t *testing.T) {

	cfg := &config.Config{
		BufferSize: 100,
	}

	s := &Sidecar{
		cfg:    cfg,
		buffer: make([]Metric, 0, 100),
	}

	s.bufferMu.Lock()

	s.buffer = append(
		s.buffer,
		Metric{
			Name:  "test",
			Value: 1.0,
		},
	)

	s.buffer = append(
		s.buffer,
		Metric{
			Name:  "test",
			Value: 2.0,
		},
	)

	s.bufferMu.Unlock()

	ctx := context.Background()

	err := s.flushBuffer(ctx)

	if err != nil {
		t.Errorf(
			"flush falhou: %v",
			err,
		)
	}

	s.bufferMu.Lock()
	size := len(s.buffer)
	s.bufferMu.Unlock()

	if size != 0 {
		t.Errorf(
			"buffer deveria estar vazio, tem %d",
			size,
		)
	}
}
