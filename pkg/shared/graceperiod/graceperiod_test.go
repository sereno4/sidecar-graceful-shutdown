package graceperiod

import (
"testing"
"time"
)

func TestManager_BeforeSIGTERM(t *testing.T) {
m := New(120, 5*time.Second)
remaining := m.RemainingTime()
if remaining < 114*time.Second || remaining > 116*time.Second {
t.Errorf("remaining incorreto: %v", remaining)
}
}

func TestManager_AfterSIGTERM(t *testing.T) {
m := New(120, 5*time.Second)
m.OnSIGTERM()
time.Sleep(100 * time.Millisecond)

remaining := m.RemainingTime()
if remaining > 115*time.Second || remaining < 110*time.Second {
t.Errorf("remaining após SIGTERM incorreto: %v", remaining)
}
}

func TestManager_Expired(t *testing.T) {
m := New(2, 5*time.Second)
m.OnSIGTERM()
time.Sleep(100 * time.Millisecond)

if !m.IsExpired() {
t.Errorf("deveria estar expirado")
}
}

func TestManager_FlushContext(t *testing.T) {
m := New(120, 5*time.Second)
m.OnSIGTERM()

ctx, cancel := m.FlushContext()
defer cancel()

deadline, ok := ctx.Deadline()
if !ok {
t.Fatal("contexto deveria ter deadline")
}
if time.Until(deadline) <= 0 {
t.Errorf("deadline no passado: %v", deadline)
}
}

func TestManager_FlushContext_Expired(t *testing.T) {
m := New(2, 5*time.Second)
m.OnSIGTERM()

ctx, cancel := m.FlushContext()
defer cancel()

select {
case <-ctx.Done():
default:
t.Error("contexto deveria estar cancelado")
}
}
