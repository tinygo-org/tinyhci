TARGET_GOVERSION := go1.14.2
GOINSTALLED := $(shell command -v go 2> /dev/null)
CURRENT_GOVERSION  := $(shell go version | awk '{print $$3}')
TARGET_TINYGOVERSION := 0.13.1
TINYGOINSTALLED := $(shell command -v tinygo 2> /dev/null)
CURRENT_TINYGOVERSION  := $(shell tinygo version | awk '{print $$3}')
NOCOLOR := \033[0m
RED     := \033[0;31m
GREEN   := \033[0;32m

all: clean test-itsybitsy-m4 test-arduino-nano33 test-arduino-uno

test-itsybitsy-m4: build/testrunner
	tinygo flash -size short -target=itsybitsy-m4 -port=/dev/itsybitsy_m4 ./itsybitsy-m4/
	@sleep 2.0s
	@echo "Running tests..."
	./build/testrunner /dev/itsybitsy_m4 115200 5

test-arduino-nano33: build/testrunner
	tinygo flash -size short -target=arduino-nano33 -port=/dev/arduino_nano33 ./arduino-nano33/
	@sleep 2.0s
	@echo "Running tests..."
	./build/testrunner /dev/arduino_nano33 115200 5

test-arduino-uno: build/testrunner
	tinygo flash -size short -target=arduino -port=/dev/arduino_uno ./arduino/
	@echo "Running tests..."
	./build/testrunner /dev/arduino_uno 57600 5

update-go:
	@test "$(CURRENT_GOVERSION)" = "$(TARGET_GOVERSION)" && ( echo "$(RED)$(TARGET_GOVERSION) has already been installed$(NOCOLOR)\n" ; exit 1 )
	wget "https://dl.google.com/go/$(TARGET_GOVERSION).linux-amd64.tar.gz" -O /tmp/go.tar.gz
	sudo rm -rf /usr/local/go
	sudo tar -xzf /tmp/go.tar.gz -C /usr/local

install-go:
ifndef GOINSTALLED
	wget "https://dl.google.com/go/$(TARGET_GOVERSION).linux-amd64.tar.gz" -O /tmp/go.tar.gz
	sudo tar -xzf /tmp/go.tar.gz -C /usr/local
	@echo "# add Go to path" >> ~/.bashrc
	@echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
	source ~/.bashrc
endif
	echo "$(GREEN)Go is now installed:$(NOCOLOR)\n"
	go version

install-tinygo:
ifndef TINYGOINSTALLED
	wget "https://github.com/tinygo-org/tinygo/releases/download/v$(TARGET_TINYGOVERSION)/tinygo_$(TARGET_TINYGOVERSION)_amd64.deb"
	sudo dpkg -i "tinygo_$(TARGET_TINYGOVERSION)_amd64.deb"
	@echo "# add TinyGo to path" >> ~/.bashrc
	@echo 'export PATH=$PATH:/usr/local/tinygo/bin' >> ~/.bashrc
	source ~/.bashrc
endif
	echo "$(GREEN)TinyGo is now installed:$(NOCOLOR)\n"
	tinygo version

update-tinygo:
	@echo "Not yet"

install-bossa:
	sudo apt install libreadline-dev libwxgtk3.0-*
	mkdir -p build
	git clone https://github.com/shumatech/BOSSA.git build/BOSSA
	make -C build/BOSSA
	sudo cp build/BOSSA/bin/bossac /usr/local/bin

install-packages:
	go get -d -u tinygo.org/x/drivers

build/testrunner:
	mkdir -p build
	go build -o build/testrunner tools/testrunner/main.go

clean:
	rm -rf build
