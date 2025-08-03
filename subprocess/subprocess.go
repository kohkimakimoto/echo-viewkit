package subprocess

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

type Subprocess struct {
	Command   string
	Args      []string
	LogPrefix string
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer
	Dir       string
	Env       []string
}

func Run(sb *Subprocess) error {
	if sb.Stdin == nil {
		sb.Stdin = os.Stdin
	}
	if sb.Stdout == nil {
		sb.Stdout = os.Stdout
	}
	if sb.Stderr == nil {
		sb.Stderr = os.Stderr
	}

	cmd := exec.Command(sb.Command, sb.Args...)
	cmd.Dir = sb.Dir
	cmd.Env = sb.Env
	cmd.Stdin = sb.Stdin

	m := new(sync.Mutex)
	wg := &sync.WaitGroup{}

	// use pipes to add a prefix to each line.
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	wg.Add(1)
	go func() {
		if err := scanLines(stdout, sb.Stdout, sb.LogPrefix, m); err != nil {
			_, _ = fmt.Fprintf(sb.Stderr, "failed to scan stdout: %v", err)
		}
		wg.Done()
	}()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}
	wg.Add(1)
	go func() {
		if err := scanLines(stderr, sb.Stderr, sb.LogPrefix, m); err != nil {
			_, _ = fmt.Fprintf(sb.Stderr, "failed to scan stderr: %v", err)
		}
		wg.Done()
	}()

	err = cmd.Run()
	if err != nil {
		return err
	}
	wg.Wait()

	return nil
}

func scanLines(src io.ReadCloser, dest io.Writer, prefix string, m *sync.Mutex) error {
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		// prevent mixing data in a line.
		m.Lock()
		_, _ = fmt.Fprintf(dest, "%s%s\n", prefix, scanner.Text())
		m.Unlock()
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan lines: %w", err)
	}
	return nil
}
