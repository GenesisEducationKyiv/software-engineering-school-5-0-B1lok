# ADR-004: Choosing a Message Broker

**Status:** Accepted  
**Date:** 2025-07-07  
**Author:** Bilokur Evgen

## Context

Our system architecture includes multiple background tasks such as weather data updates, scheduled email notifications, and subscription processing. These components need to communicate reliably in an asynchronous, decoupled manner. To facilitate this, we require a message broker that can:

- Support asynchronous communication across microservices.
- Ensure reliable message delivery (acknowledgment, retries).
- Scale with load, avoiding memory bottlenecks in producers and consumers.
- Integrate cleanly with Go applications.

The choice of a message broker must balance ease of integration for Go developers, operational complexity, reliability, and ecosystem maturity.

## Considered Options

### 1. RabbitMQ

**Pros:**
- Mature and stable with strong community and production usage.
- Reliable delivery guarantees (ack/nack, retries, dead-letter queues).
- Easy to set up locally using Docker.
- Supports multiple patterns (pub/sub, work queues, routing).
- Management UI for inspecting queues and messages.
- Persistent messages reduce memory pressure on consumers.

**Cons:**
- AMQP protocol is more complex than simpler brokers (e.g., NATS).
- Slightly more boilerplate required in Go compared to pub/sub-only brokers.

---

### 2. NATS

**Pros:**
- Extremely lightweight and fast.
- Minimal configuration — runs as a single binary.
- Simple and idiomatic Go client (`nats.go`).
- Great for fire-and-forget or transient messaging patterns.

**Cons:**
- Core NATS does not persist messages (requires JetStream for durability).
- Lacks built-in routing and advanced message guarantees without JetStream.
- Smaller ecosystem and fewer monitoring/inspection tools.
- Less suitable for critical delivery tasks like user-facing notifications.

---

### 3. Apache Kafka

**Pros:**
- Designed for high-throughput, large-scale event streaming.
- Strong durability and ordering guarantees.
- Ideal for analytics pipelines or event sourcing architectures.

**Cons:**
- Operationally heavy — requires Zookeeper or KRaft, hard to set up.
- Complex API and steeper learning curve.
- Go client integration is harder (`confluent-kafka-go` or `segmentio/kafka-go` require tuning).
- Overkill for simple job queues or task dispatching.

---

## Decision

We chose **RabbitMQ** as the message broker for our system.

RabbitMQ strikes the best balance between **reliability**, **feature set**, and **integration simplicity** for our needs. It enables reliable task dispatching and background job coordination, which are core to our subscription and notification services. Unlike NATS, it supports persistent queues and dead-lettering out of the box. Compared to Kafka, it is much easier to operate and use within a Go-based microservice environment.

## Consequences

### Positive

- We gain **reliable asynchronous communication** across services, ensuring delivery and durability.
- RabbitMQ’s **persistence** and **ack mechanisms** help prevent memory overload in high-throughput systems.
- Our **subscription microservice** can publish events without holding them in memory, and consumers can process them at their own pace.
- Enables future improvements such as **rate-limiting**, **retry policies**, and **dead-letter queues** with minimal effort.
- Easily deployable in local and cloud environments using Docker or Helm.
- Compatible with **streamed or cursor-based processing** to avoid loading entire datasets into RAM.

### Negative

- Slightly more verbose Go client code compared to simpler brokers like NATS.
- Requires maintaining and monitoring an external message broker service.
- Some operational complexity with queue configurations and delivery guarantees.

