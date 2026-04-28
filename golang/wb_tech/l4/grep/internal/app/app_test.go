package app

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"mygrep/internal/distributed"
)

func TestRunDistributedMatchesGNUgrep(t *testing.T) {
	grepBinary := findSystemGrep(t)

	s1 := httptest.NewServer(distributed.NewServer(2).Handler())
	defer s1.Close()
	s2 := httptest.NewServer(distributed.NewServer(2).Handler())
	defer s2.Close()
	s3 := httptest.NewServer(distributed.NewServer(2).Handler())
	defer s3.Close()

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "app.log")
	data := strings.Join([]string{
		"info: boot",
		"error: disk full",
		"warn: retry",
		"error: network timeout",
		"done",
	}, "\n") + "\n"

	if err := os.WriteFile(inputFile, []byte(data), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run(context.Background(), []string{
		"-pattern", "error",
		"-F",
		"-n",
		"-peers", strings.Join([]string{s1.URL, s2.URL, s3.URL}, ","),
		"-quorum", "2",
		"-shard-size", "2",
		inputFile,
	}, bytes.NewBuffer(nil), &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, stderr = %s", exitCode, stderr.String())
	}

	cmd := exec.Command(grepBinary, "-F", "-n", "error", inputFile)
	expected, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("grep command failed: %v, output: %s", err, string(expected))
	}

	if normalizeNewlines(stdout.String()) != normalizeNewlines(string(expected)) {
		t.Fatalf("distributed output = %q, want %q", stdout.String(), string(expected))
	}
}

func TestRunDistributedReachesQuorumDespiteOneFailedPeer(t *testing.T) {
	s1 := httptest.NewServer(distributed.NewServer(2).Handler())
	defer s1.Close()
	s2 := httptest.NewServer(distributed.NewServer(2).Handler())
	defer s2.Close()
	s3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer s3.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	input := bytes.NewBufferString("alpha\nerror\nbeta\n")
	exitCode := Run(context.Background(), []string{
		"-pattern", "error",
		"-F",
		"-peers", strings.Join([]string{s1.URL, s2.URL, s3.URL}, ","),
		"-quorum", "2",
		"-shard-size", "1",
		"-timeout", (200 * time.Millisecond).String(),
	}, input, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, stderr = %s", exitCode, stderr.String())
	}

	if got := stdout.String(); got != "error\n" {
		t.Fatalf("Run() output = %q, want %q", got, "error\n")
	}
}

func findSystemGrep(t *testing.T) string {
	t.Helper()

	if path, err := exec.LookPath("grep"); err == nil {
		return path
	}

	candidates := []string{
		`C:\Program Files\Git\usr\bin\grep.exe`,
		`C:\Program Files (x86)\Git\usr\bin\grep.exe`,
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	t.Skip("system grep binary not found")
	return ""
}

func normalizeNewlines(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}
