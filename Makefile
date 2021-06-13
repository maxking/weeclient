all: testclient build

build:
	go build -v

testclient: src/testclient/*.go
	go build -v ./src/testclient

clean:
	rm weeclient testclient

.PHONY: clean