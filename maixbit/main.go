package main

// Integration tests for Sipeed MAix Bit
//
// Wire up the pins, and run it while connected to the USB port.
//
// Digital read/write tests (GPIO):
//	D8  <--> G
//	D9  <--> 3V
//	D10 <--> D15
//
// Digital read/write tests (GPIOHS):
//	D21 <--> G
//	D17 <--> 3V
//	D18 <--> D19
//
// I2C tests:
// 	MAix Bit SCL (D35) <--> TMP102 SCL
// 	MAix Bit SDA (D34) <--> TMP102 SDA
// 	MAix Bit G         <--> TMP102 GND
// 	MAix Bit D33       <--> TMP102 VCC
//
import (
	"machine"

	"time"

	"tinygo.org/x/drivers/mpu6050"
)

var (
	// used by digital GPIO tests
	readV    = machine.D9
	readG    = machine.D8
	readpin  = machine.D10
	writepin = machine.D15

	// used by digital GPIO tests
	readVGPIOHS    = machine.D17
	readGGPIOHS    = machine.D21
	readpinGPIOHS  = machine.D18
	writepinGPIOHS = machine.D19

	// used by i2c tests
	accel    *mpu6050.Device
	powerpin = machine.D33
)

func main() {
	machine.Serial.Configure(machine.UARTConfig{})
	machine.I2C0.Configure(machine.I2CConfig{})

	waitForStart()

	digitalReadVoltageGPIO()
	digitalReadGroundGPIO()
	digitalWriteGPIO()
	digitalReadVoltageGPIOHS()
	digitalReadGroundGPIOHS()
	digitalWriteGPIOHS()
	i2cConnection()

	endTests()
}

// wait for keypress on serial port to start test suite.
func waitForStart() {
	time.Sleep(5 * time.Second)

	println("=== TINYGO INTEGRATION TESTS ===")
	println("Press 't' key to begin running tests...")

	for {
		if machine.Serial.Buffered() > 0 {
			data, _ := machine.Serial.ReadByte()

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

// digital read of a GPIO pin physically connected to V
func digitalReadVoltageGPIO() {
	printtest("digitalReadVoltage (GPIO)")

	readV.Configure(machine.PinConfig{Mode: machine.PinInputPullDown})
	time.Sleep(100 * time.Millisecond)

	// should be on
	if readV.Get() {
		printtestresult("pass")
		return
	}

	printtestresult("fail")
}

// digital read of a GPIO pin physically connected to G
func digitalReadGroundGPIO() {
	printtest("digitalReadGround (GPIO)")

	readG.Configure(machine.PinConfig{Mode: machine.PinInputPullUp})
	time.Sleep(100 * time.Millisecond)

	// should be off
	if readG.Get() {
		printtestresult("fail")
		return
	}

	printtestresult("pass")
}

// digital write on/off of one GPIO pin as input physically connected to a different GPIO pin as output.
func digitalWriteGPIO() {
	readpin.Configure(machine.PinConfig{Mode: machine.PinInputPullDown})
	writepin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	time.Sleep(100 * time.Millisecond)

	printtest("digitalWriteOn (GPIO)")
	writepin.High()
	time.Sleep(100 * time.Millisecond)

	// should be on
	if readpin.Get() {
		printtestresult("pass")
	} else {
		printtestresult("fail")
	}

	time.Sleep(100 * time.Millisecond)

	printtest("digitalWriteOff (GPIO)")
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

// digital read of a GPIOHS pin physically connected to V
func digitalReadVoltageGPIOHS() {
	printtest("digitalReadVoltage (GPIOHS)")

	readVGPIOHS.Configure(machine.PinConfig{Mode: machine.PinInputPullDown})
	time.Sleep(100 * time.Millisecond)

	// should be on
	if readVGPIOHS.Get() {
		printtestresult("pass")
		return
	}

	printtestresult("fail")
}

// digital read of a GPIOHS pin physically connected to G
func digitalReadGroundGPIOHS() {
	printtest("digitalReadGround (GPIOHS)")

	readGGPIOHS.Configure(machine.PinConfig{Mode: machine.PinInputPullUp})
	time.Sleep(100 * time.Millisecond)

	// should be off
	if readGGPIOHS.Get() {
		printtestresult("fail")
		return
	}

	printtestresult("pass")
}

// digital write on/off of one GPIOHS pin as input physically connected to a different GPIOHS pin as output.
func digitalWriteGPIOHS() {
	readpinGPIOHS.Configure(machine.PinConfig{Mode: machine.PinInputPullDown})
	writepinGPIOHS.Configure(machine.PinConfig{Mode: machine.PinOutput})
	time.Sleep(100 * time.Millisecond)

	printtest("digitalWriteOn (GPIOHS)")
	writepinGPIOHS.High()
	time.Sleep(100 * time.Millisecond)

	// should be on
	if readpinGPIOHS.Get() {
		printtestresult("pass")
	} else {
		printtestresult("fail")
	}

	time.Sleep(100 * time.Millisecond)

	printtest("digitalWriteOff (GPIOHS)")
	writepinGPIOHS.Low()
	time.Sleep(100 * time.Millisecond)

	// should be off
	if readpinGPIOHS.Get() {
		printtestresult("fail")
		return
	} else {
		printtestresult("pass")
	}
}

// checks to see if an attached TMP102 thermometer is connected.
func i2cConnection() {
	powerpin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	time.Sleep(100 * time.Millisecond)

	a := mpu6050.New(machine.I2C0)
	accel = &a

	printtest("i2cConnection (MPU6050)")

	// have to recycle power
	powerpin.Low()
	time.Sleep(500 * time.Millisecond)

	// turn on power and should be connected now
	powerpin.High()
	time.Sleep(500 * time.Millisecond)

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
