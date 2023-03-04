package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type Board struct {
	target      string
	displayname string
	port        string
	baud        int
	resetpause  time.Duration
	enabled     bool
}

var (
	boards = []*Board{
		&Board{
			target:      "itsybitsy-m4",
			displayname: "Adafruit ItsyBitsy-M4",
			port:        "itsybitsy_m4",
			baud:        115200,
			resetpause:  5 * time.Second,
			enabled:     true,
		},
		&Board{
			target:      "arduino",
			displayname: "Arduino Uno",
			port:        "arduino_uno",
			baud:        57600,
			resetpause:  5 * time.Second,
			enabled:     true,
		},
		&Board{
			target:      "arduino-nano33",
			displayname: "Arduino Nano33 IoT",
			port:        "arduino_nano33",
			baud:        115200,
			resetpause:  5 * time.Second,
			enabled:     true,
		},
		&Board{
			target:      "microbit",
			displayname: "bbc:microbit",
			port:        "microbit",
			baud:        115200,
			resetpause:  5 * time.Second,
			enabled:     true,
		},
		&Board{
			target:      "hifive1b",
			displayname: "SiFive HiFive1 Rev.B",
			port:        "hifive1b",
			baud:        115200,
			resetpause:  10 * time.Second,
			enabled:     true,
		},
		&Board{
			target:      "circuitplay-express",
			displayname: "Adafruit Circuit Playground Express",
			port:        "circuitplay_express",
			baud:        115200,
			resetpause:  5 * time.Second,
			enabled:     true,
		},
		&Board{
			target:      "maixbit",
			displayname: "Sipeed MAix BiT",
			port:        "maixbit00",
			baud:        115200,
			resetpause:  5 * time.Second,
			enabled:     true,
		},
		&Board{
			target:      "itsybitsy-nrf52840",
			displayname: "ItsyBitsy nRF52840",
			port:        "itsybitsy_nrf52840",
			baud:        115200,
			resetpause:  5 * time.Second,
			enabled:     true,
		},
		&Board{
			target:      "stm32f4disco-1",
			displayname: "STM32F407 Discovery",
			port:        "stm32f4disco",
			baud:        115200,
			resetpause:  5 * time.Second,
			enabled:     true,
		},
		&Board{
			target:      "pico",
			displayname: "RP2040 Pico",
			port:        "pico",
			baud:        115200,
			resetpause:  5 * time.Second,
			enabled:     true,
		},
	}
)

// GetBoard returns the board for this target.
func GetBoard(target string) *Board {
	for _, b := range boards {
		if b.target == target {
			return b
		}
	}
	return nil
}

func (board *Board) flash(sha string) (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return err.Error(), err
	}
	buildtag := fmt.Sprintf("tinygohci:%s", sha[:7])
	device := fmt.Sprintf("--device=/dev/%s:/dev/%s:rwm", board.port, board.port)
	port := fmt.Sprintf("-port=/dev/%s", board.port)
	workdir := fmt.Sprintf("/src/%s", board.target)
	out, err := exec.Command("docker", "run",
		device,
		"-v", "/media:/media:shared",
		"-v", pwd+":/src",
		"-w", workdir,
		"-v", "/dev/bus/usb:/dev/bus/usb",
		"--device-cgroup-rule", "a 189:* rwm",
		"--rm",
		buildtag,
		"tinygo", "flash",
		"-size", "short",
		"-timeout", "30s",
		"-target", board.target,
		port,
		".").CombinedOutput()
	return string(out), err
}

func (board *Board) test() (string, error) {
	port := fmt.Sprintf("/dev/%s", board.port)
	br := strconv.Itoa(board.baud)

	out, err := exec.Command("./build/testrunner", port, br, "5").CombinedOutput()
	return string(out), err
}
