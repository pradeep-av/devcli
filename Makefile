.PHONY: all build clean test docker-build

all: clean test build

build:
	@echo "Building devcli into ./build..."
	@mkdir -p build
	go build -o build/devcli main.go

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning build artifacts..."
	rm -rf build

docker-build:
	@echo "Building Docker image..."
	docker build -t devcli:latest .
