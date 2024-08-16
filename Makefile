.PHONY: all build run test clean

# Define the services
SERVICES = gateway-service account-service customer-service

# Default target
all: build

# Build all services
build:
	docker-compose up --build

# Run all services
run:
	@echo "Starting all services..."
	docker-compose up

# Test all services
test:
	for service in $(SERVICES); do \
		echo "Testing $$service..."; \
        docker-compose run --rm $$service go test -v ./...; \
	done


# Clean up
clean:
	@echo "Stopping and removing all services..."
	docker-compose down
	@for service in $(SERVICES); do \
		echo "Cleaning $$service..."; \
		docker-compose run --rm $$service go clean -testcache; \
	done
