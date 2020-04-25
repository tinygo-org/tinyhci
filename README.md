# TinyGo Integration Tests

Used to test actual hardware connections for microcontrollers.

Currently implemented integration tests for:

- Adafruit ItsyBitsy-M4

## How it works

The makefile compiles the test code, flashes it onto the connected microcontroller board, and then connects to the microcontroller serial port. The test program waits for a keypress to be detected on the serial port, and then runs thru the hardware integration tests, outputting the results back out to the serial port.

```
$ make test-itsybitsy-m4 
tinygo flash -size short -target=itsybitsy-m4 ./itsybitsy-m4/
   code    data     bss |   flash     ram
   9612      40    6360 |    9652    6400
Running tests...
digitalReadVoltage: pass
digitalReadGround: pass
digitalWriteOn: pass
digitalWriteOff: pass
i2cTests: pass
Tests complete.
```

### Digital I/O

The digital inputs and outputs are wired in a loopback form in order to test if the pins are outputting and reading the expected values. In addition, pins are wired to the voltage source, and ground, to ensure that all readings are returning known expected values.

### I2C

The I2C pins are wired to an MPU-6050 accelerometer to check if the device responds as expected when connected.

### ADC

Analog inputs are connected using a voltage divider made using resistors. The reference voltage, reference divided by 2, reference divided by 4, and ground levels can be read using the ADC.

### SPI

TODO

### PWM

TODO

### UART

TODO
