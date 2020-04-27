
test-itsybitsy-m4:
	tinygo flash -size short -target=itsybitsy-m4 ./itsybitsy-m4/
	@sleep 4.0s
	@echo "Running tests..."
	@./runtest.sh /dev/ttyACM0 115200 5.0s 1.0s

test-arduino-nano33:
	tinygo flash -size short -target=arduino-nano33 ./arduino-nano33/
	@sleep 4.0s
	@echo "Running tests..."
	@./runtest.sh /dev/ttyACM0 115200 5.0s 1.0s

test-arduino-uno:
	tinygo flash -size short -target=arduino ./arduino/
	@echo "Running tests..."
	@./runtest.sh /dev/ttyACM0 57600 5.0s 3.0s
