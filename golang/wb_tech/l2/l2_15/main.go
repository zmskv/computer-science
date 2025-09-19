package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var (
	activeProcessesMu sync.Mutex
	activeProcesses   []*exec.Cmd
)

func saveActive(cmds []*exec.Cmd) {
	activeProcessesMu.Lock()
	activeProcesses = cmds
	activeProcessesMu.Unlock()
}

func clearActive() {
	saveActive(nil)
}

func interruptAll() {
	activeProcessesMu.Lock()
	cmds := append([]*exec.Cmd(nil), activeProcesses...)
	activeProcessesMu.Unlock()

	for _, cmd := range cmds {
		if cmd != nil && cmd.Process != nil {
			_ = cmd.Process.Signal(os.Interrupt)
		}
	}
}

func expandVariables(token string) string {
	re := regexp.MustCompile(`\$[A-Za-z_][A-Za-z0-9_]*`)
	return re.ReplaceAllStringFunc(token, func(varRef string) string {
		return os.Getenv(varRef[1:])
	})
}

func tokenizeLine(input string) []string {
	replacements := []struct{ old, new string }{
		{"&&", " && "}, {"||", " || "}, {"|", " | "}, {">", " > "}, {"<", " < "},
	}
	for _, r := range replacements {
		input = strings.ReplaceAll(input, r.old, r.new)
	}
	fields := strings.Fields(input)
	for i, tok := range fields {
		fields[i] = expandVariables(tok)
	}
	return fields
}

func isInternal(cmd string) bool {
	switch cmd {
	case "cd", "pwd", "echo", "kill", "ps":
		return true
	}
	return false
}

func executeInternal(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (int, error) {
	if len(args) == 0 {
		return 0, nil
	}
	switch args[0] {
	case "cd":
		path := "~"
		if len(args) > 1 {
			path = args[1]
		}
		if path == "~" {
			if h := os.Getenv("HOME"); h != "" {
				path = h
			}
		}
		if !filepath.IsAbs(path) {
			cwd, _ := os.Getwd()
			path = filepath.Join(cwd, path)
		}
		if err := os.Chdir(path); err != nil {
			return 1, err
		}
		return 0, nil
	case "pwd":
		cwd, err := os.Getwd()
		if err != nil {
			return 1, err
		}
		fmt.Fprintln(stdout, cwd)
		return 0, nil
	case "echo":
		fmt.Fprintln(stdout, strings.Join(args[1:], " "))
		return 0, nil
	case "kill":
		if len(args) < 2 {
			return 1, errors.New("kill: missing pid")
		}
		pid, err := strconv.Atoi(args[1])
		if err != nil {
			return 1, err
		}
		proc, err := os.FindProcess(pid)
		if err != nil {
			return 1, err
		}
		if err := proc.Kill(); err != nil {
			return 1, err
		}
		return 0, nil
	case "ps":
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("tasklist")
		} else {
			cmd = exec.Command("ps", "-e", "-o", "pid,comm")
		}
		cmd.Stdin = stdin
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			return 1, err
		}
		return 0, nil
	}
	return 127, fmt.Errorf("unknown internal command: %s", args[0])
}

func buildStages(tokens []string) ([][]string, error) {
	var stages [][]string
	var current []string
	for _, t := range tokens {
		if t == "|" {
			if len(current) == 0 {
				return nil, errors.New("empty command in pipeline")
			}
			stages = append(stages, current)
			current = nil
			continue
		}
		current = append(current, t)
	}
	if len(current) > 0 {
		stages = append(stages, current)
	}
	return stages, nil
}

func handleRedirections(stage []string) (args []string, stdin io.Reader, stdout io.Writer, err error) {
	args = []string{}
	for i := 0; i < len(stage); i++ {
		switch stage[i] {
		case ">":
			if i+1 >= len(stage) {
				return nil, nil, nil, errors.New("missing file after >")
			}
			file, e := os.Create(stage[i+1])
			if e != nil {
				return nil, nil, nil, e
			}
			stdout = file
			i++
		case "<":
			if i+1 >= len(stage) {
				return nil, nil, nil, errors.New("missing file after <")
			}
			file, e := os.Open(stage[i+1])
			if e != nil {
				return nil, nil, nil, e
			}
			stdin = file
			i++
		default:
			args = append(args, stage[i])
		}
	}
	return
}

func runPipeline(tokens []string) (int, error) {
	stages, err := buildStages(tokens)
	if err != nil {
		return 1, err
	}
	if len(stages) == 0 {
		return 0, nil
	}

	if len(stages) == 1 {
		args, stdinR, stdoutW, err := handleRedirections(stages[0])
		if err != nil {
			return 1, err
		}
		if len(args) > 0 && isInternal(args[0]) {
			if stdinR == nil {
				stdinR = os.Stdin
			}
			if stdoutW == nil {
				stdoutW = os.Stdout
			}
			code, err := executeInternal(args, stdinR, stdoutW, os.Stderr)
			if c, ok := stdoutW.(io.Closer); ok && c != os.Stdout {
				_ = c.Close()
			}
			return code, err
		}
	}

	var cmds []*exec.Cmd
	var prev io.Reader = os.Stdin
	var closers []io.Closer

	for i, st := range stages {
		args, stdinR, stdoutW, err := handleRedirections(st)
		if err != nil {
			return 1, err
		}
		if len(args) == 0 {
			return 1, errors.New("empty stage")
		}
		cmd := exec.Command(args[0], args[1:]...)

		if stdinR != nil {
			cmd.Stdin = stdinR
			if rc, ok := stdinR.(io.Closer); ok {
				closers = append(closers, rc)
			}
		} else {
			cmd.Stdin = prev
		}

		if i < len(stages)-1 {
			pipe, err := cmd.StdoutPipe()
			if err != nil {
				return 1, err
			}
			prev = pipe
		} else {
			if stdoutW != nil {
				cmd.Stdout = stdoutW
				if wc, ok := stdoutW.(io.Closer); ok {
					closers = append(closers, wc)
				}
			} else {
				cmd.Stdout = os.Stdout
			}
		}

		cmd.Stderr = os.Stderr
		cmds = append(cmds, cmd)
	}

	saveActive(cmds)
	defer clearActive()

	for _, c := range cmds {
		if err := c.Start(); err != nil {
			return 1, err
		}
	}

	var finalErr error
	for i := len(cmds) - 1; i >= 0; i-- {
		if err := cmds[i].Wait(); err != nil && finalErr == nil {
			finalErr = err
		}
	}

	for _, c := range closers {
		_ = c.Close()
	}

	if finalErr != nil {
		return 1, finalErr
	}
	return 0, nil
}

func executeLine(line string) int {
	tokens := tokenizeLine(line)
	if len(tokens) == 0 {
		return 0
	}

	var segments [][]string
	var operators []string
	current := []string{}

	for _, t := range tokens {
		if t == "&&" || t == "||" {
			if len(current) == 0 {
				return 1
			}
			segments = append(segments, current)
			operators = append(operators, t)
			current = nil
		} else {
			current = append(current, t)
		}
	}
	if len(current) > 0 {
		segments = append(segments, current)
	}

	status := 0
	for i, seg := range segments {
		if i > 0 {
			switch operators[i-1] {
			case "&&":
				if status != 0 {
					continue
				}
			case "||":
				if status == 0 {
					continue
				}
			}
		}
		status, _ = runPipeline(seg)
	}
	return status
}

func main() {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		for range signalCh {
			interruptAll()
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	buffer := make([]byte, 0, 64*1024)
	scanner.Buffer(buffer, 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		_ = executeLine(line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
