package main

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
