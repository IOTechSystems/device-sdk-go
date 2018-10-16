.PHONY: build test clean prepare update

GO=CGO_ENABLED=0 go
GOFLAGS=-ldflags

MICROSERVICES=example/cmd/simple-device/simple-device
.PHONY: $(MICROSERVICES)

build: $(MICROSERVICES)
	go build ./...

cmd/simple-device:
	$(GO) build -o $@ ./example/cmd/simple-device

test:
	go test ./... -cover

clean:
	rm -f $(MICROSERVICES)

prepare:
	glide install

update:
	glide update
