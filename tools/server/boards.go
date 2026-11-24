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
	image       string
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
			enabled:     false,
		},
		&Board{
			target:      "arduino",
			displayname: "Arduino Uno",
			image:       "arduino.svg",
			port:        "arduino_uno",
			baud:        57600,
			resetpause:  5 * time.Second,
			enabled:     true,
		},
		&Board{
			target:      "arduino-nano33",
			displayname: "Arduino Nano33 IoT",
			image:       "arduino-nano33.svg",
			port:        "arduino_nano33",
			baud:        115200,
			resetpause:  15 * time.Second,
			enabled:     true,
		},
		&Board{
			target:      "microbit-v2",
			displayname: "bbc:microbit",
			port:        "microbit",
			baud:        115200,
			resetpause:  9 * time.Second,
			enabled:     false,
		},
		&Board{
			target:      "hifive1b",
			displayname: "SiFive HiFive1 Rev.B",
			port:        "hifive1b",
			baud:        115200,
			resetpause:  30 * time.Second,
			enabled:     false,
		},
		&Board{
			target:      "circuitplay-express",
			displayname: "Adafruit Circuit Playground Express",
			image:       "circuit-playground.svg",
			port:        "circuitplay_express",
			baud:        115200,
			resetpause:  15 * time.Second,
			enabled:     false,
		},
		&Board{
			target:      "maixbit",
			displayname: "Sipeed MAix BiT",
			port:        "maixbit00",
			baud:        115200,
			resetpause:  10 * time.Second,
			enabled:     false,
		},
		&Board{
			target:      "itsybitsy-nrf52840",
			displayname: "Adafruit ItsyBitsy nRF52840",
			image:       "itsybitsy-nrf52840.svg",
			port:        "itsybitsy_nrf52840",
			baud:        115200,
			resetpause:  7 * time.Second,
			enabled:     true,
		},
		&Board{
			target:      "stm32f4disco-1",
			displayname: "STM32F407 Discovery",
			image:       "stm32f407-discovery.svg",
			port:        "stm32f4disco",
			baud:        115200,
			resetpause:  5 * time.Second,
			enabled:     true,
		},
		&Board{
			target:      "pico",
			displayname: "Raspberry Pi RP2040 Pico",
			image:       "pico.svg",
			port:        "pico",
			baud:        115200,
			resetpause:  10 * time.Second,
			enabled:     true,
		},
		&Board{
			target:      "xiao-esp32c3",
			displayname: "Seeed Studio Xiao ESP32-C3",
			image:       "xiao-esp32c3.svg",
			port:        "xiao_esp32c3",
			baud:        115200,
			resetpause:  10 * time.Second,
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

func buildTag(sha string) string {
	return "tinygohci:" + sha[:7]
}

func tinygoCacheVolume(sha string) string {
	return "tinygohci-cache-" + sha[:7]
}

// dockerRunArgs returns the docker run arguments shared by the compile and flash steps.
// The Go caches are kept between builds. The TinyGo cache is only kept for one SHA.
func (board *Board) dockerRunArgs(sha string) ([]string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return []string{"run", "--rm",
		"-v", pwd + ":/src",
		"-w", "/src/" + board.target,
		"-v", "tinygohci-gomod:/root/go/pkg/mod",
		"-v", "tinygohci-gocache:/root/.cache/go-build",
		"-v", tinygoCacheVolume(sha) + ":/root/.cache/tinygo",
	}, nil
}

// compile builds the test program without flashing, to fill the TinyGo cache.
func (board *Board) compile(sha string) (string, error) {
	args, err := board.dockerRunArgs(sha)
	if err != nil {
		return err.Error(), err
	}

	args = append(args, buildTag(sha),
		"tinygo", "build",
		"-o", "/tmp/"+board.target+".elf",
		"-target", board.target,
		".")
	out, err := exec.Command("docker", args...).CombinedOutput()
	return string(out), err
}

func (board *Board) flash(sha string) (string, error) {
	args, err := board.dockerRunArgs(sha)
	if err != nil {
		return err.Error(), err
	}

	realdev, err := os.Readlink("/dev/" + board.port)
	if err != nil {
		return err.Error(), err
	}

	args = append(args,
		fmt.Sprintf("--device=/dev/%s:/dev/%s:rwm", realdev, realdev),
		"-v", "/media:/media:shared",
		"-v", "/dev/bus/usb:/dev/bus/usb",
		"--device-cgroup-rule", "a 189:* rwm",
		buildTag(sha),
		"tinygo", "flash",
		"-size", "short",
		"-timeout", "30s",
		"-target", board.target,
		"-port=/dev/"+realdev,
		".")
	out, err := exec.Command("docker", args...).CombinedOutput()
	return string(out), err
}

func (board *Board) test() (string, error) {
	realdev, err := os.Readlink("/dev/" + board.port)
	if err != nil {
		return err.Error(), err
	}
	port := fmt.Sprintf("/dev/%s", realdev)
	br := strconv.Itoa(board.baud)

	out, err := exec.Command("./build/testrunner", port, br, "2").CombinedOutput()
	return string(out), err
}
