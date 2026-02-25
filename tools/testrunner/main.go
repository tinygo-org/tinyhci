package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"go.bug.st/serial"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Not enough arguments")
		os.Exit(1)
	}

	port := os.Args[1]
	speed, _ := strconv.Atoi(os.Args[2])

	var p serial.Port
	var err error
	for i := 0; i < 3; i++ {
		p, err = serial.Open(port, &serial.Mode{BaudRate: speed})
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "serial open error: %v\n", err)
		os.Exit(1)
	}
	defer p.Close()

	buff := make([]byte, 100)
	ch := make(chan string, 1)
	go func() {
		var lineBuf strings.Builder
		for {
			n, err := p.Read(buff)
			if err != nil {
				fmt.Fprintf(os.Stderr, "serial read error: %v\n", err)
				os.Exit(1)
			}
			data := string(buff[:n])
			lineBuf.WriteString(data)
			for {
				s := lineBuf.String()
				idx := strings.IndexByte(s, '\n')
				if idx == -1 {
					break
				}
				// Send the complete line (without the newline)
				ch <- s[:idx]
				// Remove the sent line from the buffer
				lineBuf.Reset()
				lineBuf.WriteString(s[idx+1:])
			}
		}
	}()

	// Wait for device prompt before sending "t"
	promptFound := false
	for !promptFound {
		select {
		case res := <-ch:
			if strings.Contains(res, "Press 't' key to begin running tests...") {
				p.Write([]byte("t"))
				promptFound = true
			}
		case <-time.After(5 * time.Second):
			fmt.Println("Timeout waiting for device prompt. trying anyhow...")
			p.Write([]byte("t"))
			promptFound = true
		}
	}

	var tapLines []string
	var planCount int
	var planSeen bool
	var result strings.Builder
	timeout := time.After(60 * time.Second)

	for {
		select {
		case res := <-ch:
			result.WriteString(res + "\n")
			lines := strings.Split(res, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				tapLines = append(tapLines, line)
				if !planSeen && strings.HasPrefix(line, "1..") {
					planSeen = true
					planCount, _ = strconv.Atoi(strings.TrimPrefix(line, "1.."))
				}
			}
			// Once we've seen the plan and enough test lines, break
			if planSeen && countTestLines(tapLines) >= planCount {
				goto PARSE
			}
		case <-timeout:
			fmt.Println("Timeout waiting for TAP output")
			os.Exit(1)
		}
	}

PARSE:
	fmt.Println(result.String())

	failed := false
	testLines := extractTestLines(tapLines)
	for _, line := range testLines {
		if strings.HasPrefix(line, "not ok") &&
			!strings.Contains(line, "# TODO") &&
			!strings.Contains(line, "# SKIP") {
			failed = true
			break
		}
	}

	if failed {
		os.Exit(1)
	}
	os.Exit(0)
}

// countTestLines returns the number of lines starting with "ok" or "not ok"
func countTestLines(lines []string) int {
	count := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "ok") || strings.HasPrefix(line, "not ok") {
			count++
		}
	}
	return count
}

// extractTestLines returns only the lines starting with "ok" or "not ok"
func extractTestLines(lines []string) []string {
	var tests []string
	for _, line := range lines {
		if strings.HasPrefix(line, "ok") || strings.HasPrefix(line, "not ok") {
			tests = append(tests, line)
		}
	}
	return tests
}
