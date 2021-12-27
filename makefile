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
