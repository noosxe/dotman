package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/noosxe/dotman/internal/config"
	"github.com/noosxe/dotman/internal/journal"
	"github.com/spf13/cobra"
)

var (
	stateFilters     []string
	operationFilters []string
)

var journalCmd = &cobra.Command{
	Use:   "journal",
	Short: "Show the status of actions from the journal",
	Long: `Show the status of actions from the journal, including completed, failed, and current operations.
The journal keeps track of all operations performed by dotman.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Validate state filters
		for _, state := range stateFilters {
			switch journal.EntryState(state) {
			case journal.EntryStateCurrent, journal.EntryStateCompleted, journal.EntryStateFailed:
				// Valid state
			default:
				return fmt.Errorf("invalid state '%s'. Valid states are: current, completed, failed", state)
			}
		}

		// Validate operation filters
		for _, op := range operationFilters {
			switch journal.OperationType(op) {
			case journal.OperationTypeAdd, journal.OperationTypeRemove, journal.OperationTypeLink:
				// Valid operation
			default:
				return fmt.Errorf("invalid operation '%s'. Valid operations are: add, remove, link", op)
			}
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		cfg, err := config.LoadConfig(configPath, fsys)
		if err != nil {
			return fmt.Errorf("error loading config: %v", err)
		}

		// Initialize journal manager with the correct path
		jm := journal.NewJournalManager(fsys, filepath.Join(cfg.DotmanDir, "journal"))

		// List entries with state filters
		var allEntries []*journal.JournalEntry
		if len(stateFilters) == 0 {
			// If no state filters specified, get all entries
			entries, err := jm.ListEntries("")
			if err != nil {
				return fmt.Errorf("error listing journal entries: %v", err)
			}
			allEntries = entries
		} else {
			// Get entries for each specified state
			for _, state := range stateFilters {
				entries, err := jm.ListEntries(journal.EntryState(state))
				if err != nil {
					return fmt.Errorf("error listing journal entries for state '%s': %v", state, err)
				}
				allEntries = append(allEntries, entries...)
			}
		}

		// Filter by operation if specified
		if len(operationFilters) > 0 {
			filteredEntries := make([]*journal.JournalEntry, 0)
			for _, entry := range allEntries {
				for _, op := range operationFilters {
					if string(entry.Operation) == op {
						filteredEntries = append(filteredEntries, entry)
						break
					}
				}
			}
			allEntries = filteredEntries
		}

		if len(allEntries) == 0 {
			var filterMsg string
			if len(stateFilters) > 0 || len(operationFilters) > 0 {
				filterMsg = "No journal entries found"
				if len(stateFilters) > 0 {
					filterMsg += fmt.Sprintf(" in states: %s", strings.Join(stateFilters, ", "))
				}
				if len(operationFilters) > 0 {
					filterMsg += fmt.Sprintf(" with operations: %s", strings.Join(operationFilters, ", "))
				}
			} else {
				filterMsg = "No journal entries found"
			}
			fmt.Println(filterMsg)
			return nil
		}

		// Print entries in reverse chronological order
		for i := len(allEntries) - 1; i >= 0; i-- {
			entry := allEntries[i]
			fmt.Printf("\nOperation: %s\n", entry.Operation)
			fmt.Printf("ID: %s\n", entry.ID)
			fmt.Printf("Timestamp: %s\n", entry.Timestamp.Format(time.RFC3339))
			fmt.Printf("State: %s\n", entry.State)
			if entry.Source != "" {
				fmt.Printf("Source: %s\n", entry.Source)
			}
			if entry.Target != "" {
				fmt.Printf("Target: %s\n", entry.Target)
			}

			// Print steps
			if len(entry.Steps) > 0 {
				fmt.Println("\nSteps:")
				for _, step := range entry.Steps {
					fmt.Printf("  - %s: %s\n", step.Type, step.Status)
					if step.Description != "" {
						fmt.Printf("    Description: %s\n", step.Description)
					}
					if step.Error != "" {
						fmt.Printf("    Error: %s\n", step.Error)
					}
					if step.Details != "" {
						fmt.Printf("    Details: %s\n", step.Details)
					}
					if !step.StartTime.IsZero() {
						fmt.Printf("    Started: %s\n", step.StartTime.Format(time.RFC3339))
					}
					if !step.EndTime.IsZero() {
						fmt.Printf("    Ended: %s\n", step.EndTime.Format(time.RFC3339))
					}
				}
			}
			fmt.Println("----------------------------------------")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(journalCmd)

	// Add state filter flag
	journalCmd.Flags().StringSliceVarP(&stateFilters, "state", "s", nil, "Filter entries by state (current, completed, failed). Can be specified multiple times.")

	// Add operation filter flag
	journalCmd.Flags().StringSliceVarP(&operationFilters, "operation", "o", nil, "Filter entries by operation type (add, remove, link). Can be specified multiple times.")
}
