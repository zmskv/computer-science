package app

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"mygrep/internal/distributed"
	"mygrep/internal/input"
	"mygrep/internal/search"
)

func Run(ctx context.Context, args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("mygrep", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		serve      = fs.Bool("serve", false, "run as a worker node")
		listen     = fs.String("listen", ":8080", "worker listen address")
		workers    = fs.Int("workers", runtime.NumCPU(), "number of worker goroutines")
		pattern    = fs.String("pattern", "", "search pattern")
		fixed      = fs.Bool("F", false, "interpret pattern as a fixed string")
		ignoreCase = fs.Bool("i", false, "ignore case distinctions")
		lineNumber = fs.Bool("n", false, "print line numbers")
		invert     = fs.Bool("v", false, "invert the sense of matching")
		peersFlag  = fs.String("peers", "", "comma-separated worker URLs")
		quorum     = fs.Int("quorum", 0, "required number of identical worker responses per shard")
		shardSize  = fs.Int("shard-size", 128, "number of lines per shard")
		timeout    = fs.Duration("timeout", 3*time.Second, "per-request timeout")
	)

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *serve {
		if err := distributed.Serve(ctx, *listen, *workers); err != nil {
			_, _ = fmt.Fprintf(stderr, "serve: %v\n", err)
			return 1
		}
		return 0
	}

	if *pattern == "" {
		_, _ = fmt.Fprintln(stderr, "pattern is required")
		return 2
	}

	lines, err := input.ReadAll(stdin, fs.Args())
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "input: %v\n", err)
		return 1
	}

	opts := search.Options{
		Pattern:    *pattern,
		Fixed:      *fixed,
		IgnoreCase: *ignoreCase,
		Invert:     *invert,
		LineNumber: *lineNumber,
	}

	peers := splitPeers(*peersFlag)
	matches, err := distributed.Run(ctx, lines, opts, distributed.CoordinatorConfig{
		Peers:       peers,
		Quorum:      *quorum,
		ShardSize:   *shardSize,
		Parallelism: *workers,
		Timeout:     *timeout,
	})
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "search: %v\n", err)
		return 1
	}

	includeSource := len(fs.Args()) > 1
	_, _ = io.WriteString(stdout, search.Format(matches, includeSource, opts.LineNumber))
	return 0
}

func splitPeers(raw string) []string {
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	peers := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			peers = append(peers, part)
		}
	}
	return peers
}

func RunMain() {
	os.Exit(Run(context.Background(), os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
