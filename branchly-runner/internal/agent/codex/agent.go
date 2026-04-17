package codex

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
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
	codexPath, err := exec.LookPath("codex")
	if err != nil {
		return "", fmt.Errorf("codex cli not found in PATH: %w", err)
	}
	cmd := exec.CommandContext(ctx, codexPath, "--approval-mode", "full-auto", "-q", input.Prompt)
	cmd.Dir = input.WorkDir
	cmd.Stdin = strings.NewReader("")
	env := append(os.Environ(), "CI=true", "TERM=dumb")
	if input.APIKey != "" {
		env = append(env, "OPENAI_API_KEY="+input.APIKey)
	}
	cmd.Env = env

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("stderr pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("start codex: %w", err)
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

	var stderrLines []string
	var stderrMu sync.Mutex
	const stderrCap = 40
	pushStderr := func(line string) {
		stderrMu.Lock()
		defer stderrMu.Unlock()
		stderrLines = append(stderrLines, line)
		if len(stderrLines) > stderrCap {
			stderrLines = stderrLines[len(stderrLines)-stderrCap:]
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
			pushStderr(line)
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
		stderrMu.Lock()
		tail := strings.TrimSpace(strings.Join(stderrLines, "\n"))
		stderrMu.Unlock()
		if tail == "" {
			tail = "(no stderr; check OPENAI_API_KEY is set in the runner environment)"
		}
		if s != "" {
			return s, fmt.Errorf("codex exited: %w — %s", err, tail)
		}
		return "", fmt.Errorf("codex exited: %w — %s", err, tail)
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
