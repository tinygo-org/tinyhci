// Integration tests for Xiao ESP32-C3
//
// Wire up the pins, and run it while connected to the USB port.
//
// Digital read/write tests (GPIO):
//
//	D0  <--> 3V3
//	D2  <--> D3
//
// I2C tests:
//
//	Xiao ESP32-C3 SCL (D5) <--> MPU6050 SCL
//	Xiao ESP32-C3 SDA (D4) <--> MPU6050 SDA
//	Xiao ESP32-C3 G        <--> MPU6050 GND
//	Xiao ESP32-C3 3V3      <--> MPU6050 VCC
//
// SPI tests:
//
//	Xiao ESP32-C3 CDO - D10 <--> Xiao ESP32-C3 CDI - D9
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/mpu6050"
	"tinygo.org/x/tap"
)

func main() {
	waitForStart()

	t := tap.New()
	t.Header(4)

	t.Ok(digitalReadVoltageGPIO(), "digitalReadVoltage (GPIO)")
	t.Ok(digitalWriteGPIO(), "digitalWrite (GPIO)")
	t.Ok(i2cConnection(), "i2cConnection (MPU6050)")
	t.Ok(spiTxRx(), "spiTxRx (SPI)")

	endTests()
}

// Wait for a signal to start tests (e.g., from serial)
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

// Signal end of tests (optional)
func endTests() {
	// Implement as needed for your environment.
}

// Example test functions
func digitalReadVoltageGPIO() bool {
	readV := machine.D0
	readV.Configure(machine.PinConfig{Mode: machine.PinInput})
	return readV.Get()
}

func digitalWriteGPIO() bool {
	writepin := machine.D3
	readpin := machine.D2
	writepin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	readpin.Configure(machine.PinConfig{Mode: machine.PinInput})

	writepin.High()
	time.Sleep(100 * time.Millisecond)
	if !readpin.Get() {
		return false
	}
	writepin.Low()
	time.Sleep(100 * time.Millisecond)
	return !readpin.Get()
}

func i2cConnection() bool {
	i2c := machine.I2C0
	i2c.Configure(machine.I2CConfig{})
	a := mpu6050.New(machine.I2C0)
	accel := &a

	err := accel.Configure()
	if err != nil {
		return false
	}
	time.Sleep(500 * time.Millisecond)

	if !accel.Connected() {
		return false
	}

	return true
}

func spiTxRx() bool {
	spi0 := machine.SPI2
	spi0.Configure(machine.SPIConfig{
		SCK:       machine.SPI_SCK_PIN,
		SDO:       machine.SPI_SDO_PIN,
		SDI:       machine.SPI_SDI_PIN,
		Frequency: 4000000,
	})

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
