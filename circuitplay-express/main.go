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

	"time"

	"tinygo.org/x/drivers/lis3dh"
	"tinygo.org/x/tap"
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

	t := tap.New()
	t.Header(7)

	if digitalReadVoltage() {
		t.Pass("digitalReadVoltage (GPIO)")
	} else {
		t.Fail("digitalReadVoltage (GPIO)")
	}

	if digitalReadGround() {
		t.Pass("digitalReadGround (GPIO)")
	} else {
		t.Fail("digitalReadGround (GPIO)")
	}

	if digitalWrite() {
		t.Pass("digitalWrite (GPIO)")
	} else {
		t.Fail("digitalWrite (GPIO)")
	}

	if analogReadVoltage() {
		t.Pass("analogReadVoltage (ADC)")
	} else {
		t.Fail("analogReadVoltage (ADC)")
	}

	if analogReadGround() {
		t.Pass("analogReadGround (ADC)")
	} else {
		t.Fail("analogReadGround (ADC)")
	}

	if analogReadHalfVoltage() {
		t.Pass("analogReadHalfVoltage (ADC)")
	} else {
		t.Fail("analogReadHalfVoltage (ADC)")
	}

	if i2cConnection() {
		t.Pass("i2cConnection (MPU6050)")
	} else {
		t.Fail("i2cConnection (MPU6050)")
	}

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

func endTests() {}

// digital read of pin physically connected to V
func digitalReadVoltage() bool {
	readV.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	time.Sleep(200 * time.Millisecond)

	// should be on
	return readV.Get()
}

// digital read of pin physically connected to G
func digitalReadGround() bool {
	readG.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	time.Sleep(200 * time.Millisecond)

	// should be off
	return !readG.Get()
}

// digital write on/off of pin as input physically connected to different pin as output.
func digitalWrite() bool {
	readpin.Configure(machine.PinConfig{Mode: machine.PinInput})
	time.Sleep(200 * time.Millisecond)

	writepin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	time.Sleep(200 * time.Millisecond)

	writepin.Low()
	time.Sleep(200 * time.Millisecond)

	// should be off
	if readpin.Get() {
		return false
	}

	time.Sleep(200 * time.Millisecond)

	writepin.High()
	time.Sleep(200 * time.Millisecond)

	// should be on
	if !readpin.Get() {
		return false
	}

	return true
}

// analog read of pin connected to supply voltage.
func analogReadVoltage() bool {
	analogV.Configure(machine.ADCConfig{})
	time.Sleep(200 * time.Millisecond)

	// should be close to max
	var avg int
	for i := 0; i < 10; i++ {
		v := analogV.Get()
		avg += int(v)
		time.Sleep(10 * time.Millisecond)
	}
	avg /= 10
	val := uint16(avg)

	if val < maxanalog-allowedvariance {
		return false
	}

	return true
}

// analog read of pin connected to ground.
func analogReadGround() bool {
	analogG.Configure(machine.ADCConfig{})
	time.Sleep(500 * time.Millisecond)

	// should be close to zero
	var avg int
	for i := 0; i < 10; i++ {
		v := analogG.Get()
		avg += int(v)
		time.Sleep(10 * time.Millisecond)
	}
	avg /= 10
	val := uint16(avg)

	if val > allowedvariance {
		return false
	}

	return true
}

// analog read of pin connected to supply voltage that has been divided by 2
// using resistors.
func analogReadHalfVoltage() bool {
	analogHalf.Configure(machine.ADCConfig{})
	time.Sleep(200 * time.Millisecond)

	// should be around half the max
	var avg int
	for i := 0; i < 10; i++ {
		v := analogHalf.Get()
		avg += int(v)
		time.Sleep(10 * time.Millisecond)
	}
	avg /= 10
	val := uint16(avg)

	if val > maxanalog/2+allowedvariance || val < maxanalog/2-allowedvariance {
		return false
	}

	return true
}

// checks to see if an attached lis3dh accelerometer is connected.
func i2cConnection() bool {
	a := lis3dh.New(machine.I2C1)
	accel = &a
	accel.Address = lis3dh.Address1 // address on the Circuit Playground Express

	accel.Configure()
	time.Sleep(400 * time.Millisecond)

	if !accel.Connected() {
		return false
	}

	return true
}
