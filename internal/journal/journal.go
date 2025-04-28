package journal

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	dotmanfs "github.com/noosxe/dotman/internal/fs"
)

// JournalEntry represents a single journal entry
type JournalEntry struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Operation string    `json:"operation"`
	Source    string    `json:"source,omitempty"`
	Target    string    `json:"target,omitempty"`
	State     string    `json:"state"`
	Checksum  string    `json:"checksum,omitempty"`
	Steps     []Step    `json:"steps"`
}

// Step represents a single step in a journal entry
type Step struct {
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	Error       string    `json:"error,omitempty"`
	Description string    `json:"description,omitempty"`
	Source      string    `json:"source,omitempty"`
	Target      string    `json:"target,omitempty"`
	Details     string    `json:"details,omitempty"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time,omitempty"`
}

// JournalManager manages journal entries
type JournalManager struct {
	fsys       dotmanfs.FileSystem
	journalDir string
}

// NewJournalManager creates a new JournalManager
func NewJournalManager(fsys dotmanfs.FileSystem, journalDir string) *JournalManager {
	return &JournalManager{
		fsys:       fsys,
		journalDir: journalDir,
	}
}

// Initialize creates the journal directory structure
func (jm *JournalManager) Initialize() error {
	// Create main journal directory
	if err := jm.fsys.MkdirAll(jm.journalDir, 0755); err != nil {
		return fmt.Errorf("error creating journal directory: %v", err)
	}

	// Create subdirectories
	subdirs := []string{"current", "completed", "failed"}
	for _, dir := range subdirs {
		path := filepath.Join(jm.journalDir, dir)
		if err := jm.fsys.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("error creating %s directory: %v", dir, err)
		}
	}

	return nil
}

// CreateEntry creates a new journal entry
func (jm *JournalManager) CreateEntry(operation, source, target string) (*JournalEntry, error) {
	entry := &JournalEntry{
		ID:        generateOperationID(operation),
		Timestamp: time.Now(),
		Operation: operation,
		Source:    source,
		Target:    target,
		State:     "current",
		Steps:     make([]Step, 0),
	}

	// Save the entry
	if err := jm.saveEntry(entry); err != nil {
		return nil, err
	}

	return entry, nil
}

// UpdateEntry updates an existing journal entry
func (jm *JournalManager) UpdateEntry(entry *JournalEntry) error {
	return jm.saveEntry(entry)
}

// MoveEntry moves a journal entry to a different state directory
func (jm *JournalManager) MoveEntry(entry *JournalEntry, newState string) error {
	oldPath := filepath.Join(jm.journalDir, entry.State, entry.ID+".json")

	// Update the state
	entry.State = newState

	// Write to new location with updated state
	if err := jm.saveEntry(entry); err != nil {
		return fmt.Errorf("error writing entry: %v", err)
	}

	// Remove old file
	if err := jm.fsys.Remove(oldPath); err != nil {
		return fmt.Errorf("error removing old entry: %v", err)
	}

	return nil
}

// GetEntry retrieves a journal entry by ID
func (jm *JournalManager) GetEntry(id string) (*JournalEntry, error) {
	// Try to find the entry in any state directory
	states := []string{"current", "completed", "failed"}
	for _, state := range states {
		path := filepath.Join(jm.journalDir, state, id+".json")
		if _, err := jm.fsys.Stat(path); err == nil {
			return jm.readEntry(path)
		}
	}
	return nil, fmt.Errorf("entry not found: %s", id)
}

// ListEntries lists all journal entries in a given state
func (jm *JournalManager) ListEntries(state string) ([]*JournalEntry, error) {
	dir := filepath.Join(jm.journalDir, state)
	entries := make([]*JournalEntry, 0)

	// Read directory
	dirFile, err := jm.fsys.Open(dir)
	if err != nil {
		return nil, fmt.Errorf("error opening directory: %v", err)
	}
	defer dirFile.Close()

	// Read all entries
	dirEntries, err := dirFile.(fs.ReadDirFile).ReadDir(-1)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %v", err)
	}

	for _, entry := range dirEntries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			path := filepath.Join(dir, entry.Name())
			journalEntry, err := jm.readEntry(path)
			if err != nil {
				return nil, fmt.Errorf("error reading entry %s: %v", entry.Name(), err)
			}
			entries = append(entries, journalEntry)
		}
	}

	return entries, nil
}

// Helper functions

func (jm *JournalManager) saveEntry(entry *JournalEntry) error {
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling entry: %v", err)
	}

	path := filepath.Join(jm.journalDir, entry.State, entry.ID+".json")
	return jm.fsys.WriteFile(path, data, 0644)
}

func (jm *JournalManager) readEntry(path string) (*JournalEntry, error) {
	data, err := jm.fsys.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	var entry JournalEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("error unmarshaling entry: %v", err)
	}

	return &entry, nil
}

func generateOperationID(operation string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s-%d", operation, timestamp)
}
