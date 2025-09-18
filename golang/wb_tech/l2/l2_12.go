package l2

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type options struct {
	after       int
	before      int
	countOnly   bool
	ignoreCase  bool
	invert      bool
	fixed       bool
	lineNumbers bool
	pattern     string
	file        string
}

func parseFlags() options {
	cfg := options{}
	var context int

	flag.IntVar(&cfg.after, "A", 0, "print N strings after each match")
	flag.IntVar(&cfg.before, "B", 0, "print N strings before each match")
	flag.IntVar(&context, "C", 0, "print N lines before and after each match")
	flag.BoolVar(&cfg.countOnly, "c", false, "print only the count")
	flag.BoolVar(&cfg.ignoreCase, "i", false, "ignore case")
	flag.BoolVar(&cfg.invert, "v", false, "invert match")
	flag.BoolVar(&cfg.fixed, "F", false, "match fixed string")
	flag.BoolVar(&cfg.lineNumbers, "n", false, "print line numbers")
	flag.Parse()

	if context > 0 {
		cfg.after = context
		cfg.before = context
	}

	if flag.NArg() < 1 {
		fmt.Println("Usage: grep [flags] pattern [file]")
		os.Exit(1)
	}

	cfg.pattern = flag.Arg(0)
	if flag.NArg() >= 2 {
		cfg.file = flag.Arg(1)
	}

	return cfg
}

func readLines(file string) ([]string, error) {
	var lines []string
	var scanner *bufio.Scanner

	if file != "" {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		scanner = bufio.NewScanner(f)
	} else {
		scanner = bufio.NewScanner(os.Stdin)
	}

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func compilePattern(cfg options) (*regexp.Regexp, error) {
	if cfg.fixed {
		return nil, nil
	}
	pat := cfg.pattern
	if cfg.ignoreCase {
		pat = "(?i)" + pat
	}
	return regexp.Compile(pat)
}

func findMatches(lines []string, cfg options, re *regexp.Regexp) []int {
	matches := make([]int, 0)
	for i, line := range lines {
		text := line
		if cfg.ignoreCase && cfg.fixed {
			text = strings.ToLower(line)
		}

		match := false
		if cfg.fixed {
			match = strings.Contains(text, cfg.pattern)
		} else {
			match = re.MatchString(text)
		}
		if cfg.invert {
			match = !match
		}

		if match {
			matches = append(matches, i)
		}
	}
	return matches
}

func buildOutput(lines []string, matches []int, cfg options) map[int]struct{} {
	output := make(map[int]struct{})
	for _, idx := range matches {
		start := idx - cfg.before
		if start < 0 {
			start = 0
		}
		end := idx + cfg.after
		if end >= len(lines) {
			end = len(lines) - 1
		}
		for i := start; i <= end; i++ {
			output[i] = struct{}{}
		}
	}
	return output
}

func printOutput(lines []string, output map[int]struct{}, cfg options) {
	for i := 0; i < len(lines); i++ {
		if _, ok := output[i]; ok {
			if cfg.lineNumbers {
				fmt.Printf("%d:%s\n", i+1, lines[i])
			} else {
				fmt.Println(lines[i])
			}
		}
	}
}

func RunGrep() {
	cfg := parseFlags()
	lines, err := readLines(cfg.file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	re, err := compilePattern(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid regex pattern: %v\n", err)
		os.Exit(1)
	}

	matches := findMatches(lines, cfg, re)

	if cfg.countOnly {
		fmt.Println(len(matches))
		return
	}

	output := buildOutput(lines, matches, cfg)
	printOutput(lines, output, cfg)
}
