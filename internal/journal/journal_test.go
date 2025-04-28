package journal

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/noosxe/dotman/internal/fs"
)

func TestJournalManager(t *testing.T) {
	// Create a mock filesystem
	mockFS := fs.NewMockFileSystem(nil)
	journalDir := "test/journal"

	// Create journal manager
	jm := NewJournalManager(mockFS, journalDir)

	// Test initialization
	if err := jm.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Test creating an entry
	entry, err := jm.CreateEntry("add", "source/file", "target/file")
	if err != nil {
		t.Fatalf("CreateEntry failed: %v", err)
	}

	// Verify entry fields
	if entry.Operation != "add" {
		t.Errorf("Expected operation 'add', got '%s'", entry.Operation)
	}
	if entry.Source != "source/file" {
		t.Errorf("Expected source 'source/file', got '%s'", entry.Source)
	}
	if entry.Target != "target/file" {
		t.Errorf("Expected target 'target/file', got '%s'", entry.Target)
	}
	if entry.State != "current" {
		t.Errorf("Expected state 'current', got '%s'", entry.State)
	}

	// Test updating an entry
	entry.Steps = append(entry.Steps, Step{
		Type:   "copy",
		Status: "completed",
	})
	if err := jm.UpdateEntry(entry); err != nil {
		t.Fatalf("UpdateEntry failed: %v", err)
	}

	// Test moving an entry
	if err := jm.MoveEntry(entry, "completed"); err != nil {
		t.Fatalf("MoveEntry failed: %v", err)
	}

	// Test retrieving the entry
	retrieved, err := jm.GetEntry(entry.ID)
	if err != nil {
		t.Fatalf("GetEntry failed: %v", err)
	}

	// Verify retrieved entry
	if retrieved.ID != entry.ID {
		t.Errorf("Expected ID '%s', got '%s'", entry.ID, retrieved.ID)
	}
	if retrieved.State != "completed" {
		t.Errorf("Expected state 'completed', got '%s'", retrieved.State)
	}
	if len(retrieved.Steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(retrieved.Steps))
	}

	// Test listing entries
	entries, err := jm.ListEntries("completed")
	if err != nil {
		t.Fatalf("ListEntries failed: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry in completed state, got %d", len(entries))
	}
}

func TestJournalEntrySerialization(t *testing.T) {
	// Create a test entry
	entry := &JournalEntry{
		ID:        "test-123",
		Timestamp: time.Now(),
		Operation: "add",
		Source:    "source/file",
		Target:    "target/file",
		State:     "pending",
		Steps: []Step{
			{
				Type:   "copy",
				Status: "completed",
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal back
	var unmarshaled JournalEntry
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify fields
	if unmarshaled.ID != entry.ID {
		t.Errorf("Expected ID '%s', got '%s'", entry.ID, unmarshaled.ID)
	}
	if unmarshaled.Operation != entry.Operation {
		t.Errorf("Expected operation '%s', got '%s'", entry.Operation, unmarshaled.Operation)
	}
	if unmarshaled.Source != entry.Source {
		t.Errorf("Expected source '%s', got '%s'", entry.Source, unmarshaled.Source)
	}
	if unmarshaled.Target != entry.Target {
		t.Errorf("Expected target '%s', got '%s'", entry.Target, unmarshaled.Target)
	}
	if unmarshaled.State != entry.State {
		t.Errorf("Expected state '%s', got '%s'", entry.State, unmarshaled.State)
	}
	if len(unmarshaled.Steps) != len(entry.Steps) {
		t.Errorf("Expected %d steps, got %d", len(entry.Steps), len(unmarshaled.Steps))
	}
}
