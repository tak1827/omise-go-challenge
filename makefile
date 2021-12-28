VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=mchain \
	-X github.com/cosmos/cosmos-sdk/version.ServerName=cosmmindend \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT)

.PHONY: test
test:
	go test -race ./...

bench:
	go test ./... -bench=. -benchtime=10s

fmt:
	go fmt ./...

lint:
	go vet ./...

build:
	go build -o challenge -gcflags '-m'

time:
	gtime -f '%P %Uu %Ss %er %MkB %C' "$@" ./challenge

srv:
	go run test/srv.go

run:
	go run ./ ./data/fng.1000.csv.rot128

install: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/cosmmindend