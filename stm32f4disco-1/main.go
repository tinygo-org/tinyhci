package main

// Integration tests for STM32F407 Discovery
//
// Wire up the pins, and run it while connected to the USB port
// as well as connect a UART/USB adaptor to pins PA2/PA3.
//
// Digital read/write tests:
//	PD3 <--> G
//	PD4 <--> 3V
//	PD1 <--> PD2
//
// I2C tests:
// 	STM32F407 SCL (PB6) <--> MPU-6050 SCL
// 	STM32F407 SDA (PB9) <--> MPU-6050 SDA
// 	STM32F407 G <--> MPU-6050 GND
// 	STM32F407 3V <--> MPU-6050 VCC
//
import (
	"machine"

	"time"

	"tinygo.org/x/drivers/mpu6050"
)

var (
	// used by digital tests
	readV    = machine.PD4
	readG    = machine.PD3
	readpin  = machine.PD2
	writepin = machine.PD1

	// used by i2c tests
	accel *mpu6050.Device

	serial = machine.UART0
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})

	waitForStart()

	digitalReadVoltage()
	digitalReadGround()
	digitalWrite()
	i2cConnection()

	endTests()
}

// wait for keypress on serial port to start test suite.
func waitForStart() {
	time.Sleep(5 * time.Second)

	println("=== TINYGO INTEGRATION TESTS ===")
	println("Press 't' key to begin running tests...")

	for {
		if serial.Buffered() > 0 {
			data, _ := serial.ReadByte()

			if data != 't' {
				time.Sleep(100 * time.Millisecond)
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

// digital read of a pin physically connected to V
func digitalReadVoltage() {
	printtest("digitalReadVoltage")

	readV.Configure(machine.PinConfig{Mode: machine.PinInput})
	time.Sleep(100 * time.Millisecond)

	// should be on
	if readV.Get() {
		printtestresult("pass")
		return
	}

	printtestresult("fail")
}

// digital read of a pin physically connected to G
func digitalReadGround() {
	printtest("digitalReadGround")

	readG.Configure(machine.PinConfig{Mode: machine.PinInput})
	time.Sleep(100 * time.Millisecond)

	// should be off
	if readG.Get() {
		printtestresult("fail")
		return
	}

	printtestresult("pass")
}

// digital write on/off of one pin as input physically connected to a different pin as output.
func digitalWrite() {
	readpin.Configure(machine.PinConfig{Mode: machine.PinInput})
	writepin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	time.Sleep(100 * time.Millisecond)

	printtest("digitalWriteOn")
	writepin.High()
	time.Sleep(100 * time.Millisecond)

	// should be on
	if readpin.Get() {
		printtestresult("pass")
	} else {
		printtestresult("fail")
	}

	time.Sleep(100 * time.Millisecond)

	printtest("digitalWriteOff")
	writepin.Low()
	time.Sleep(100 * time.Millisecond)

	// should be off
	if readpin.Get() {
		printtestresult("fail")
		return
	} else {
		printtestresult("pass")
	}
}

// checks to see if an attached MPU-6050 accelerometer is connected.
func i2cConnection() {
	a := mpu6050.New(machine.I2C0)
	accel = &a
	printtest("i2cConnection")

	accel.Configure()
	time.Sleep(100 * time.Millisecond)

	if accel.Connected() {
		printtestresult("pass")
		return
	}

	printtestresult("fail")
	return
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
