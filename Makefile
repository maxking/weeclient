all: testclient build

build:
	go build -v

testclient:
	go build -v ./src/testclient

clean:
	rm weechat-go testclient

.PHONY: clean