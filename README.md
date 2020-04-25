# TinyGo Integration Tests

Used to test actual hardware connections for microcontrollers.

```
$ make test-itsybitsy-m4 
tinygo flash -size short -target=itsybitsy-m4 ./itsybitsy-m4/
   code    data     bss |   flash     ram
   8512       8    6360 |    8520    6368
Running tests...
digitalReadVoltage: pass
digitalReadGround: pass
digitalWriteOn: pass
digitalWriteOff: pass
Tests complete.
```
