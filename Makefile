BINARY := smedje
MODULE := github.com/smedje/smedje
GOFLAGS := -trimpath
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)

.PHONY: build test lint bench install clean fmt vet web-dev web-build build-public landing-preview

build: web-build
	go build $(GOFLAGS) -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT)" -o $(BINARY) ./cmd/smedje

test:
	go test ./...

lint: vet fmt
	@echo "lint: ok"

vet:
	go vet ./...

fmt:
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)

bench:
	go test -bench=. -benchmem ./...

install: build
	go install $(GOFLAGS) -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT)" ./cmd/smedje

clean:
	rm -f $(BINARY)
	rm -rf frontend/dist internal/web/dist frontend/node_modules
	go clean -cache

web-dev:
	cd frontend && npm run dev

web-build:
	cd frontend && npm install --silent && npm run build

build-public: web-build
	go build $(GOFLAGS) -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.publicMode=true" -o $(BINARY) ./cmd/smedje

landing-preview:
	python3 -m http.server -d landing 8000
