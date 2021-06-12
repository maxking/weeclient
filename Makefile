all: testclient build

build:
	go build -v

testclient:
	go build -v ./src/testclient

clean:
	rm weeclient testclient

.PHONY: clean