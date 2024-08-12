build: genapi
	go build -o ./bin/server ./cmd/server

run: build
	bin/server

build_apigen:
	go build -o ./bin/apigen ./cmd/apigen

genapi: build_apigen
	./bin/apigen ./internal/api/api.go ./internal/api/api_handlers.go && gofmt -w ./internal/api/api_handlers.go
