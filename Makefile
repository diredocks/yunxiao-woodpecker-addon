tidy:
	@go mod tidy

audit:
	@go mod verify
	@go vet ./...

lint:
	@golangci-lint run

test:
	@go test -race ./...

test/cover:
	@go test -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

build:
	@CGO_ENABLED=0 go build \
		-ldflags "-s -w -X 'yunxiao-woodpecker-addon/pkg/version.Version=dev' -X 'yunxiao-woodpecker-addon/pkg/version.BuildTime=$(shell date)'" \
		-o ./yunxiao-woodpecker-addon
