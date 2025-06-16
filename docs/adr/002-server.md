# ADR-002: Choosing an HTTP server framework

**Status:** Accepted  
**Date:** 2025-06-09  
**Author:** Bilokur Evgen

## Context

Our weather notification system needs to expose a set of RESTful endpoints for users to subscribe, confirm, unsubscribe, and fetch weather data. The HTTP server must be lightweight, performant, easy to use, and integrate well with Go's standard library and middleware patterns.

The application is written in Go and follows a Domain-Driven Design (DDD) architecture. We need an HTTP framework that allows clean separation of concerns, clear routing, and straightforward integration with request validation, logging, and error handling.

## Considered Options

### 1. Go `net/http` (standard library)

**Pros:**
- Comes built-in with Go, no dependencies.
- Very performant and low-level.
- Fine-grained control over the request lifecycle.

**Cons:**
- Verbose to write even simple routing logic.
- No built-in support for middleware chaining or parameter binding.
- Requires a lot of boilerplate for typical REST patterns.

---

### 2. Gin

**Pros:**
- Very fast and lightweight framework with good performance.
- Clear and expressive routing syntax.
- Built-in middleware support (logging, recovery, CORS, etc.).
- Easy binding of JSON payloads and URL parameters to structs.
- Large community and good documentation.

**Cons:**
- Adds a dependency, slightly more abstract than `net/http`.
- Some learning curve if coming from raw `net/http`.

---

### 3. Echo

**Pros:**
- Similar to Gin, also performant and lightweight.
- Good support for request validation and file uploads.
- Slightly more modular than Gin in some areas.

**Cons:**
- Smaller community and ecosystem than Gin.
- Documentation and examples are less mature.

---

### 4. Fiber

**Pros:**
- Inspired by Express.js (Node.js), familiar structure for web developers.
- Extremely fast due to usage of `fasthttp`.
- Great for high-throughput APIs.

**Cons:**
- Not compatible with Go’s `net/http` (which limits some libraries).
- Less idiomatic and more opinionated.

## Decision

We chose **Gin** as the HTTP server framework. It offers the best balance between performance, ease of use, and developer experience. Its routing system is expressive and clear, and it integrates seamlessly with middleware and JSON binding — which are essential for a clean REST API in our system.

## Consequences

### Positive

- Fast development with minimal boilerplate for REST endpoints.
- Built-in support for routing, middleware, request binding, and error handling.
- Enables clean and maintainable controller structure in the interface layer.
- Well-documented and widely used in the Go ecosystem.

### Negative

- Adds a third-party dependency.
- Less control compared to raw `net/http`, though still flexible enough for our use case.
