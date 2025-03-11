.PHONY: build
build:
	@echo "build go code.."
	go build

.PHONY: run
run:
	@echo "starting go code.."
	go run main.go

.PHONY: generateSQLC
generateSQLC: 
	@echo "regenerate sqlc code.."
	sqlc generate

.PHONY: lint
lint:
	@echo "checking for lint errors..."
	go vet ./...
