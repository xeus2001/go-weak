GOBIN := $(shell pwd)/bin
SRC := $(shell find . -type f -name '*.go' -path "./pkg/weak/*")
LIBNAME := weak

.PHONY: setup-ubuntu
setup-ubuntu:
	@sudo apt-get install grc

.PHONY: test
test:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) grc --config=grc.conf go test -cover -coverprofile=coverage.out -v github.com/xeus2001/go-weak/pkg/weak

.PHONY: test-result
test-result:
	@go tool cover -html=coverage.out

.PHONY: clean
clean:
	@rm -f coverage.out

.PHONY: simplify
simplify:
	@gofmt -s -l -w $(SRC)
