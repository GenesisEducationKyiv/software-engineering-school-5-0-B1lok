# Weather Forecast Subscription Service

This repository contains a fully implemented **Weather Forecast API service**, which allows users to **subscribe to regular weather updates** for a selected city.

The service follows a predefined API contract (Swagger 2.0) and includes full backend functionality, a subscription HTML page, unit and integration tests, Docker-based deployments, and a complete CI/CD pipeline to Google Cloud Platform (GCP).

---

## Features

- **REST API** that fully complies with the provided Swagger specification.
- **Subscription system** with email confirmation and unsubscribe endpoints.
- **Support for hourly and daily weather updates**.
- **Persisted data in a relational database**.
- **Database migrations** are automatically applied on service startup.
- **Unit tests** with mocks for core services and repositories.
- **Integration tests** using [Testcontainers](https://www.testcontainers.org/).
- **CI/CD Pipeline**:
  - Runs all tests (unit + integration).
  - Builds and pushes a Docker image to **Docker Hub**.
  - Deploys the app on a **GCP Virtual Machine**.
- **Production configuration**:
  - Uses a **secure GCP-hosted database**.
  - Configured **NGINX reverse proxy**.
  - HTTPS enabled via SSL key.
  - **Custom domain name** configured using https://www.namecheap.com/
- **HTML subscription page** for easier end-user access.
  
---

## API Specification

The API follows this contract:

- `GET /weather?city=CityName` - Get current weather for the given city.
- `POST /subscribe` - Subscribe with email, city, and update frequency (hourly or daily).
- `GET /confirm/{token}` - Confirm a new subscription via email token.
- `GET /unsubscribe/{token}` - Unsubscribe via email token.

---

## Testing

- **Unit Tests**: Implemented for core services and repository layers using mocks.
- **Integration Tests**: Use **Testcontainers** to spin up real PostgreSQL and simulate environment.
- **Test Coverage**: Includes subscription flows, weather retrieval, and token validation logic.

---

## CI/CD Pipeline

CI/CD is configured to:

1. Build and test the application using GitHub Actions.
2. Run integration tests with Docker + Testcontainers.
3. Build and push the application image to **Docker Hub**.
4. SSH into a **GCP VM**, pull the new image, and restart the application.

Secrets (API keys, SSH keys, etc.) are stored securely in GitHub Secrets.

---

## Deployment Architecture

| Component       | Technology Used        |
|----------------|------------------------|
| API Backend     | Go                     |
| DB (Prod)       | GCP Cloud SQL (PostgreSQL) |
| HTML Frontend   | Static HTML + CSS      |
| Reverse Proxy   | NGINX                  |
| SSL             | Namecheap Ssl key |
| Domain          | Custom Domain + DNS    |
| Containerization| Docker + Docker Compose |
| CI/CD           | GitHub Actions         |
| Hosting         | Google Cloud Platform  |

---

##  Production Environment Setup

- Custom domain: https://b1lokcode.me/  - hosted application available here
- SSL certificate installed.
- NGINX configured as a reverse proxy.
- Environment variables stored securely on server.
- Database connection is encrypted and secured by GCP firewall rules.

---

## Local Development

To run the Weather Subscription API locally, follow the steps below.

---

### Prerequisites

- [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/)
- [Git](https://git-scm.com/)
- Create your api key at https://www.weatherapi.com/
- Create your email key from your google account
---

### Step 1: Clone the Repository

```bash
git clone https://github.com/B1lok/weather-api.git
cd weather-api
```

### Step 2: Create `.env` File

Create a `.env` file in the project root and populate it with the following environment variables:

```env
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=weather
SERVER_HOST=http://localhost:8080/
SERVER_PORT=8080
WEATHER_API_URL=your-api-url
WEATHER_API_KEY=your-api-key
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USERNAME=your-email-adress-for-authentication
EMAIL_PASSWORD=your-api-key
EMAIL_FROM=your-email-adress
```
### Step 3: Start with Docker Compose

Build and run the application using Docker Compose:

```bash
docker compose -f docker-compose.base.yml -f docker-compose.dev.yml up --build
```
Application now available at http://localhost:8080/ with a custom subscription page

Observability
================================================================

Infrastructure Alerts
---------------------

### CPU Usage Alert

*   **Threshold**: CPU usage > 80% for 5 consecutive minutes

*   **Rationale**: High CPU usage can indicate resource exhaustion, inefficient code execution, or unexpected load spikes. Early detection prevents service degradation and allows for proactive scaling or optimization.


### Memory Usage Alert

*   **Threshold**: RAM usage > 85% for 3 consecutive minutes

*   **Rationale**: Memory leaks or high memory consumption can lead to application crashes or slow performance. Monitoring RAM usage ensures stable operation and helps identify memory-related issues before they impact users.


Application Performance Alerts
------------------------------

### Response Time Alert (95th Percentile)

*   **Threshold**: 95th percentile response time > 500ms for any endpoint

*   **Rationale**: The 95th percentile provides insight into the user experience for the majority of requests while accounting for occasional slower responses. Weather data requests should be fast to maintain good user experience.


### Server Error Rate Alert

*   **Threshold**: More than 10 HTTP 5xx errors within a 5-minute window

*   **Rationale**: Server errors (500-599) indicate critical issues such as database connectivity problems, third-party API failures, or application bugs. Quick detection allows for immediate investigation and prevents cascading failures that could affect weather data delivery.


### Database Connection Alert

*   **Threshold**: Database connection pool utilization > 90%

*   **Rationale**: Weather subscription services rely heavily on database operations for subscription tracking. Connection pool exhaustion can cause service unavailability.

---------------------------------

*   **Threshold**: Average database response time > 100ms over a 5-minute window

*   **Rationale**: Consistently high database latency can indicate performance bottlenecks or infrastructure issues. This can degrade user experience and delay critical operations like subscription processing and message delivery.

Log Retention Policy by Log Level
---------------------------------

### TRACE Level Logs

*   **Retention Period**: 3 days in hot storage only

*   **Cleanup Schedule**: Daily automated cleanup at 2 AM UTC

*   **Archive Strategy**: No archiving (immediate deletion after 3 days)

*   **Rationale**: TRACE logs contain the most granular debugging information and generate extremely high volume. They are primarily used for immediate debugging sessions and active development. Extended retention would create massive storage costs with minimal operational value. The 3-day window allows for recent issue investigation while preventing storage bloat.


### DEBUG Level Logs

*   **Retention Period**: 7 days in hot storage, no archiving

*   **Cleanup Schedule**: Daily automated cleanup at 2:15 AM UTC

*   **Archive Strategy**: No archiving (immediate deletion after 7 days)

*   **Rationale**: DEBUG logs provide detailed application flow information useful for troubleshooting complex issues. One week retention covers most debugging scenarios for weather service logic, subscription processing, and API integrations. DEBUG logs are too verbose for long-term storage but more valuable than TRACE for operational troubleshooting.


### INFO Level Logs

*   **Retention Period**: 30 days in hot storage, 6 months in cold storage

*   **Cleanup Schedule**: Daily automated cleanup at 2:30 AM UTC

*   **Archive Strategy**: Compress and move to cold storage after 30 days

*   **Rationale**: INFO logs capture normal application operations, user activities, and business events. This data is valuable for understanding usage patterns, subscription trends, and service performance over time. 30-day immediate access supports operational monitoring, while 6-month archive enables seasonal analysis for weather subscription patterns.


### WARN Level Logs

*   **Retention Period**: 60 days in hot storage, 12 months in cold storage

*   **Cleanup Schedule**: Daily automated cleanup at 2:45 AM UTC

*   **Archive Strategy**: Compress and categorize warnings by type before archiving

*   **Rationale**: WARN logs indicate potential issues that don't immediately break functionality but may lead to problems. Weather services often have intermittent external API issues, temporary network problems, or degraded performance that manifest as warnings. Extended retention helps identify recurring warning patterns and prevents them from escalating to errors.


### ERROR Level Logs

*   **Retention Period**: 90 days in hot storage, 24 months in cold storage

*   **Cleanup Schedule**: Weekly automated cleanup with error categorization

*   **Archive Strategy**: Categorize by error type, maintain full context and stack traces

*   **Rationale**: ERROR logs represent actual failures that impact user experience or system functionality. Long retention is critical for root cause analysis, identifying recurring issues, and understanding error patterns across different weather conditions and seasonal usage variations. Two-year archive supports compliance requirements and long-term system reliability analysis.


Storage Strategy
----------------

### Hot Storage (Immediate Access)

*   **Technology**: SSD-based storage with full-text search indexing

*   **Access Time**: < 1 second query response

*   **Cost**: Higher storage cost, optimized for frequent access


### Cold Storage (Archive)

*   **Technology**: Object storage (S3, Azure Blob) with compression

*   **Access Time**: 1-5 minutes retrieval time

*   **Cost**: Lower storage cost, optimized for long-term retention

Log Access by Role
------------------

### TRACE Logs
*   **Who can access**: Developers, DevOps engineers
*   **Restrictions**: Development environment only
*   **Audit**: All access attempts logged

### DEBUG Logs
*   **Who can access**: Developers, QA, L2+ support
*   **Restrictions**: Business hours only (except incidents)
*   **Audit**: Access time and search queries logged

### INFO Logs
*   **Who can access**: All tech staff, product managers, analysts
*   **Restrictions**: Support sees anonymized data only
*   **Audit**: Basic access logging

### WARN Logs
*   **Who can access**: Tech staff, operations, management
*   **Restrictions**: Management gets summary reports only
*   **Audit**: Detailed logging for technical access

### ERROR Logs
*   **Who can access**: All tech roles, management, legal (on request)
*   **Restrictions**: Legal needs approval
*   **Audit**: Full audit with access reason

Log Access Audit
-----------------

### What we track
*   Who accessed (user, role, IP address)
*   When and what they viewed
*   Search queries used
*   Data exported

### Audit storage
*   **Retention**: 2 years
*   **Location**: Separate secure database
*   **Reports**: Monthly automated reports
*   **Alerts**: For suspicious activity