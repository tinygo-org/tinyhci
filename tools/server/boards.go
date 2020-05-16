package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

type Board struct {
	target      string
	displayname string
	port        string
	baud        int
}

var (
	boards = []Board{
		Board{
			target:      "itsybitsy-m4",
			displayname: "Adafruit ItsyBitsy-M4",
			port:        "itsybitsy_m4",
			baud:        115200,
		},
		Board{
			target:      "arduino",
			displayname: "Arduino Uno",
			port:        "arduino_uno",
			baud:        57600,
		},
		Board{
			target:      "arduino-nano33",
			displayname: "Arduino Nano33 IoT",
			port:        "arduino_nano33",
			baud:        115200,
		},
	}
)

func (board Board) flash() (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return err.Error(), err
	}
	device := fmt.Sprintf("--device=/dev/%s", board.port)
	port := fmt.Sprintf("-port=/dev/%s", board.port)
	file := fmt.Sprintf("/src/%s/main.go", board.target)
	out, err := exec.Command("docker", "run",
		device,
		"-v", "/media:/media:shared",
		"-v", pwd+":/src",
		"tinygohci:latest",
		"tinygo", "flash",
		"-target", board.target, port, file).CombinedOutput()
	return string(out), err
}

func (board Board) test() (string, error) {
	port := fmt.Sprintf("/dev/%s", board.port)
	br := strconv.Itoa(board.baud)

	out, err := exec.Command("./build/testrunner", port, br, "5").CombinedOutput()
	return string(out), err
}
