.PHONY: all build clean install test

all: clean test build

build:
	go build -o ./bin/envx .

clean:
	rm ./bin/envx

install:
	go install github.com/almahoozi/envx@latest

test:
