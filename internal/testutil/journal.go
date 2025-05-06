package testutil

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/noosxe/dotman/internal/fs"
	"github.com/noosxe/dotman/internal/journal"
)

// VerifyStep checks if a journal step matches the expected values
func VerifyStep(t *testing.T, step journal.Step, expectedType journal.StepType, expectedStatus journal.StepStatus, expectedDescription string) {
	t.Helper()

	if step.Type != expectedType {
		t.Fatalf("expected step type '%s', got '%s'", expectedType, step.Type)
	}

	if step.Status != expectedStatus {
		t.Fatalf("expected step status '%s', got '%s'", expectedStatus, step.Status)
	}

	if step.Description != expectedDescription {
		t.Fatalf("expected step description '%s', got '%s'", expectedDescription, step.Description)
	}
}

// VerifyStepWithSourceTarget checks if a journal step matches the expected values including source and target
func VerifyStepWithSourceTarget(t *testing.T, step journal.Step, expectedType journal.StepType, expectedStatus journal.StepStatus, expectedDescription, expectedSource, expectedTarget string) {
	t.Helper()

	VerifyStep(t, step, expectedType, expectedStatus, expectedDescription)

	if step.Source != expectedSource {
		t.Fatalf("expected step source '%s', got '%s'", expectedSource, step.Source)
	}

	if step.Target != expectedTarget {
		t.Fatalf("expected step target '%s', got '%s'", expectedTarget, step.Target)
	}
}

// VerifyStepWithDetails checks if a journal step matches the expected values including details
func VerifyStepWithDetails(t *testing.T, step journal.Step, expectedType journal.StepType, expectedStatus journal.StepStatus, expectedDescription, expectedDetails string) {
	t.Helper()

	VerifyStep(t, step, expectedType, expectedStatus, expectedDescription)

	if step.Details != expectedDetails {
		t.Fatalf("expected step details '%s', got '%s'", expectedDetails, step.Details)
	}
}

// VerifyStepWithError checks if a journal step matches the expected values including error
func VerifyStepWithError(t *testing.T, step journal.Step, expectedType journal.StepType, expectedStatus journal.StepStatus, expectedDescription, expectedError string) {
	t.Helper()

	VerifyStep(t, step, expectedType, expectedStatus, expectedDescription)

	if step.Error != expectedError {
		t.Fatalf("expected step error '%s', got '%s'", expectedError, step.Error)
	}
}

// VerifyEntry checks if a journal entry matches the expected values
func VerifyEntry(t *testing.T, entry *journal.JournalEntry, expectedOperation journal.OperationType, expectedState journal.EntryState) {
	t.Helper()

	if entry.Operation != expectedOperation {
		t.Fatalf("expected operation '%s', got '%s'", expectedOperation, entry.Operation)
	}

	if entry.State != expectedState {
		t.Fatalf("expected state '%s', got '%s'", expectedState, entry.State)
	}
}

// VerifyEntryWithSourceTarget checks if a journal entry matches the expected values including source and target
func VerifyEntryWithSourceTarget(t *testing.T, entry *journal.JournalEntry, expectedOperation journal.OperationType, expectedState journal.EntryState, expectedSource, expectedTarget string) {
	t.Helper()

	VerifyEntry(t, entry, expectedOperation, expectedState)

	if entry.Source != expectedSource {
		t.Fatalf("expected source '%s', got '%s'", expectedSource, entry.Source)
	}

	if entry.Target != expectedTarget {
		t.Fatalf("expected target '%s', got '%s'", expectedTarget, entry.Target)
	}
}

// VerifyEntryWithSteps checks if a journal entry matches the expected values and has the expected number of steps
func VerifyEntryWithSteps(t *testing.T, entry *journal.JournalEntry, expectedOperation journal.OperationType, expectedState journal.EntryState, expectedStepCount int) {
	t.Helper()

	VerifyEntry(t, entry, expectedOperation, expectedState)

	if len(entry.Steps) != expectedStepCount {
		t.Fatalf("expected %d steps, got %d", expectedStepCount, len(entry.Steps))
	}
}

// VerifyEntryWithChecksum checks if a journal entry matches the expected values including checksum
func VerifyEntryWithChecksum(t *testing.T, entry *journal.JournalEntry, expectedOperation journal.OperationType, expectedState journal.EntryState, expectedChecksum string) {
	t.Helper()

	VerifyEntry(t, entry, expectedOperation, expectedState)

	if entry.Checksum != expectedChecksum {
		t.Fatalf("expected checksum '%s', got '%s'", expectedChecksum, entry.Checksum)
	}
}

// SetupJournalManager creates and initializes a journal manager for testing
func SetupJournalManager(t *testing.T, fsys fs.FileSystem, dotmanDir string) *journal.JournalManager {
	jm := journal.NewJournalManager(fsys, filepath.Join(dotmanDir, "journal"))
	if err := jm.Initialize(); err != nil {
		t.Fatalf("failed to initialize journal: %v", err)
	}
	return jm
}

// SetupContextWithJournal creates a context with a journal manager and entry
func SetupContextWithJournal(t *testing.T, jm *journal.JournalManager, operationType journal.OperationType, source, target string) context.Context {
	ctx := context.Background()
	ctx = journal.WithJournalManager(ctx, jm)

	entry, err := jm.CreateEntry(operationType, source, target)
	if err != nil {
		t.Fatalf("failed to create journal entry: %v", err)
	}

	ctx = journal.WithJournalEntry(ctx, entry)
	return ctx
}

// VerifyJournalEntryCount verifies that there are exactly the expected number of entries with the given state
func VerifyJournalEntryCount(t *testing.T, jm *journal.JournalManager, state journal.EntryState, expectedCount int) {
	t.Helper()

	entries, err := jm.ListEntries(state)
	if err != nil {
		t.Fatalf("failed to get journal entries: %v", err)
	}

	if len(entries) != expectedCount {
		t.Fatalf("expected %d entries with state '%s', got %d", expectedCount, state, len(entries))
	}
}
