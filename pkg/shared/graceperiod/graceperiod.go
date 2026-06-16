package graceperiod

import (
"context"
"fmt"
"time"
)

type Manager struct {
gracePeriod  time.Duration
safetyMargin time.Duration
sigtermTime  time.Time
}

func New(gracePeriodSec int, safetyMargin time.Duration) *Manager {
return &Manager{
gracePeriod:  time.Duration(gracePeriodSec) * time.Second,
safetyMargin: safetyMargin,
}
}

func (m *Manager) OnSIGTERM() {
m.sigtermTime = time.Now()
}

func (m *Manager) RemainingTime() time.Duration {
if m.sigtermTime.IsZero() {
return m.gracePeriod - m.safetyMargin
}

elapsed := time.Since(m.sigtermTime)
remaining := m.gracePeriod - elapsed - m.safetyMargin

if remaining <= 0 {
return 0
}
return remaining
}

func (m *Manager) IsExpired() bool {
return m.RemainingTime() <= 0
}

func (m *Manager) FlushContext() (context.Context, context.CancelFunc) {
remaining := m.RemainingTime()

if remaining <= 2*time.Second {
ctx, cancel := context.WithCancel(context.Background())
cancel()
return ctx, cancel
}

deadline := remaining - 1*time.Second
return context.WithTimeout(context.Background(), deadline)
}

func (m *Manager) String() string {
return fmt.Sprintf("gracePeriod=%v safetyMargin=%v remaining=%v",
m.gracePeriod, m.safetyMargin, m.RemainingTime())
}
