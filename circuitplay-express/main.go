package main

// Integration tests for Circuit Playground Express
//
// Wire up the pins, and run it while connected to the USB port.
//
// Digital read/write tests:
//	A0 <--> G
//	A1 <--> 3V
//	A4 <--> A5
//
// Analog read tests:
//	A1 <--> 3.3V
//	A3 <--> 3.3V/2 (use voltage divider)
//	A0 <--> G
//
// I2C tests:
// 	Uses the built in lis3dh device
//
import (
	"machine"
	"strconv"

	"time"

	"tinygo.org/x/drivers/lis3dh"
)

var (
	// used by digital tests
	readV    = machine.A1
	readG    = machine.A0
	readpin  = machine.A5
	writepin = machine.A4

	// used by analog tests
	analogV    = machine.ADC{machine.A1}
	analogHalf = machine.ADC{machine.A3}
	analogG    = machine.ADC{machine.A0}

	// used by i2c tests
	accel *lis3dh.Device
)

const (
	maxanalog       = 65535
	allowedvariance = 4096
)

func main() {
	machine.Serial.Configure(machine.UARTConfig{})
	machine.I2C1.Configure(machine.I2CConfig{SCL: machine.SCL1_PIN, SDA: machine.SDA1_PIN})
	machine.InitADC()

	waitForStart()

	digitalReadVoltage()
	digitalReadGround()
	digitalWrite()
	analogReadVoltage()
	analogReadGround()
	analogReadHalfVoltage()
	i2cConnection()

	endTests()
}

// wait for keypress on serial port to start test suite.
func waitForStart() {
	time.Sleep(3 * time.Second)

	println("=== TINYGO INTEGRATION TESTS ===")
	println("Press 't' key to begin running tests...")

	for {
		if machine.Serial.Buffered() > 0 {
			data, _ := machine.Serial.ReadByte()

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
func analogReadVoltage() {
	analogV.Configure(machine.ADCConfig{})
	time.Sleep(200 * time.Millisecond)

	printtest("analogReadVoltage")

	// should be close to max
	var avg int
	for i := 0; i < 10; i++ {
		v := analogV.Get()
		avg += int(v)
		time.Sleep(10 * time.Millisecond)
	}
	avg /= 10
	val := uint16(avg)

	if val >= maxanalog-allowedvariance {
		printtestresult("pass")

		return
	} else {
		printtestresult("fail")
		printfailexpected("'val >= 65535-" + strconv.Itoa(allowedvariance) + "'")
		printfailactual(val)
	}
}

// analog read of pin connected to ground.
func analogReadGround() {
	analogG.Configure(machine.ADCConfig{})
	time.Sleep(500 * time.Millisecond)

	printtest("analogReadGround")

	// should be close to zero
	var avg int
	for i := 0; i < 10; i++ {
		v := analogG.Get()
		avg += int(v)
		time.Sleep(10 * time.Millisecond)
	}
	avg /= 10
	val := uint16(avg)

	if val <= allowedvariance {
		printtestresult("pass")
		return
	} else {
		printtestresult("fail")
		printfailexpected("'val <= " + strconv.Itoa(allowedvariance) + "'")
		printfailactual(val)
	}
}

// analog read of pin connected to supply voltage that has been divided by 2
// using resistors.
func analogReadHalfVoltage() {
	analogHalf.Configure(machine.ADCConfig{})
	time.Sleep(200 * time.Millisecond)

	printtest("analogReadHalfVoltage")

	// should be around half the max
	var avg int
	for i := 0; i < 10; i++ {
		v := analogHalf.Get()
		avg += int(v)
		time.Sleep(10 * time.Millisecond)
	}
	avg /= 10
	val := uint16(avg)

	if val <= maxanalog/2+allowedvariance && val >= maxanalog/2-allowedvariance {
		printtestresult("pass")
		return
	}

	printtestresult("fail")
	printfailexpected("'val <= 65535/2+" + strconv.Itoa(allowedvariance) + " && val >= 65535/2-" + strconv.Itoa(allowedvariance) + "'")
	printfailactual(val)
}

// checks to see if an attached lis3dh accelerometer is connected.
func i2cConnection() {
	a := lis3dh.New(machine.I2C1)
	accel = &a
	accel.Address = lis3dh.Address1 // address on the Circuit Playground Express

	printtest("i2cConnection")
	accel.Configure()
	time.Sleep(400 * time.Millisecond)

	if !accel.Connected() {
		printtestresult("fail")
		return
	}

	printtestresult("pass")
}

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
