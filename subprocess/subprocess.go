package subprocess

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

// LogFormatter is a function type that formats a single line of output
type LogFormatter func(line string) string

type Subprocess struct {
	Command         string
	Args            []string
	StdoutFormatter LogFormatter // Function to format stdout lines
	StderrFormatter LogFormatter // Function to format stderr lines
	Stdin           io.Reader
	Stdout          io.Writer
	Stderr          io.Writer
	Dir             string
	Env             []string
}

// Run executes the subprocess without context (for backward compatibility)
func Run(sb *Subprocess) error {
	return RunWithContext(context.Background(), sb)
}

// RunWithContext executes the subprocess with the provided context
func RunWithContext(ctx context.Context, sb *Subprocess) error {
	if sb.Stdin == nil {
		sb.Stdin = os.Stdin
	}
	if sb.Stdout == nil {
		sb.Stdout = os.Stdout
	}
	if sb.Stderr == nil {
		sb.Stderr = os.Stderr
	}

	// Use Background context if ctx is nil
	if ctx == nil {
		ctx = context.Background()
	}

	// Use exec.CommandContext to enable context control
	cmd := exec.CommandContext(ctx, sb.Command, sb.Args...)
	cmd.Dir = sb.Dir
	cmd.Env = sb.Env
	cmd.Stdin = sb.Stdin

	m := new(sync.Mutex)
	wg := &sync.WaitGroup{}

	// Variables to collect errors from goroutines
	var stdoutErr, stderrErr error

	// Use pipes to add a prefix to each line
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := scanLines(stdout, sb.Stdout, sb.StdoutFormatter, m); err != nil {
			stdoutErr = fmt.Errorf("failed to scan stdout: %w", err)
		}
	}()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := scanLines(stderr, sb.Stderr, sb.StderrFormatter, m); err != nil {
			stderrErr = fmt.Errorf("failed to scan stderr: %w", err)
		}
	}()

	// Start the process
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Wait for process completion (ensures process is fully terminated)
	processErr := cmd.Wait()

	// Wait for stdout/stderr processing completion
	wg.Wait()

	// Return process error first, then scan errors
	if processErr != nil {
		return processErr
	}

	if stdoutErr != nil {
		return stdoutErr
	}
	if stderrErr != nil {
		return stderrErr
	}

	return nil
}

func scanLines(src io.ReadCloser, dest io.Writer, formatter LogFormatter, m *sync.Mutex) error {
	defer src.Close()
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		line := scanner.Text()

		// Apply formatter if provided
		if formatter != nil {
			line = formatter(line)
		}

		// Prevent mixing data in a line
		m.Lock()
		_, _ = fmt.Fprintf(dest, "%s\n", line)
		m.Unlock()
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan lines: %w", err)
	}
	return nil
}

// Built-in formatter functions

// PrefixFormatter creates a formatter that adds a prefix to each line
func PrefixFormatter(prefix string) LogFormatter {
	return func(line string) string {
		return prefix + line
	}
}
