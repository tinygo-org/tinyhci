
test-itsybitsy-m4:
	tinygo flash -size short -target=itsybitsy-m4 ./itsybitsy-m4/
	@sleep 4.0s
	@echo "Running tests..."
	@./runtest.sh
