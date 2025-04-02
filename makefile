.PHONY: all test clean build

all: clean test build

build:
	go build -o ./bin/envx .

clean:
	rm ./bin/envx

test:
