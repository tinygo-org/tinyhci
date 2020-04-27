package main

// Integration tests for Arduino Uno
//
// Wire up the pins, and run it while connected to the USB port.
//
// Digital read/write tests:
//	D12 <--> G
//	D11 <--> 5V
//	D10 <--> D9
//
// Analog read tests:
//	A0 <--> 5V
//	A1 <--> 5V/2 (use voltage divider)
//	A2 <--> G
//
// I2C tests:
// 	Arduino A5 <--> MPU-6050 SCL
// 	Arduino A4 <--> MPU-6050 SDA
// 	Arduino G <--> MPU-6050 GND
// 	Arduino 5V <--> MPU-6050 VCC
//
import (
	"machine"

	"time"

	"tinygo.org/x/drivers/mpu6050"
)

var (
	// used by digital tests
	readV    machine.Pin = 11
	readG    machine.Pin = 12
	readpin  machine.Pin = 9
	writepin machine.Pin = 10

	// used by analog tests
	analogV    = machine.ADC{machine.ADC0}
	analogHalf = machine.ADC{machine.ADC1}
	analogG    = machine.ADC{machine.ADC2}

	// used by i2c tests
	accel    *mpu6050.Device
	powerpin machine.Pin = 7

	serial = machine.UART0
)

const (
	maxanalog       = 65535
	allowedvariance = 256
)

func main() {
	serial.Configure(machine.UARTConfig{BaudRate: 57600})
	machine.I2C0.Configure(machine.I2CConfig{})
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
	println("Tests complete.")

	// tests done, now sleep waiting for baud reset to load new code
	for {
		time.Sleep(1 * time.Second)
	}
}

// digital read of D11 pin physically connected to V
func digitalReadVoltage() {
	print("digitalReadVoltage:")

	readV.Configure(machine.PinConfig{Mode: machine.PinInput})

	// should be on
	if readV.Get() {
		println(" pass")
		return
	}

	println(" fail")
}

// digital read of D12 pin physically connected to G
func digitalReadGround() {
	print("digitalReadGround:")

	readG.Configure(machine.PinConfig{Mode: machine.PinInput})

	// should be off
	if readG.Get() {
		println(" fail")
		return
	}

	println(" pass")
}

// digital write on/off of D9 pin as input physically connected to D10 pin as output.
func digitalWrite() {
	readpin.Configure(machine.PinConfig{Mode: machine.PinInput})
	writepin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	print("digitalWriteOn:")
	writepin.High()
	time.Sleep(100 * time.Millisecond)

	// should be on
	if readpin.Get() {
		println(" pass")
	} else {
		println(" fail")
	}

	time.Sleep(100 * time.Millisecond)

	print("digitalWriteOff:")
	writepin.Low()
	time.Sleep(100 * time.Millisecond)

	// should be off
	if readpin.Get() {
		println(" fail")
		return
	} else {
		println(" pass")
	}
}

// analog read of pin connected to supply voltage.
func analogReadVoltage() {
	analogV.Configure()

	print("analogReadVoltage:")

	// should be close to max
	val := analogV.Get()
	if val >= maxanalog-allowedvariance {
		println(" pass")

		return
	} else {
		println(" fail")
		print("  expected: ")
		print("'val >= 65535-256'")
		print(", actual: ")
		println(val)
	}
}

// analog read of pin connected to ground.
func analogReadGround() {
	analogG.Configure()

	print("analogReadGround:")

	// should be close to zero
	val := analogG.Get()
	if val <= allowedvariance {
		println(" pass")
		return
	} else {
		println(" fail")

		print("  expected: ")
		print("'val <= 256'")
		print(", actual: ")
		println(val)
	}
}

// analog read of pin connected to supply voltage that has been divided by 2
// using resistors.
func analogReadHalfVoltage() {
	analogHalf.Configure()

	print("analogReadHalfVoltage:")

	// should be around half the max
	val := analogHalf.Get()
	if val <= maxanalog/2+allowedvariance && val >= maxanalog/2-allowedvariance {
		println(" pass")
		return
	}
	println(" fail")

	print("  expected: ")
	print("'val <= 65535/2+256 && val >= 65535/2-256'")
	print(", actual: ")
	println(val)
}

// checks to see if an attached MPU-6050 accelerometer is connected.
func i2cConnection() {
	powerpin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	a := mpu6050.New(machine.I2C0)
	accel = &a

	print("i2cConnection:")
	accel.Configure()

	if !accel.Connected() {
		println(" fail")
		return
	}

	println(" pass")
}
