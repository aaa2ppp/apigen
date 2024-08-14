build: genapi
	go build -o ./bin/server ./cmd/server

run: build
	bin/server

genapi: apigen_tool
	tools/apigen/bin/apigen ./internal/service && gofmt -w ./internal/service/*_apigen.go

apigen_tool:
	cd tools/apigen && make build

.PHONY: build run genapi apigen_tool
