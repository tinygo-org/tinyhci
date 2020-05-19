# TinyGo Integration Tests

Used to test actual hardware connections for microcontrollers. It is intended to provide smoke test implementations that exercise the basic functionality for each kind of hardware interface for each supported microcontroller.

Currently implemented integration tests for:

- Adafruit ItsyBitsy-M4
- Arduino Nano33-IoT
- Arduino Uno

## Hardware Continuous Integration (HCI) System

The TinyGo HCI system is a Github application that uses the webhook interface.

It listens for pull requests to the target repository, and then will do the following:

- [x] Create a new docker image that downloads and installs the binary build of TinyGo based on the pull request SHA
- [x] Create a Github check suite for the PR (https://developer.github.com/v3/checks/)
- [x] Flash the hardware tests onto each of the supported microcontroller boards using the docker image
- [x] Execute the hardware tests for each of the supported microcontroller boards using the test runner
- [x] Create a Github check run in the check suite for this SHA with the test results for each MCU to either "success" or "failed" based on the pass/fail for each as they are executed by the HCI system.

## Test Runner

The process of running the hardware tests is:

- compile the test code for that MCU
- flash test code onto the MCU
- MCU test program waits for a keypress to be detected on the serial port
- HCI system connects to the MCU via serial port, and sends the key code to start the test run
- MCU runs thru the hardware integration tests, outputting the results back out to the serial port

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

## Docker containerized builds

We run each set of checks using a docker container with the associated `tinygo` binary for simplicity and greater security.

To build it:

```
docker build -t tinygohci -f tools/docker/Dockerfile --build-arg TINYGO_DOWNLOAD_URL=https://13064-136505169-gh.circle-artifacts.com/0/tmp/tinygo.linux-amd64.tar.gz .
```

Now we can use the `tinygohci:latest` image to build/flash our program.

```
docker run --device=/dev/ttyACM0 -v /media:/media:shared tinygohci:latest tinygo flash -target circuitplay-express examples/blinky1
```

## Thanks

Thanks to @maruel for the work on GoHCI which has been an influence on this project. 

Also thanks to Github for providing the free code hosting and CircleCI for providing the CI services that are the foundation for this project.

