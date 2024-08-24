# HeliProject Project

This project consists of three microservices: `gateway-service`, `account-service`, and `customer-service`. These microservices are built using Golang and communicate with each other via gRPC. The project also includes a PostgreSQL database and RabbitMQ for event messaging.

## Services

### Gateway Service

The `gateway-service` acts as a gateway for the other two microservices and provides RESTful APIs for user registration, login, logout, deposit, withdraw, balance inquiry, and transaction history.

### Account Service

The `account-service` is responsible for managing customer accounts, including deposit and withdraw operations. It also handles balance inquiries and transaction history. The service uses PostgreSQL for data storage and RabbitMQ for event messaging.

### Customer Service

The `customer-service` manages customer records, including user registration, login, and logout. It uses PostgreSQL for data storage and provides gRPC endpoints for user management.

## Project Dependencies

- Golang
- Docker
- Docker Compose
- PostgreSQL
- RabbitMQ

## Running the Project

### Prerequisites

Ensure you have Docker and Docker Compose installed on your machine.

### Build and Run the Services

To build and run the services, use the provided Makefile. Navigate to the directory containing the `Makefile` and run the following commands:

Be aware that these commands should be executed inside "git bash for windows" or linux terminals.

1. **Build all services**:

    ```sh
    make build
    ```

2. **Run all services**:

    ```sh
    make run
    ```

### Testing the Services

To run the tests for all the services, use the following command:

```sh
make test

Clean Up
To stop and remove all the services, use the following command:

make clean

API Documentation
    The gateway-service includes Swagger documentation for the APIs. Once the services are running, you can access the Swagger UI at:

http://localhost:8080/swagger/index.html

Environment Variables
    The following environment variables are used in the services:

    POSTGRES_HOST: Hostname for the PostgreSQL database.
    POSTGRES_USER: Username for the PostgreSQL database.
    POSTGRES_PASSWORD: Password for the PostgreSQL database.
    POSTGRES_DB: Database name for the PostgreSQL database.
    RABBITMQ_HOST: Hostname for RabbitMQ.
    These variables are defined in the docker-compose.yml file.

Directory Structure
.
├── gateway-service
│   ├── Dockerfile
│   ├── main.go
│   ├── account_service.go
│   ├── account_service_test.go
│   └── ...
├── account-service
│   ├── Dockerfile
│   ├── main.go
│   ├── account_service.go
│   ├── account_service_test.go
│   └── ...
├── customer-service
│   ├── Dockerfile
│   ├── main.go
│   └── ...
├── docker-compose.yml
└── Makefile

I've created an insomnia json file, named "test_api_insomnia.json" that could be imported into Insomnia software for testing the APIs.
