package main

// Integration tests for SiFive HiFive1B
//
// Wire up the pins, and run it while connected to the USB port.
//
// Digital read/write tests:
//	D4 <--> G
//	D2 <--> 3V
//	D5 <--> D3
//
// I2C tests:
// 	HiFive1b SCL (D19) <--> MPU-6050 SCL
// 	HiFive1b SDA (D18) <--> MPU-6050 SDA
// 	HiFive1b G <--> MPU-6050 GND
// 	HiFive1b D9 <--> MPU-6050 VCC
//
// SPI (SPI1) tests:
// 	HiFive1b CDO - D11 <--> HiFive1b CDI - D12
//
// UART:
//  HiFive1b RX (D0) <--> FTDI TX
//  HiFive1b TX (D1) <--> FTDI RX
import (
	"machine"
	"strconv"

	"time"

	"tinygo.org/x/drivers/mpu6050"
)

var (
	// used by digital tests
	readV    = machine.D2
	readG    = machine.D4
	readpin  = machine.D3
	writepin = machine.D5

	// used by i2c tests
	accel    *mpu6050.Device
	powerpin = machine.D9
)

func main() {
	machine.Serial.Configure(machine.UARTConfig{})
	machine.I2C0.Configure(machine.I2CConfig{})
	machine.SPI1.Configure(machine.SPIConfig{
		SCK:       machine.SPI1_SCK_PIN,
		SDO:       machine.SPI1_SDO_PIN,
		SDI:       machine.SPI1_SDI_PIN,
		Frequency: 4000000,
	})

	waitForStart()

	digitalReadVoltage()
	digitalReadGround()
	digitalWrite()
	i2cConnection()
	spiTx()
	// TODO: fix SPI read
	// spiRx()

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
	readpin.Low()
	writepin.Low()
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
	powerpin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	time.Sleep(100 * time.Millisecond)

	a := mpu6050.New(machine.I2C0)
	accel = &a

	printtest("i2cConnection")

	// have to recycle power
	powerpin.Low()
	time.Sleep(500 * time.Millisecond)

	// turn on power and should be connected now
	powerpin.High()
	time.Sleep(500 * time.Millisecond)

	machine.I2C0.Configure(machine.I2CConfig{})

	err := accel.Configure()
	if err != nil {
	    printtestresult(err.Error())
	    return
	}
	time.Sleep(400 * time.Millisecond)

	if !accel.Connected() {
		printtestresult("fail")
		return
	}

	printtestresult("pass")
}

// checks if it is possible to send by spi
func spiTx() {
	spi1 := machine.SPI1

	from := make([]byte, 8)
	for i, _ := range from {
		from[i] = byte(i + 1)
	}
	to := make([]byte, len(from))

	printtest("spiTx")
	err := spi1.Tx(from, to)
	if err != nil {
		printtestresult("fail")
	} else {
		printtestresult("pass")
	}
}

// checks if it is possible to receive by spi
func spiRx() {
	spi1 := machine.SPI1

	from := make([]byte, 8)
	for i, _ := range from {
		from[i] = byte(i + 1)
	}
	to := make([]byte, len(from))

	printtest("spiRx")
	err := spi1.Tx(from, to)
	if err != nil {
		printtestresult("fail")
	}

	for i, _ := range from {
		if from[i] != to[i] {
			printtestresult("fail")
			printfailexpected("from[" + strconv.Itoa(i) + "] != to[" + strconv.Itoa(i) + "]: " + strconv.Itoa(int(from[i])))
			printfailactual(strconv.Itoa(int(to[i])))
			return
		}
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

func printfailactual(val string) {
	println("        actual:", val)
}
