# TinyGo Integration Tests

Used to test actual hardware connections for microcontrollers. It is intended to provide smoke test implementations that exercise the basic functionality for each kind of hardware interface for each supported microcontroller.

Currently implemented integration tests for:

- Adafruit ItsyBitsy-M4
- Arduino Nano33-IoT
- Arduino Uno

## Hardware Continuous Integration (HCI) System

The TinyGo HCI system is a Github application that uses the webhook interface.

It listens for pull requests to the target repository, and then will do the following:

- [ ] Pull down the latest binary build of TinyGo based on the pull request
- [ ] Create a check run for the PR (https://developer.github.com/v3/checks/)
- [x] Execute the hardware tests for each of the supported microcontroller boards
- [ ] Update the check run with the test results for each MCU to either "success" or "failed" based on the pass/fail for each MCU's as they are executed by the HCI system.

Thanks to @maruel for the work on GoHCI which has been an influence on this project.

## How it works

The makefile compiles the test code, flashes it onto the connected board, and then connects to the microcontroller serial port. The test program waits for a keypress to be detected on the serial port, and then runs thru the hardware integration tests, outputting the results back out to the serial port.

```
$ make test-itsybitsy-m4 
tinygo flash -size short -target=itsybitsy-m4 ./itsybitsy-m4/
   code    data     bss |   flash     ram
  11020      40    6360 |   11060    6400
Running tests...
digitalReadVoltage: pass
digitalReadGround: pass
digitalWriteOn: pass
digitalWriteOff: pass
analogReadVoltage: pass
analogReadGround: pass
analogReadHalfVoltage: pass
i2cConnectionNoPower: pass
i2cConnectionPower: pass
Tests complete.
```

### Digital I/O

The digital inputs and outputs are wired in a loopback form in order to test if the pins are outputting and reading the expected values. In addition, pins are wired to the voltage source, and ground, to ensure that all readings are returning known expected values.

### I2C

The I2C pins are wired to an MPU-6050 accelerometer to check if the device responds as expected when connected.

### ADC

Analog inputs are connected using a voltage divider made using two resistors. The reference voltage, reference divided by 2, and ground level voltage can then be read using the ADC.

### SPI

TODO

### PWM

TODO

### UART

TODO
