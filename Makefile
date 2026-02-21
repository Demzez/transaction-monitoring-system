run:
	CONFIG_PATH="./config/config.yaml" go run ./cmd/transaction-monitoring-system  
test:
	CONFIG_PATH="/home/demzez/dev/golang/transaction-monitoring-system/config/config.yaml" go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm coverage.out