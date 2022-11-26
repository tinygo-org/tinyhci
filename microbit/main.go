package main

// Integration tests for BBC:microbit
//
// Wire up the pins, and run it while connected to the USB port.
//
// Digital read/write tests:
//	P16 <--> G
//	P3 <--> 3V
//	P4 <--> P5
//
// Analog read tests:
//	P0 <--> 3.3V
//	P1 <--> 3.3V/2 (use voltage divider)
//	P2 <--> G
//
// I2C tests:
// 	Uses built-in MAG3110 I2C device
//
// SPI tests:
// 	CDI (P14) <--> CDO (P15)
//
import (
	"machine"

	"strconv"
	"time"

	"tinygo.org/x/drivers/mag3110"
)

var (
	// used by digital tests
	readV    = machine.P3
	readG    = machine.P16
	readpin  = machine.P5
	writepin = machine.P4

	// used by analog tests
	analogV    = machine.ADC{machine.P0}
	analogHalf = machine.ADC{machine.P1}
	analogG    = machine.ADC{machine.P2}

	// used by i2c tests
	mag *mag3110.Device
)

const (
	maxanalog       = 65535
	allowedvariance = 4096
)

func main() {
	machine.Serial.Configure(machine.UARTConfig{})
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

// checks to see if the onboard MAG3110 is connected.
func i2cConnection() {
	printtest("i2cConnection")

	m := mag3110.New(machine.I2C0)
	m.Configure()
	mag = &m

	if !mag.Connected() {
		printtestresult("fail")
		return
	}

	printtestresult("pass")
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
