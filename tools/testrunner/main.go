package main

// @./runtest.sh /dev/ttyACM0 115200 5.0s 1.0s

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"go.bug.st/serial"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Not enough arguments")
		os.Exit(1)
	}

	port := os.Args[1]
	speed, _ := strconv.Atoi(os.Args[2])

	// open serial port
	p, err := serial.Open(port, &serial.Mode{BaudRate: speed})
	if err != nil {
		fmt.Fprintf(os.Stderr, "serial open error: %v\n", err)
		os.Exit(1)
	}
	defer p.Close()

	// Reads up to 100 bytes
	buff := make([]byte, 100)
	ch := make(chan string, 1)
	go func() {
		for {
			n, err := p.Read(buff)
			if err != nil {
				fmt.Fprintf(os.Stderr, "serial read error: %v\n", err)
				os.Exit(1)
			}
			ch <- string(buff[:n])
		}
	}()

	result := ""
START:
	for {
		select {
		case res := <-ch:
			result = result + res
		case <-time.After(10 * time.Second):
			fmt.Println("no serial data from device yet. begin running tests anyhow...")
			break START
		}

		if strings.Contains(result, "begin running tests...") {
			break START
		}
	}

	// send "t" to start tests
	p.Write([]byte("t"))

	// get test result
	result = ""
	for {
		select {
		case res := <-ch:
			result = result + res
		case <-time.After(10 * time.Second):
			fmt.Println("no serial data from device")
			os.Exit(1)
		}

		if strings.Contains(result, "Tests complete.") {
			fmt.Println(result)
			if strings.Contains(result, "fail") {
				os.Exit(1)
			}
			os.Exit(0)
		}
	}
}
