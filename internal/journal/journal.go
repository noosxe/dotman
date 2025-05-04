package journal

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	dotmanfs "github.com/noosxe/dotman/internal/fs"
)

// StepStatus represents the possible states of a step
type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
)

// StepType represents the possible types of steps
type StepType string

const (
	StepTypeVerify  StepType = "verify"
	StepTypeCopy    StepType = "copy"
	StepTypeMove    StepType = "move"
	StepTypeSymlink StepType = "symlink"
	StepTypeGit     StepType = "git"
)

// OperationType represents the possible types of operations
type OperationType string

const (
	OperationTypeAdd    OperationType = "add"
	OperationTypeRemove OperationType = "remove"
	OperationTypeLink   OperationType = "link"
)

// EntryState represents the possible states of a journal entry
type EntryState string

const (
	EntryStateCurrent   EntryState = "current"
	EntryStateCompleted EntryState = "completed"
	EntryStateFailed    EntryState = "failed"
)

// JournalEntry represents a single journal entry
type JournalEntry struct {
	ID        string        `json:"id"`
	Timestamp time.Time     `json:"timestamp"`
	Operation OperationType `json:"operation"`
	Source    string        `json:"source,omitempty"`
	Target    string        `json:"target,omitempty"`
	State     EntryState    `json:"state"`
	Checksum  string        `json:"checksum,omitempty"`
	Steps     []Step        `json:"steps"`
}

// Context keys for journal-related values
type contextKey string

const (
	journalManagerKey contextKey = "journal_manager"
	journalEntryKey   contextKey = "journal_entry"
)

// WithJournalManager adds a JournalManager to the context
func WithJournalManager(ctx context.Context, jm *JournalManager) context.Context {
	return context.WithValue(ctx, journalManagerKey, jm)
}

// WithJournalEntry adds a JournalEntry to the context
func WithJournalEntry(ctx context.Context, entry *JournalEntry) context.Context {
	return context.WithValue(ctx, journalEntryKey, entry)
}

// GetJournalManager retrieves the JournalManager from the context
func GetJournalManager(ctx context.Context) (*JournalManager, error) {
	jm, ok := ctx.Value(journalManagerKey).(*JournalManager)
	if !ok {
		return nil, fmt.Errorf("journal manager not found in context")
	}
	return jm, nil
}

// GetJournalEntry retrieves the JournalEntry from the context
func GetJournalEntry(ctx context.Context) (*JournalEntry, error) {
	entry, ok := ctx.Value(journalEntryKey).(*JournalEntry)
	if !ok {
		return nil, fmt.Errorf("journal entry not found in context")
	}
	return entry, nil
}

// AddStep creates and adds a new step to the journal entry and saves it
func (e *JournalEntry) AddStep(ctx context.Context, stepType StepType, description string, source, target string) (*Step, error) {
	jm, err := GetJournalManager(ctx)
	if err != nil {
		return nil, err
	}

	step := Step{
		Type:        stepType,
		Status:      StepStatusPending,
		Description: description,
		Source:      source,
		Target:      target,
		StartTime:   time.Now(),
	}
	e.Steps = append(e.Steps, step)
	if err := jm.UpdateEntry(e); err != nil {
		return nil, fmt.Errorf("error saving step: %v", err)
	}
	return &e.Steps[len(e.Steps)-1], nil
}

// StartStep marks a step as running and saves the entry
func StartStep(ctx context.Context, step *Step) error {
	entry, err := GetJournalEntry(ctx)
	if err != nil {
		return err
	}
	jm, err := GetJournalManager(ctx)
	if err != nil {
		return err
	}

	step.Status = StepStatusRunning
	return jm.UpdateEntry(entry)
}

// CompleteStep marks a step as completed and saves the entry
func CompleteStep(ctx context.Context, step *Step, details string) error {
	entry, err := GetJournalEntry(ctx)
	if err != nil {
		return err
	}
	jm, err := GetJournalManager(ctx)
	if err != nil {
		return err
	}

	step.Status = StepStatusCompleted
	step.Details = details
	step.EndTime = time.Now()
	return jm.UpdateEntry(entry)
}

// FailStep marks a step as failed and saves the entry
func FailStep(ctx context.Context, step *Step, err error) error {
	entry, err2 := GetJournalEntry(ctx)
	if err2 != nil {
		return err2
	}
	jm, err2 := GetJournalManager(ctx)
	if err2 != nil {
		return err2
	}

	step.Status = StepStatusFailed
	step.Error = err.Error()
	step.EndTime = time.Now()
	return jm.UpdateEntry(entry)
}

// FailEntry marks the last step as failed and moves the entry to the failed state
func FailEntry(ctx context.Context, err error) error {
	entry, err2 := GetJournalEntry(ctx)
	if err2 != nil {
		return err2
	}
	jm, err2 := GetJournalManager(ctx)
	if err2 != nil {
		return err2
	}

	// Get the last step
	if len(entry.Steps) == 0 {
		return fmt.Errorf("no steps in entry")
	}
	step := &entry.Steps[len(entry.Steps)-1]

	// Update step status
	step.Status = StepStatusFailed
	step.Error = err.Error()
	step.EndTime = time.Now()

	// Update entry
	if err := jm.UpdateEntry(entry); err != nil {
		return err
	}

	// Move entry to failed state
	return jm.MoveEntry(entry, EntryStateFailed)
}

// CompleteEntry moves the entry to the completed state
func CompleteEntry(ctx context.Context) error {
	entry, err := GetJournalEntry(ctx)
	if err != nil {
		return err
	}
	jm, err := GetJournalManager(ctx)
	if err != nil {
		return err
	}

	// Update entry
	if err := jm.UpdateEntry(entry); err != nil {
		return err
	}

	// Move entry to completed state
	return jm.MoveEntry(entry, EntryStateCompleted)
}

// AddStepToCurrentEntry creates a new step in the current journal entry from context
func AddStepToCurrentEntry(ctx context.Context, stepType StepType, description string, source, target string) (*Step, error) {
	entry, err := GetJournalEntry(ctx)
	if err != nil {
		return nil, err
	}
	return entry.AddStep(ctx, stepType, description, source, target)
}

// Step represents a single step in a journal entry
type Step struct {
	Type        StepType   `json:"type"`
	Status      StepStatus `json:"status"`
	Error       string     `json:"error,omitempty"`
	Description string     `json:"description,omitempty"`
	Source      string     `json:"source,omitempty"`
	Target      string     `json:"target,omitempty"`
	Details     string     `json:"details,omitempty"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     time.Time  `json:"end_time,omitempty"`
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
func (jm *JournalManager) CreateEntry(operation OperationType, source, target string) (*JournalEntry, error) {
	entry := &JournalEntry{
		ID:        generateOperationID(string(operation)),
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
func (jm *JournalManager) MoveEntry(entry *JournalEntry, newState EntryState) error {
	oldPath := filepath.Join(jm.journalDir, string(entry.State), entry.ID+".json")

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
	states := []EntryState{EntryStateCurrent, EntryStateCompleted, EntryStateFailed}
	for _, state := range states {
		path := filepath.Join(jm.journalDir, string(state), id+".json")
		if _, err := jm.fsys.Stat(path); err == nil {
			return jm.readEntry(path)
		}
	}
	return nil, fmt.Errorf("entry not found: %s", id)
}

// ListEntries lists all journal entries in a given state
func (jm *JournalManager) ListEntries(state string) ([]*JournalEntry, error) {
	entries := make([]*JournalEntry, 0)

	// If state is empty, list entries from all states
	states := []EntryState{EntryStateCurrent, EntryStateCompleted, EntryStateFailed}
	if state != "" {
		// If state is specified, only use that state
		states = []EntryState{EntryState(state)}
	}

	// Read entries from each state directory
	for _, s := range states {
		dir := filepath.Join(jm.journalDir, string(s))

		// Read directory
		dirFile, err := jm.fsys.Open(dir)
		if err != nil {
			// Skip if directory doesn't exist
			continue
		}

		// Read all entries
		dirEntries, err := dirFile.(fs.ReadDirFile).ReadDir(-1)
		dirFile.Close() // Close immediately after reading
		if err != nil {
			return nil, fmt.Errorf("error reading directory %s: %v", dir, err)
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
	}

	return entries, nil
}

// Helper functions

func (jm *JournalManager) saveEntry(entry *JournalEntry) error {
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling entry: %v", err)
	}

	path := filepath.Join(jm.journalDir, string(entry.State), entry.ID+".json")
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
