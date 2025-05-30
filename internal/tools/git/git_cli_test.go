package git

import (
	"os"
	"strings"
	"testing"
)

func TestGitStatusCli(t *testing.T) {
	// Missing app name
	os.Args = []string{"layered-code", "tool", "git_status"}
	err := GitStatusCli()
	if err == nil || !strings.Contains(err.Error(), "requires exactly 1 argument") {
		t.Error("Expected error for missing app name")
	}

	// Too many args
	os.Args = []string{"layered-code", "tool", "git_status", "app1", "app2"}
	err = GitStatusCli()
	if err == nil || !strings.Contains(err.Error(), "requires exactly 1 argument") {
		t.Error("Expected error for too many args")
	}
}

func TestGitDiffCli(t *testing.T) {
	// Missing app name
	os.Args = []string{"layered-code", "tool", "git_diff"}
	err := GitDiffCli()
	if err == nil || !strings.Contains(err.Error(), "requires at least 1 argument") {
		t.Error("Expected error for missing app name")
	}
}

func TestGitCommitCli(t *testing.T) {
	// Missing app name
	os.Args = []string{"layered-code", "tool", "git_commit"}
	err := GitCommitCli()
	if err == nil || !strings.Contains(err.Error(), "requires at least 1 argument") {
		t.Error("Expected error for missing app name")
	}

	// Missing message flag value
	os.Args = []string{"layered-code", "tool", "git_commit", "app", "-m"}
	err = GitCommitCli()
	if err == nil || !strings.Contains(err.Error(), "-m flag requires a message") {
		t.Error("Expected error for missing message")
	}
}

func TestGitLogCli(t *testing.T) {
	// Missing app name
	os.Args = []string{"layered-code", "tool", "git_log"}
	err := GitLogCli()
	if err == nil || !strings.Contains(err.Error(), "requires at least 1 argument") {
		t.Error("Expected error for missing app name")
	}

	// Invalid limit
	os.Args = []string{"layered-code", "tool", "git_log", "app", "-n", "invalid"}
	err = GitLogCli()
	if err == nil || !strings.Contains(err.Error(), "invalid limit") {
		t.Error("Expected error for invalid limit")
	}
}

func TestGitAddCli(t *testing.T) {
	// Missing app name
	os.Args = []string{"layered-code", "tool", "git_add"}
	err := GitAddCli()
	if err == nil || !strings.Contains(err.Error(), "requires at least 1 argument") {
		t.Error("Expected error for missing app name")
	}
}

func TestGitInitCli(t *testing.T) {
	// Missing app name
	os.Args = []string{"layered-code", "tool", "git_init"}
	err := GitInitCli()
	if err == nil || !strings.Contains(err.Error(), "requires at least 1 argument") {
		t.Error("Expected error for missing app name")
	}
}

func TestGitResetCli(t *testing.T) {
	// Missing app name
	os.Args = []string{"layered-code", "tool", "git_reset"}
	err := GitResetCli()
	if err == nil || !strings.Contains(err.Error(), "requires at least 2 arguments") {
		t.Error("Expected error for missing arguments")
	}

	// Missing commit hash
	os.Args = []string{"layered-code", "tool", "git_reset", "app"}
	err = GitResetCli()
	if err == nil || !strings.Contains(err.Error(), "requires at least 2 arguments") {
		t.Error("Expected error for missing commit hash")
	}

	// Invalid mode
	os.Args = []string{"layered-code", "tool", "git_reset", "app", "abc123", "invalid"}
	err = GitResetCli()
	if err == nil || !strings.Contains(err.Error(), "invalid reset mode") {
		t.Error("Expected error for invalid mode")
	}
}

func TestGitRevertCli(t *testing.T) {
	// Missing app name
	os.Args = []string{"layered-code", "tool", "git_revert"}
	err := GitRevertCli()
	if err == nil || !strings.Contains(err.Error(), "requires at least 2 arguments") {
		t.Error("Expected error for missing arguments")
	}

	// Missing commit hash
	os.Args = []string{"layered-code", "tool", "git_revert", "app"}
	err = GitRevertCli()
	if err == nil || !strings.Contains(err.Error(), "requires at least 2 arguments") {
		t.Error("Expected error for missing commit hash")
	}
}

func TestGitCheckoutCli(t *testing.T) {
	// Missing app name
	os.Args = []string{"layered-code", "tool", "git_checkout"}
	err := GitCheckoutCli()
	if err == nil || !strings.Contains(err.Error(), "requires at least 2 arguments") {
		t.Error("Expected error for missing arguments")
	}

	// Missing target
	os.Args = []string{"layered-code", "tool", "git_checkout", "app"}
	err = GitCheckoutCli()
	if err == nil || !strings.Contains(err.Error(), "requires at least 2 arguments") {
		t.Error("Expected error for missing target")
	}

	// Missing files after --files flag
	os.Args = []string{"layered-code", "tool", "git_checkout", "app", "--files"}
	err = GitCheckoutCli()
	if err == nil || !strings.Contains(err.Error(), "--files requires at least one file") {
		t.Error("Expected error for missing files after --files")
	}
}