BINARY := ./bin/usdt-service
MAIN := ./cmd/app/main.go
PKG := ./...

build:
	go build -o $(BINARY) $(MAIN)

run:
	go run $(MAIN)

test:
	go test -v -cover $(PKG)

docker-build:
	docker build -t grpc-usdt-rate .

docker-up: docker-build
	docker-compose up -d

docker-down:
	docker-compose down

lint:
	golangci-lint run
