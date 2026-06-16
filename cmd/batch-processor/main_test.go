package main

import (
	"testing"
)

func TestProcessBatch(t *testing.T) {

	done := make(chan bool)

	go func() {

		for i := 0; i < 5; i++ {
			done <- true
			return
		}

	}()

	select {

	case <-done:
		t.Log("batch finalizado")

	}

}
