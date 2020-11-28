package main

// Integration tests for ESP32 mini32
//
// Wire up the pins, and run it while connected to the USB port.
//
// Digital read/write tests:
//	IO27 <--> G
//	IO26 <--> 3V
//	IO25 <--> IO32
//
// Analog read tests:
//	A1 <--> 3.3V
//	A3 <--> 3.3V/2 (use voltage divider)
//	A0 <--> G
//
// I2C tests:
// 	Not yet
//
import (
	"machine"

	"time"
)

var (
	// used by digital tests
	readV    = machine.IO26
	readG    = machine.IO27
	readpin  = machine.IO25
	writepin = machine.IO32

	// used by analog tests
	// analogV    = machine.ADC{machine.A1}
	// analogHalf = machine.ADC{machine.A3}
	// analogG    = machine.ADC{machine.A0}

	// used by i2c tests...soon.

	serial = machine.UART0
)

const (
	maxanalog       = 65535
	allowedvariance = 4096
)

func main() {
	serial.Configure(machine.UARTConfig{})
	//machine.InitADC()

	waitForStart()

	digitalReadVoltage()
	digitalReadGround()
	digitalWrite()
	// analogReadVoltage()
	// analogReadGround()
	// analogReadHalfVoltage()

	endTests()
}

// wait for keypress on serial port to start test suite.
func waitForStart() {
	time.Sleep(3 * time.Second)

	println("=== TINYGO INTEGRATION TESTS ===")
	println("Press 't' key to begin running tests...")

	for {
		if serial.Buffered() > 0 {
			data, _ := serial.ReadByte()

			if data != 't' {
				time.Sleep(200 * time.Millisecond)
			}
			return
		}
	}
}

func endTests() {
	println("\n### Tests complete.")

	// tests done, now sleep waiting for baud reset to load new code
	for {
		time.Sleep(1 * time.Second)
	}
}

// digital read of pin physically connected to V
func digitalReadVoltage() {
	printtest("digitalReadVoltage")

	readV.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	time.Sleep(200 * time.Millisecond)

	// should be on
	if readV.Get() {
		printtestresult("pass")
		return
	}

	printtestresult("fail")
}

// digital read of pin physically connected to G
func digitalReadGround() {
	printtest("digitalReadGround")

	readG.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	time.Sleep(200 * time.Millisecond)

	// should be off
	if readG.Get() {
		printtestresult("fail")
		return
	}

	printtestresult("pass")
}

// digital write on/off of pin as input physically connected to different pin as output.
func digitalWrite() {
	readpin.Configure(machine.PinConfig{Mode: machine.PinInput})
	time.Sleep(200 * time.Millisecond)

	writepin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	time.Sleep(200 * time.Millisecond)

	printtest("digitalWriteOff")
	writepin.Low()
	time.Sleep(200 * time.Millisecond)

	// should be off
	if readpin.Get() {
		printtestresult("fail")
		return
	} else {
		printtestresult("pass")
	}

	time.Sleep(200 * time.Millisecond)

	printtest("digitalWriteOn")
	writepin.High()
	time.Sleep(200 * time.Millisecond)

	// should be on
	if readpin.Get() {
		printtestresult("pass")
	} else {
		printtestresult("fail")
	}
}

// analog read of pin connected to supply voltage.
// func analogReadVoltage() {
// 	analogV.Configure()
// 	time.Sleep(200 * time.Millisecond)

// 	printtest("analogReadVoltage")

// 	// should be close to max
// 	var avg int
// 	for i := 0; i < 10; i++ {
// 		v := analogV.Get()
// 		avg += int(v)
// 		time.Sleep(10 * time.Millisecond)
// 	}
// 	avg /= 10
// 	val := uint16(avg)

// 	if val >= maxanalog-allowedvariance {
// 		printtestresult("pass")

// 		return
// 	} else {
// 		printtestresult("fail")
// 		printfailexpected("'val >= 65535-" + strconv.Itoa(allowedvariance) + "'")
// 		printfailactual(val)
// 	}
// }

// analog read of pin connected to ground.
// func analogReadGround() {
// 	analogG.Configure()
// 	time.Sleep(500 * time.Millisecond)

// 	printtest("analogReadGround")

// 	// should be close to zero
// 	var avg int
// 	for i := 0; i < 10; i++ {
// 		v := analogG.Get()
// 		avg += int(v)
// 		time.Sleep(10 * time.Millisecond)
// 	}
// 	avg /= 10
// 	val := uint16(avg)

// 	if val <= allowedvariance {
// 		printtestresult("pass")
// 		return
// 	} else {
// 		printtestresult("fail")
// 		printfailexpected("'val <= " + strconv.Itoa(allowedvariance) + "'")
// 		printfailactual(val)
// 	}
// }

// analog read of pin connected to supply voltage that has been divided by 2
// using resistors.
// func analogReadHalfVoltage() {
// 	analogHalf.Configure()
// 	time.Sleep(200 * time.Millisecond)

// 	printtest("analogReadHalfVoltage")

// 	// should be around half the max
// 	var avg int
// 	for i := 0; i < 10; i++ {
// 		v := analogHalf.Get()
// 		avg += int(v)
// 		time.Sleep(10 * time.Millisecond)
// 	}
// 	avg /= 10
// 	val := uint16(avg)

// 	if val <= maxanalog/2+allowedvariance && val >= maxanalog/2-allowedvariance {
// 		printtestresult("pass")
// 		return
// 	}

// 	printtestresult("fail")
// 	printfailexpected("'val <= 65535/2+" + strconv.Itoa(allowedvariance) + " && val >= 65535/2-" + strconv.Itoa(allowedvariance) + "'")
// 	printfailactual(val)
// }

func printtest(testname string) {
	print("- " + testname + " = ")
}

func printtestresult(result string) {
	println("***" + result + "***")
}

func printfailexpected(reason string) {
	println("        expected:", reason)
}

func printfailactual(val uint16) {
	println("        actual:", val)
}
