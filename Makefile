
test-itsybitsy-m4:
	tinygo flash -size short -target=itsybitsy-m4 ./itsybitsy-m4/
	@sleep 4.0s
	@echo "Running tests..."
	@./runtest.sh

test-arduino-nano33:
	tinygo flash -size short -target=arduino-nano33 ./arduino-nano33/
	@sleep 4.0s
	@echo "Running tests..."
	@./runtest.sh
