package main

// Integration tests for RP2040 Pico
//
// Wire up the pins, and run it while connected to the USB port.
//
// Digital read/write tests:
//	GP6 <--> G
//	GP7 <--> 3V
//	GP8 <--> GP9
//
// Analog read tests:
//	ADC0 (GP26) <--> 3.3V
//	ADC1 (GP27) <--> 3.3V/2 (use voltage divider)
//	ADC2 (GP28) <--> G
//
// I2C tests:
// 	Pico GP5 SCL <--> MPU-6050 SCL
// 	Pico GP4 SDA <--> MPU-6050 SDA
// 	Pico G <--> MPU-6050 GND
// 	Pico GP3 <--> MPU-6050 VCC
//
// SPI tests:
// 	Pico GP16 SDO <--> GP19 Pico SDI
import (
	"device/rp"
	"machine"
	"runtime"

	"time"
	"unsafe"

	"tinygo.org/x/drivers/mpu6050"
	"tinygo.org/x/tap"
)

var (
	// used by digital tests
	readV    = machine.GP7
	readG    = machine.GP6
	readpin  = machine.GP8
	writepin = machine.GP9

	// used by analog tests
	analogV    = machine.ADC{machine.ADC0}
	analogHalf = machine.ADC{machine.ADC1}
	analogG    = machine.ADC{machine.ADC2}

	// used by i2c tests
	powerpin = machine.GP22
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

	t := tap.New()
	t.Header(10)

	t.Ok(digitalReadVoltage(), "digitalReadVoltage (GPIO)")
	t.Ok(digitalReadGround(), "digitalReadGround (GPIO)")
	t.Ok(digitalWrite(), "digitalWrite (GPIO)")
	t.Ok(analogReadVoltage(), "analogReadVoltage (ADC)")
	t.Ok(analogReadGround(), "analogReadGround (ADC)")
	t.Ok(analogReadHalfVoltage(), "analogReadHalfVoltage (ADC)")
	t.Ok(i2cConnection(), "i2cConnection (I2C)")
	t.Ok(spiDMARecovery(), "spiDMARecovery (SPI)")
	t.Ok(spiDMAChaining(), "spiDMAChaining (SPI)")
	t.Ok(spiTxRx(), "spiTxRx (SPI)")

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

func endTests() {}

// digital read of D11 pin physically connected to V
func digitalReadVoltage() bool {
	readV.Configure(machine.PinConfig{Mode: machine.PinInput})
	time.Sleep(100 * time.Millisecond)

	// should be on
	return readV.Get()
}

// digital read of D12 pin physically connected to G
func digitalReadGround() bool {
	readG.Configure(machine.PinConfig{Mode: machine.PinInput})
	time.Sleep(100 * time.Millisecond)

	// should be off
	return !readG.Get()
}

// digital write on/off of D9 pin as input physically connected to D10 pin as output.
func digitalWrite() bool {
	readpin.Configure(machine.PinConfig{Mode: machine.PinInput})
	writepin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	time.Sleep(100 * time.Millisecond)

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
	time.Sleep(100 * time.Millisecond)

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
	time.Sleep(100 * time.Millisecond)

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
	time.Sleep(100 * time.Millisecond)

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

// checks to see if an attached MPU-6050 accelerometer is connected.
func i2cConnection() bool {
	powerpin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	time.Sleep(100 * time.Millisecond)

	accel := mpu6050.New(machine.I2C0)

	// should not be connected when not powered
	powerpin.Low()
	time.Sleep(100 * time.Millisecond)
	if accel.Connected() {
		return false
	}

	// turn on power and should be connected now
	powerpin.High()
	time.Sleep(100 * time.Millisecond)

	machine.I2C0.Configure(machine.I2CConfig{})

	err := accel.Configure()
	if err != nil {
		return false
	}
	time.Sleep(100 * time.Millisecond)

	if !accel.Connected() {
		return false
	}

	return true
}

func spiDMARecovery() bool {
	spi0 := machine.SPI0
	err := spi0.Configure(machine.SPIConfig{
		SCK:       machine.SPI0_SCK_PIN,
		SDO:       machine.SPI0_SDO_PIN,
		SDI:       machine.SPI0_SDI_PIN,
		Frequency: 4000000,
	})
	if err != nil {
		return false
	}

	// DMA cannot read SIO registers on RP2040 or RP2350.
	bad := unsafe.Slice((*byte)(unsafe.Pointer(uintptr(0xd0000000))), 64)
	if err := spi0.Tx(bad, nil); err != machine.ErrSPIDMA {
		return false
	}

	good := make([]byte, 64)
	for i := range good {
		good[i] = byte(i)
	}
	return spi0.Tx(good, nil) == nil
}

func spiDMAChaining() bool {
	err := machine.SPI1.Configure(machine.SPIConfig{
		SCK:       machine.GP10,
		SDO:       machine.GP11,
		SDI:       machine.NoPin,
		Frequency: 4000000,
	})
	if err != nil {
		return false
	}

	// Keep the next 64 bytes valid if SPI1 restarts SPI0's DMA channel.
	from := make([]byte, 128)
	for i := range from {
		from[i] = byte(i)
	}
	if err := machine.SPI0.Tx(from[:64], nil); err != nil {
		return false
	}

	err = machine.SPI1.Tx([]byte{0x55, 0xaa}, nil)
	// Allow an unintended 64-byte SPI0 transfer to finish at 4 MHz.
	time.Sleep(time.Millisecond)
	for machine.SPI0.Bus.SSPSR.HasBits(rp.SPI0_SSPSR_BSY) {
	}
	unexpected := machine.SPI0.Bus.SSPSR.HasBits(rp.SPI0_SSPSR_RNE)
	for machine.SPI0.Bus.SSPSR.HasBits(rp.SPI0_SSPSR_RNE) {
		machine.SPI0.Bus.SSPDR.Get()
	}
	machine.SPI0.Bus.SSPICR.Set(rp.SPI0_SSPICR_RORIC)
	runtime.KeepAlive(from)

	return err == nil && !unexpected
}

// checks if it is possible to send/receive by spi
func spiTxRx() bool {
	from := make([]byte, 8)
	for i := range from {
		from[i] = byte(i)
	}
	to := make([]byte, len(from))

	err := machine.SPI0.Tx(from, to)
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
