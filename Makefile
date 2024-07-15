# Install dependencies
deps:
	go mod download

# Check the code with a linter (golangci-lint)
lint:
	golangci-lint run

# Run tests
test:
	go test -tags=test -count=1 ./...

# Run tests with coverage computation
test_coverage:
	go test -tags=test -count=1 ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

# Generate swagger specification based on code
swag:
	swag init

# Build the project (production)
build:
	go build -tags=production -ldflags="-s -w" -o build/goserver.exe main.go

# Build the project (debug)
build_debug:
	go build -tags=debug -o build/goserver.exe main.go

# Run the project (production)
run:
	go run -tags=production main.go

# Run the project (debug)
run_debug:
	go run -tags=debug main.go

# Build a Docker image (production)
docker_build:
	docker build -f Dockerfile . --tag go-leaderboard

# Build a Docker image (debug)
docker_build_debug:
	docker build -f Dockerfile_debug . --tag go-leaderboard

# Run a Docker image (production)
docker_run: docker_build
	docker run -p 8415:8415 --rm --name go-leaderboard go-leaderboard

# Run a Docker image (debug)
docker_run_debug: docker_build_debug
	docker run -p 8415:8415 --rm --name go-leaderboard go-leaderboard
