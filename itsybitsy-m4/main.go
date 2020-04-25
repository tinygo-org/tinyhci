package main

// Integration tests for ItsyBitsy-M4
//
// Wire up the pins, and run it while connected to the USB port.
//
// Digital read/write tests:
//	D12 <--> G
//	D11 <--> 3V
//	D10 <--> D9
//
// I2C tests:
// 	ItsyBitsy-M4 SCL <--> MPU-6050 SCL
// 	ItsyBitsy-M4 SDA <--> MPU-6050 SDA
// 	ItsyBitsy-M4 G <--> MPU-6050 GND
// 	ItsyBitsy-M4 D7 <--> MPU-6050 VCC
//
import (
	"machine"

	"time"

	"tinygo.org/x/drivers/mpu6050"
)

var (
	readV    = machine.D11
	readG    = machine.D12
	readpin  = machine.D9
	writepin = machine.D10
	powerpin = machine.D7

	serial = machine.UART0
	accel  *mpu6050.Device
)

func main() {
	serial.Configure(machine.UARTConfig{})
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
	time.Sleep(3 * time.Second)

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

// checks to see if an attached MPU-6050 accelerometer is connected.
func i2cConnection() {
	powerpin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	a := mpu6050.New(machine.I2C0)
	accel = &a

	print("i2cConnectionNoPower:")

	// should not be connected when not powered
	powerpin.Low()
	time.Sleep(1 * time.Second)
	if accel.Connected() {
		println(" fail")
	} else {
		println(" pass")
	}

	print("i2cConnectionPower:")
	// turn on power and should be connected now
	powerpin.High()
	time.Sleep(500 * time.Millisecond)

	accel.Configure()
	time.Sleep(500 * time.Millisecond)

	if !accel.Connected() {
		println(" fail")
		return
	}

	println(" pass")
}
