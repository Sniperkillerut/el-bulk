package cache

import (
	"testing"
	"time"
)

func TestTTLMap(t *testing.T) {
	tm := NewTTLMap[string](100 * time.Millisecond)

	// Test Set and Get
	tm.Set("key1", "value1", 200*time.Millisecond)
	val, ok := tm.Get("key1")
	if !ok || val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}

	// Test Expiration
	time.Sleep(300 * time.Millisecond)
	_, ok = tm.Get("key1")
	if ok {
		t.Error("Expected key1 to be expired")
	}

	// Test Delete
	tm.Set("key2", "value2", 1*time.Minute)
	tm.Delete("key2")
	_, ok = tm.Get("key2")
	if ok {
		t.Error("Expected key2 to be deleted")
	}

	// Test Clear
	tm.Set("key3", "value3", 1*time.Minute)
	tm.Set("key4", "value4", 1*time.Minute)
	tm.Clear()
	_, ok = tm.Get("key3")
	if ok {
		t.Error("Expected key3 to be cleared")
	}
	_, ok = tm.Get("key4")
	if ok {
		t.Error("Expected key4 to be cleared")
	}
}
