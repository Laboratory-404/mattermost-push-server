.PHONY: all build build-linux run test install clean docker-build

all: build

build:
	@echo "Building binaries for host OS..."
	go build -ldflags="-s -w" -o build/mattermost-push-proxy ./cmd/server
	go build -ldflags="-s -w" -o build/test-push ./cmd/test-push

build-linux:
	@echo "Cross-compiling Linux binaries..."
	bash scripts/build.sh

run: build
	./build/mattermost-push-proxy -config config/config.example.yaml

test:
	go test -v ./...

tidy:
	go mod tidy

clean:
	rm -rf build/

docker-build:
	docker build -t mattermost-push-proxy:latest -f docker/Dockerfile .

install:
	@echo "Installing on Ubuntu..."
	sudo bash scripts/install.sh
