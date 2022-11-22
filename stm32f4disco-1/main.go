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
// Analog read tests:
//	PC1 <--> 3.3V
//	PC4 <--> 3.3V/2 (use voltage divider)
//	PC5 <--> G
//
// I2C tests:
// 	STM32F407 SCL (PB6) <--> MPU-6050 SCL
// 	STM32F407 SDA (PB9) <--> MPU-6050 SDA
// 	STM32F407 G <--> MPU-6050 GND
// 	STM32F407 3V <--> MPU-6050 VCC
//
// SPI tests:
// 	STM32F407 SPI0 SDO (PA7) <--> STM32F407 SPI0 SDI (PA6)
//  *note that SPI0 and SPI1 are the same since there is no SPI0
//
import (
	"machine"

	"strconv"
	"time"

	"tinygo.org/x/drivers/mpu6050"
)

var (
	// used by digital tests
	readV    = machine.PD4
	readG    = machine.PD3
	readpin  = machine.PD2
	writepin = machine.PD1

	// used by analog tests
	analogV    = machine.ADC{machine.PC1}
	analogHalf = machine.ADC{machine.PC4}
	analogG    = machine.ADC{machine.PC5}

	// used by i2c tests
	accel *mpu6050.Device
)

const (
	maxanalog       = 65535
	allowedvariance = 4096
)

func main() {
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

// analog read of pin connected to supply voltage.
func analogReadVoltage() {
	analogV.Configure(machine.ADCConfig{})
	time.Sleep(100 * time.Millisecond)

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
	time.Sleep(100 * time.Millisecond)

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
		printfailexpected("'val <= 65535/2+" + strconv.Itoa(allowedvariance) + " && val >= 65535/2-" + strconv.Itoa(allowedvariance) + "'")
		printfailactual(val)
	}
}

// analog read of pin connected to supply voltage that has been divided by 2
// using resistors.
func analogReadHalfVoltage() {
	analogHalf.Configure(machine.ADCConfig{})
	time.Sleep(100 * time.Millisecond)

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
	printfailexpected("'val <= 65535/2+4096 && val >= 65535/2-4096'")
	printfailactual(val)
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

// checks if it is possible to send/receive by spi
func spiTxRx() {
	spi0 := machine.SPI0
	spi0.Configure(machine.SPIConfig{
		SCK:       machine.SPI0_SCK_PIN,
		SDO:       machine.SPI0_SDO_PIN,
		SDI:       machine.SPI0_SDI_PIN,
		Frequency: 4000000,
	})

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
