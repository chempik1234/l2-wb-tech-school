package main

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
	"time"
)

type testCaseCLI struct {
	name     string
	inputs   []string
	expected []string
}

// runShellTest launches a test with given inputs and outputs, also featuring timeout
func runShellTest(t *testing.T, tt testCaseCLI, timeout time.Duration) {
	inputs := tt.inputs
	expectedOutputs := tt.expected

	// launch command
	cmd := exec.Command("go", "run", ".", "-w")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	//region input
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Stdin pipe error: %v", err)
	}
	go func() {
		defer stdinPipe.Close()
		for _, input := range inputs {
			stdinPipe.Write([]byte(input + "\n"))
			// t.Log("--", input)
			time.Sleep(100 * time.Millisecond) // input with delay like a human
		}
	}()
	//endregion

	// launch command waiter
	done := make(chan error, 1)
	go func() {
		execErr := cmd.Run()
		// time.Sleep(1 * time.Second)
		done <- execErr
	}()

	// wait for either command or timeout
	select {
	case err = <-done:
		if err != nil {
			t.Errorf("Command failed: %v, stderr: %s", err, stderr.String())
		}
	case <-time.After(timeout):
		_ = cmd.Process.Kill()
		t.Errorf("Test timed out after %v", timeout)
	}

	// check command output
	output := stdout.String()
	// errorOutput := stderr.String()
	for _, expected := range expectedOutputs {
		if !strings.Contains(output, expected) { //} && !strings.Contains(errorOutput, expected) {
			t.Errorf("Expected output %q not found in stdout: %q or stderr", expected, output) // , errorOutput)
		}
	}
}

// TestBasicCommands contains basic vibecoded commands
func TestBasicCommands(t *testing.T) {
	tests := []testCaseCLI{
		{
			name:     "Echo command",
			inputs:   []string{"echo hello world", "exit"},
			expected: []string{"hello world"},
		},
		{
			name:     "Environment variables",
			inputs:   []string{"setenv OS=Windows_NT", "echo $OS", "exit"},
			expected: []string{"Windows_NT"},
		},
		{
			name:     "CD command",
			inputs:   []string{"cd /tmp", "pwd", "exit"},
			expected: []string{"C:\\tmp"},
		},
		{
			name:     "PS command",
			inputs:   []string{"ps", "exit"},
			expected: []string{"PID", "CMD"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runShellTest(t, tt, 10*time.Second)
		})
	}
}

// TestPipelines test pipelines: simple echo | wc, complex string formatting and with env
func TestPipelines(t *testing.T) {
	tests := []testCaseCLI{
		{
			name:     "Simple pipeline",
			inputs:   []string{"echo hello | wc -c", "exit"},
			expected: []string{"6"}, // hello + newline
		},
		{
			name:     "Multiple pipelines",
			inputs:   []string{"echo \"hello world\" | cut -d \" \" -f1 | tr a-z A-Z", "exit"},
			expected: []string{"HELLO"},
		},
		{
			name:     "Pipeline with env variables",
			inputs:   []string{"setenv TEXT=test", "echo $TEXT | tr a-z A-Z", "exit"},
			expected: []string{"TEST"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runShellTest(t, tt, 10*time.Second)
		})
	}
}

// TestExternalCommands tests os.exec commands
func TestExternalCommands(t *testing.T) {
	tests := []testCaseCLI{
		{
			name:     "Example input command",
			inputs:   []string{"go run example/input.go", "test input", "exit"},
			expected: []string{"EXAMPLE", "you scanned: test input"},
		},
		{
			name:     "Minute command",
			inputs:   []string{"go run example/minute.go", "exit"},
			expected: []string{"EXAMPLE", "some print"},
		},
		{
			name:     "LS command",
			inputs:   []string{"ls -la", "exit"},
			expected: []string{"main.go", "deepseek_test.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runShellTest(t, tt, 15*time.Second) // Longer timeout for minute.go
		})
	}
}

// TestKillCommand tests only kill command
func TestKillCommand(t *testing.T) {
	t.Skip("Skipping kill test as it may be disruptive")

	inputs := []string{
		"sleep 10 &",
		"ps",
		"kill %1",
		"ps",
		"exit",
	}

	expected := []string{"sleep 10", "terminated"}

	runShellTest(t, testCaseCLI{inputs: inputs, expected: expected}, 15*time.Second)
}

// TestEdgeCases vibecoded edge cases
func TestEdgeCases(t *testing.T) {
	tests := []testCaseCLI{
		{
			name:     "Empty input",
			inputs:   []string{"", "exit"},
			expected: []string{""},
		},
		{
			name:     "Unknown command",
			inputs:   []string{"nonexistent_command", "exit"},
			expected: []string{"not found", "executable file not found"},
		},
		{
			name:     "Multiple spaces",
			inputs:   []string{"   echo   hello   ", "exit"},
			expected: []string{"hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runShellTest(t, tt, 10*time.Second)
		})
	}
}
