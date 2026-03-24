package claudecode

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	"github.com/branchly/branchly-runner/internal/domain"
)

type Agent struct{}

func New() *Agent {
	return &Agent{}
}

func (a *Agent) Run(ctx context.Context, input domain.AgentInput) (string, error) {
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return "", fmt.Errorf("claude code cli not found in PATH: %w", err)
	}
	cmd := exec.CommandContext(ctx, claudePath, "--dangerously-skip-permissions", "--print", input.Prompt)
	cmd.Dir = input.WorkDir
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("stderr pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("start claude: %w", err)
	}
	var lastLine string
	var lineMu sync.Mutex
	setLast := func(s string) {
		t := strings.TrimSpace(s)
		if t != "" {
			lineMu.Lock()
			lastLine = t
			lineMu.Unlock()
		}
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		scanLines(stdout, func(line string) {
			if input.OnLog != nil {
				input.OnLog(domain.LogLevelInfo, line)
			}
			setLast(line)
		})
	}()
	go func() {
		defer wg.Done()
		scanLines(stderr, func(line string) {
			if input.OnLog != nil {
				input.OnLog(domain.LogLevelError, line)
			}
		})
	}()
	wg.Wait()
	if err := cmd.Wait(); err != nil {
		lineMu.Lock()
		s := lastLine
		lineMu.Unlock()
		if s != "" {
			return s, fmt.Errorf("claude exited: %w", err)
		}
		return "", fmt.Errorf("claude exited: %w", err)
	}
	lineMu.Lock()
	defer lineMu.Unlock()
	return lastLine, nil
}

func scanLines(r io.Reader, fn func(string)) {
	br := bufio.NewReaderSize(r, 64*1024)
	for {
		line, err := br.ReadString('\n')
		if line != "" {
			line = strings.TrimSuffix(line, "\n")
			line = strings.TrimSuffix(line, "\r")
			fn(line)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return
		}
	}
}
