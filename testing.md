### Running Tests with Testcontainers locally

To run integration tests that use **Testcontainers**, make sure the following prerequisites are met:

#### Prerequisites

- Docker is installed and running on your machine.
- Go is installed (version 1.20+ recommended).

#### Run the Tests

Use the following command to run all tests, including those using Testcontainers:

```bash
go test ./... -v
```