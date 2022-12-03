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
	"strconv"

	"time"

	"tinygo.org/x/drivers/mpu6050"
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

	digitalReadVoltage()
	digitalReadGround()
	digitalWrite()
	analogReadVoltage()
	analogReadGround()
	analogReadHalfVoltage()
	i2cConnection()
	spiTxRx()

	endTests()
}

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

func endTests() {
	println("\n### Tests complete.")

	// tests done, now sleep waiting for baud reset to load new code
	for {
		time.Sleep(1 * time.Second)
	}
}

// digital read of D11 pin physically connected to V
func digitalReadVoltage() {
	printtest("digitalReadVoltage")

	readV.Configure(machine.PinConfig{Mode: machine.PinInput})

	// should be on
	if readV.Get() {
		printtestresult("pass")
		return
	}

	printtestresult("fail")
}

// digital read of D12 pin physically connected to G
func digitalReadGround() {
	printtest("digitalReadGround")

	readG.Configure(machine.PinConfig{Mode: machine.PinInput})

	// should be off
	if readG.Get() {
		printtestresult("fail")
		return
	}

	printtestresult("pass")
}

// digital write on/off of D9 pin as input physically connected to D10 pin as output.
func digitalWrite() {
	readpin.Configure(machine.PinConfig{Mode: machine.PinInput})
	writepin.Configure(machine.PinConfig{Mode: machine.PinOutput})

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

// analog read of pin connected to supply voltage.
func analogReadVoltage() {
	analogV.Configure(machine.ADCConfig{})

	printtest("analogReadVoltage")

	// should be close to max
	val := analogV.Get()
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

	printtest("analogReadGround")

	// should be close to zero
	val := analogG.Get()
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

	printtest("analogReadHalfVoltage")

	// should be around half the max
	val := analogHalf.Get()
	if val <= maxanalog/2+allowedvariance && val >= maxanalog/2-allowedvariance {
		printtestresult("pass")
		return
	}
	printtestresult("fail")

	printfailexpected("'val <= 65535/2+" + strconv.Itoa(allowedvariance) + " && val >= 65535/2-" + strconv.Itoa(allowedvariance) + "'")
	printfailactual(val)
}

// checks to see if an attached MPU-6050 accelerometer is connected.
func i2cConnection() {
	a := mpu6050.New(machine.I2C0)
	accel = &a

	printtest("i2cConnection")
	accel.Configure()

	if !accel.Connected() {
		printtestresult("fail")
		return
	}

	printtestresult("pass")
}

// checks if it is possible to send/receive by spi
func spiTxRx() {
	spi0 := machine.SPI0
	spi0.Configure(machine.SPIConfig{})

	from := make([]byte, 8)
	for i := range from {
		from[i] = byte(i)
	}
	to := make([]byte, len(from))

	printtest("spiTx")
	err := spi0.Tx(from, to)
	if err != nil {
		printtestresult("fail")
	} else {
		printtestresult("pass")
	}

	printtest("spiRx")
	for i := range from {
		if from[i] != to[i] {
			printtestresult("fail")
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

func printfailactual(val uint16) {
	println("        actual:", val)
}
