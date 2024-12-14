.PHONY: build

dev: build
	cd build && ./goantib

build: 
	mkdir -p ./build
	go build -o ./build/goantib ./cmd/goantib
