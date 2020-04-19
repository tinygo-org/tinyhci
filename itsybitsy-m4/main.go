package main

// Integration tests for ItsyBitsy-M4
//
// Wire up the pins, and run it while connected to the USB port.
//
// Digital i/o tests:
//	D12 <--> G
//	D11 <--> 3V
//	D10 <--> D9
//
import (
	"machine"

	"time"
)

var (
	readV    = machine.D11
	readG    = machine.D12
	readpin  = machine.D9
	writepin = machine.D10
)

func main() {
	time.Sleep(3 * time.Second)

	println("Starting tests...")

	verifyDigitalReadV()
	verifyDigitalReadG()
	verifyDigitalWrite()

	println("Tests complete.")
}

// digital read of D11 pin physically connected to V
func verifyDigitalReadV() {
	print("verifyDigitalReadV:")

	readV.Configure(machine.PinConfig{Mode: machine.PinInput})

	// should be on
	if readV.Get() {
		println(" pass")
		return
	}

	println(" fail")
}

// digital read of D12 pin physically connected to G
func verifyDigitalReadG() {
	print("verifyDigitalReadG:")

	readG.Configure(machine.PinConfig{Mode: machine.PinInput})

	// should be off
	if readG.Get() {
		println(" fail")
		return
	}

	println(" pass")
}

// digital write on/off of D9 pin as input physically connected to D10 pin as output.
func verifyDigitalWrite() {
	readpin.Configure(machine.PinConfig{Mode: machine.PinInput})
	writepin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	print("verifyDigitalWrite On:")
	writepin.High()
	time.Sleep(10 * time.Millisecond)

	// should be on
	if readpin.Get() {
		println(" pass")
	} else {
		println(" fail")
	}

	print("verifyDigitalWrite Off:")
	writepin.Low()
	time.Sleep(10 * time.Millisecond)

	// should be off
	if readpin.Get() {
		println(" fail")
		return
	} else {
		println(" pass")
	}
}
