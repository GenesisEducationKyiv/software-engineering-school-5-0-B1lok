Testing Guide
=============

This project supports three types of tests:

*   **Unit Tests**

*   **Integration Tests**

*   **End-to-End (E2E) Tests**


Each test type uses Go build tags and can be run via Makefile commands.

Unit Tests
----------

Test isolated functions without external dependencies (e.g., database, network).
Requires Go installed.

**Run:**

```bash
make unit
```

**What it does:**

*   Runs go test with -tags=unit

*   Only runs tests with //go:build unit or +build unit


Integration Tests
-----------------

Test components working together (e.g., services + DB), without full app startup. Requires Docker.

**Run:**

```bash
make integration
```

**What it does:**

*   Runs go test with -tags=integration

*   Requires tests to be marked with //go:build integration


End-to-End (E2E) Tests
----------------------

Test full app behavior using real services (e.g., DB, API, mocks). Requires Docker.

**One-time setup:**

```bash
make install-playwright
```

**Run:**

```bash
make e2e
```

**What it does:**

*   Creates a Docker network (app-network) if missing

*   Starts mocks/services from scripts/docker-compose.yml

*   Starts test version of the app from docker-compose.base.yml + docker-compose.test.yml

*   Runs go test with -tags=e2e


Cleanup
-------

Stop and remove containers and volumes:

```bash
make stop 
```

Marking Tests
-------------

Use build tags at the top of test files:

```bash
//go:build unit
// +build unit 
```

```bash
//go:build integration
// +build integration 
```

```bash
//go:build e2e
// +build e2e 
```