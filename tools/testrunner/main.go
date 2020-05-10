package main

// @./runtest.sh /dev/ttyACM0 115200 5.0s 1.0s

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"go.bug.st/serial"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Not enough arguments")
		os.Exit(1)
	}

	port := os.Args[1]
	speed, _ := strconv.Atoi(os.Args[2])
	//delayResponse, _ := strconv.Atoi(os.Args[3])

	// open serial port
	p, err := serial.Open(port, &serial.Mode{BaudRate: speed})
	if err != nil {
		fmt.Fprintf(os.Stderr, "serial open error: %v\n", err)
		os.Exit(1)
	}
	defer p.Close()

	// Reads up to 100 bytes
	buff := make([]byte, 100)
	result := ""
	for {
		n, err := p.Read(buff)
		if err != nil {
			fmt.Fprintf(os.Stderr, "serial read error: %v\n", err)
			os.Exit(1)
		}
		if n == 0 {
			fmt.Println("no serial data from device")
			os.Exit(1)
		}
		result = result + string(buff[:n])
		if strings.Contains(result, "begin running tests...") {
			break
		}
	}

	// send "t" to start tests
	p.Write([]byte("t"))

	// get test result
	result = ""
	for {
		// Reads up to 100 bytes
		n, err := p.Read(buff)
		if err != nil {
			fmt.Fprintf(os.Stderr, "serial read error: %v\n", err)
			os.Exit(1)
		}
		if n == 0 {
			fmt.Println("no serial data from device")
			os.Exit(1)
		}
		result = result + string(buff[:n])
		if strings.Contains(result, "Tests complete.") {
			fmt.Println(result)
			if strings.Contains(result, "error") {
				os.Exit(1)
			}
			os.Exit(0)
		}
	}
}
