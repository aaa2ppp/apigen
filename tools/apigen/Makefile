build: bin/apigen

bin/apigen: ./cmd/apigen/*.go ./internal/apigen/*.go
	go build -o bin/apigen ./cmd/apigen

codegen: build
	bin/apigen ./test && gofmt -w ./test/*_apigen.go

test: codegen
	go test -v ./test

.PHONY: build codegen test
