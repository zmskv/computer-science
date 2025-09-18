package l2

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func parsingFields(fieldsArg string) map[int]struct{} {
	fields := make(map[int]struct{})
	parts := strings.Split(fieldsArg, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "-") {
			rangeParts := strings.SplitN(part, "-", 2)
			if len(rangeParts) != 2 {
				continue
			}
			start, err1 := strconv.Atoi(rangeParts[0])
			end, err2 := strconv.Atoi(rangeParts[1])
			if err1 != nil || err2 != nil || start > end || start < 1 {
				continue
			}
			for i := start; i <= end; i++ {
				fields[i-1] = struct{}{}
			}
		} else {
			idx, err := strconv.Atoi(part)
			if err != nil || idx < 1 {
				continue
			}
			fields[idx-1] = struct{}{}
		}
	}
	return fields
}

func processLine(line, delimiter string, fields map[int]struct{}) string {
	cols := strings.Split(line, delimiter)
	var selected []string
	for i := 0; i < len(cols); i++ {
		if _, ok := fields[i]; ok {
			selected = append(selected, cols[i])
		}
	}
	return strings.Join(selected, delimiter)
}

func RunCut() {
	fieldsArg := flag.String("f", "", "fields to extract (comma-separated, ranges allowed)")
	delimiter := flag.String("d", "\t", "field delimiter")
	separated := flag.Bool("s", false, "only lines containing delimiter")
	flag.Parse()

	if *fieldsArg == "" {
		fmt.Fprintln(os.Stderr, "Error: -f flag is required")
		os.Exit(1)
	}

	fields := parsingFields(*fieldsArg)
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Text()
		if *separated && !strings.Contains(line, *delimiter) {
			continue
		}
		fmt.Println(processLine(line, *delimiter, fields))
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input:", err)
		os.Exit(1)
	}
}
