package gui

import (
	"context"
)

// DialogProvider defines the interface for displaying user input dialogs
type DialogProvider interface {
	ShowInputDialog(ctx context.Context, prompt string, title string) (string, error)
	CheckDependencies() error
}
