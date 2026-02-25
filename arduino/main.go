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
	"tinygo.org/x/tap"
)

var (
	// used by digital tests
	readV    = machine.D6
	readG    = machine.D7
	readpin  = machine.D10
	writepin = machine.D9

	// used by analog tests
	analogV    = machine.ADC{machine.ADC0}
	analogHalf = machine.ADC{machine.ADC1}
	analogG    = machine.ADC{machine.ADC2}

	// used by i2c tests
	accel *mpu6050.Device
)

const (
	maxanalog       = 65535
	allowedvariance = 3277
)

func main() {
	machine.Serial.Configure(machine.UARTConfig{BaudRate: 57600})
	machine.I2C0.Configure(machine.I2CConfig{})
	machine.InitADC()

	waitForStart()

	t := tap.New()
	t.Header(8)
	t.Ok(digitalReadVoltage(), "digitalReadVoltage (GPIO)")
	t.Ok(digitalReadGround(), "digitalReadGround (GPIO)")
	t.Ok(digitalWrite(), "digitalWrite (GPIO)")
	t.Ok(analogReadVoltage(), "analogReadVoltage (ADC)")
	t.Ok(analogReadGround(), "analogReadGround (ADC)")
	t.Ok(analogReadHalfVoltage(), "analogReadHalfVoltage (ADC)")
	t.Ok(i2cConnection(), "i2cConnection (MPU6050)")
	t.Ok(spiTxRx(), "spiTxRx (SPI)")

	endTests()
}

func endTests() {}

// wait for keypress on serial port to start test suite.
func waitForStart() {

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

// digital read of D11 pin physically connected to V
func digitalReadVoltage() bool {
	readV.Configure(machine.PinConfig{Mode: machine.PinInput})

	// should be on
	return readV.Get()
}

// digital read of D12 pin physically connected to G
func digitalReadGround() bool {
	readG.Configure(machine.PinConfig{Mode: machine.PinInput})

	// should be off
	return !readG.Get()
}

// digital write on/off of D9 pin as input physically connected to D10 pin as output.
func digitalWrite() bool {
	readpin.Configure(machine.PinConfig{Mode: machine.PinInput})
	writepin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	writepin.High()
	time.Sleep(100 * time.Millisecond)

	// should be on
	if !readpin.Get() {
		return false
	}

	time.Sleep(100 * time.Millisecond)

	writepin.Low()
	time.Sleep(100 * time.Millisecond)

	// should be off
	if readpin.Get() {
		return false
	}
	return true
}

// analog read of pin connected to supply voltage.
func analogReadVoltage() bool {
	analogV.Configure(machine.ADCConfig{})

	// should be close to max
	val := analogV.Get()
	if val >= maxanalog-allowedvariance {
		return true
	}
	return false
}

// analog read of pin connected to ground.
func analogReadGround() bool {
	analogG.Configure(machine.ADCConfig{})

	// should be close to zero
	val := analogG.Get()
	if val > allowedvariance {
		return false
	}

	return true
}

// analog read of pin connected to supply voltage that has been divided by 2
// using resistors.
func analogReadHalfVoltage() bool {
	analogHalf.Configure(machine.ADCConfig{})

	// should be around half the max
	val := analogHalf.Get()
	if val > maxanalog/2+allowedvariance || val < maxanalog/2-allowedvariance {
		return false
	}

	return true
}

// checks to see if an attached MPU-6050 accelerometer is connected.
func i2cConnection() bool {
	a := mpu6050.New(machine.I2C0)
	accel = &a

	accel.Configure()

	if !accel.Connected() {
		return false
	}

	return true
}

// checks if it is possible to send/receive by spi
func spiTxRx() bool {
	spi0 := machine.SPI0
	spi0.Configure(machine.SPIConfig{})

	from := make([]byte, 8)
	for i := range from {
		from[i] = byte(i)
	}
	to := make([]byte, len(from))

	err := spi0.Tx(from, to)
	if err != nil {
		return false
	}

	for i := range from {
		if from[i] != to[i] {
			return false
		}
	}
	return true
}
