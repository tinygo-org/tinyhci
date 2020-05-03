TARGET_VERSION := go1.14.2
CURRENT_VERSION  := $(shell go version | awk '{print $$3}')
NOCOLOR := \033[0m
RED     := \033[0;31m

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

update-go:
	@test "$(CURRENT_VERSION)" = "$(TARGET_VERSION)" && ( echo "$(RED)$(TARGET_VERSION) has already been installed$(NOCOLOR)\n" ; exit 1 )
	wget "https://dl.google.com/go/$(GO_VERSION).linux-amd64.tar.gz" -O /tmp/go.tar.gz
	sudo rm -rf /usr/local/go
	sudo tar -xzf /tmp/go.tar.gz -C /usr/local
