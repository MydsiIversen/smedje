BINARY := smedje
MODULE := github.com/smedje/smedje
GOFLAGS := -trimpath

.PHONY: build test lint bench install clean fmt vet

build:
	go build $(GOFLAGS) -o $(BINARY) ./cmd/smedje

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

install:
	go install $(GOFLAGS) ./cmd/smedje

clean:
	rm -f $(BINARY)
	go clean -cache
