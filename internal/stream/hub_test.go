package stream

import (
	"sync"
	"testing"
	"time"

	"github.com/divijg19/Pulse/internal/model"
)

func TestHub_AddRemove(t *testing.T) {
	hub := NewHub()
	ch := make(chan model.Event, 10)

	hub.Add(ch)
	hub.Broadcast(model.Event{Type: "result", Data: "hello"})

	select {
	case event := <-ch:
		if event.Type != "result" {
			t.Fatalf("event type = %q", event.Type)
		}
		if event.Data != "hello" {
			t.Fatalf("event data = %v", event.Data)
		}
	case <-time.After(time.Second):
		t.Fatal("did not receive broadcast")
	}

	hub.Remove(ch)

	_, ok := <-ch
	if ok {
		t.Fatal("channel should be closed after remove")
	}
}

func TestHub_RemoveIdempotent(t *testing.T) {
	hub := NewHub()
	ch := make(chan model.Event, 10)

	hub.Add(ch)
	hub.Remove(ch)
	hub.Remove(ch)
}

func TestHub_Broadcast_Multiple(t *testing.T) {
	hub := NewHub()

	ch1 := make(chan model.Event, 100)
	ch2 := make(chan model.Event, 100)
	ch3 := make(chan model.Event, 100)

	hub.Add(ch1)
	hub.Add(ch2)
	hub.Add(ch3)

	for i := range 10 {
		hub.Broadcast(model.Event{Type: "result", Data: i})
	}

	hub.Remove(ch1)
	hub.Remove(ch2)
	hub.Remove(ch3)

	for _, ch := range []<-chan model.Event{ch1, ch2, ch3} {
		count := 0
		for event := range ch {
			count++
			if event.Type != "result" {
				t.Fatalf("event type = %q", event.Type)
			}
		}
		if count != 10 {
			t.Fatalf("received %d events (expected 10)", count)
		}
	}
}

func TestHub_SlowConsumer(t *testing.T) {
	hub := NewHub()

	fast := make(chan model.Event, 100)
	slow := make(chan model.Event, 1)

	hub.Add(fast)
	hub.Add(slow)

	for i := range 20 {
		hub.Broadcast(model.Event{Type: "result", Data: i})
	}

	hub.Remove(fast)
	hub.Remove(slow)

	fastCount := 0
	for range fast {
		fastCount++
	}

	slowCount := 0
	for range slow {
		slowCount++
	}

	if fastCount != 20 {
		t.Fatalf("fast consumer got %d (expected 20)", fastCount)
	}
	if slowCount == 20 {
		t.Fatalf("slow consumer got all 20 — non-blocking broadcast may not be working")
	}
	if slowCount == 0 {
		t.Fatalf("slow consumer got 0 — should have received at least 1 before buffer filled")
	}
}

func TestHub_ConcurrentAccess(t *testing.T) {
	hub := NewHub()

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch := make(chan model.Event, 10)
			hub.Add(ch)
			for range 5 {
				hub.Broadcast(model.Event{Type: "result"})
			}
			hub.Remove(ch)
		}()
	}

	wg.Wait()
}

func TestHub_BroadcastEmpty(t *testing.T) {
	hub := NewHub()
	hub.Broadcast(model.Event{Type: "result", Data: "no clients"})
}

func TestHub_ClientOrder(t *testing.T) {
	hub := NewHub()
	ch := make(chan model.Event, 10)
	hub.Add(ch)

	for i := range 5 {
		hub.Broadcast(model.Event{Type: "result", Data: i})
	}

	hub.Remove(ch)

	expected := 0
	for event := range ch {
		val, ok := event.Data.(int)
		if !ok {
			t.Fatalf("data is not int: %T", event.Data)
		}
		if val != expected {
			t.Fatalf("received %d (expected %d)", val, expected)
		}
		expected++
	}
}
