package gui

import (
	"runtime"
	"testing"
)

func TestZenityDialogCreation(t *testing.T) {
	dialog := NewZenityDialog()
	if dialog == nil {
		t.Fatal("Failed to create ZenityDialog")
	}
}

func TestDependencyCheck(t *testing.T) {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		err := CheckDependencies()
		if err != nil {
			t.Logf("Dependency check failed: %v", err)
			// Don't fail the test as CI environments might not have zenity installed
		}
	} else {
		t.Skip("Skipping dependency check on unsupported OS")
	}
}

// Note: We don't test the actual dialog display functions here because they're interactive
// and would block automated testing. Those would be better tested manually or with a mock.
