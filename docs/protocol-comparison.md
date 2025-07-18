Protocol Comparison Report: HTTP vs gRPC
========================================

Executive Summary
-----------------

This report presents a comprehensive performance comparison between HTTP and gRPC protocols for a weather API endpoint. 

Test Environment
----------------

### API Endpoint

*   **HTTP**: GET http://localhost:8080/api/weather/current?city=Kyiv

*   **gRPC**: weather.WeatherService.GetCurrentWeather with payload {"city": "Kyiv"}


### Testing Tools

*   **HTTP**: hey - HTTP load testing tool

*   **gRPC**: ghz - gRPC load testing tool


### Test Scenarios

#### Test 1: Fixed Request Count

*   **Objective**: Measure execution time for a fixed number of requests

*   **Parameters**: 100,000 requests with 100 concurrent connections

*   **Focus**: Total execution time and throughput analysis


#### Test 2: Fixed Duration

*   **Objective**: Measure maximum throughput within a time constraint

*   **Parameters**: 10-second duration with 10 concurrent connections

*   **Focus**: Request volume and sustained performance


Test Results Analysis
---------------------

### Test 1: Fixed Request Count (100,000 requests, 100 connections)

#### HTTP Performance

*   **Total Execution Time**: 22.61 seconds

*   **Throughput**: 4,423.60 requests/second

*   **Average Latency**: 22.6 ms

*   **P95 Latency**: 35.5 ms

*   **P99 Latency**: 82.8 ms

*   **Success Rate**: 100% (all 200 OK responses)


#### gRPC Performance

*   **Total Execution Time**: 12.07 seconds

*   **Throughput**: 8,285.92 requests/second

*   **Average Latency**: 11.94 ms

*   **P95 Latency**: 15.47 ms

*   **P99 Latency**: 22.57 ms

*   **Success Rate**: 100% (all OK responses)


#### Test 1 Performance Comparison

| Metric         | HTTP           | gRPC          | gRPC Advantage        |
|----------------|----------------|---------------|-----------------------|
| Execution Time | 22.61s         | 12.07s        | 46.6% faster          |
| Throughput     | 4,423.60 RPS   | 8,285.92 RPS  | 87.3% higher          |
| Average Latency| 22.6ms         | 11.94ms       | 47.2% lower           |
| P95 Latency    | 35.5ms         | 15.47ms       | 56.4% lower           |
| P99 Latency    | 82.8ms         | 22.57ms       | 72.7% lower           |


### Test 2: Fixed Duration (10 seconds, 10 connections)

#### HTTP Performance

*   **Total Requests**: 45,156

*   **Throughput**: 4,514.52 requests/second

*   **Average Latency**: 2.2 ms

*   **P95 Latency**: 3.5 ms

*   **P99 Latency**: 7.0 ms

*   **Success Rate**: 100% (all 200 OK responses)


#### gRPC Performance

*   **Total Requests**: 76,881

*   **Throughput**: 7,687.86 requests/second

*   **Average Latency**: 1.21 ms

*   **P95 Latency**: 2.19 ms

*   **P99 Latency**: 3.24 ms

*   **Success Rate**: 99.99% (7 unavailable responses out of 76,881)


#### Test 2 Performance Comparison

| Metric          | HTTP          | gRPC         | gRPC Advantage       |
|-----------------|---------------|--------------|----------------------|
| Total Requests  | 45,156        | 76,881       | 70.3% more requests   |
| Throughput      | 4,514.52 RPS  | 7,687.86 RPS | 70.3% higher          |
| Average Latency | 2.2ms         | 1.21ms       | 45.0% lower           |
| P95 Latency     | 3.5ms         | 2.19ms       | 37.4% lower           |
| P99 Latency     | 7.0ms         | 3.24ms       | 53.7% lower           |

Key Findings
------------

### Performance Advantages of gRPC

1.  **Superior Throughput**: gRPC consistently delivered 70-87% higher throughput across both test scenarios

2.  **Lower Latency**: gRPC showed 37-72% lower latency at all percentiles

3.  **Faster Execution**: Fixed request count completed 46.6% faster with gRPC

4.  **Better Resource Utilization**: Higher request processing efficiency under load

Conclusion
----------

The performance comparison reveals gRPC's significant advantages in terms of throughput, latency, and execution speed. gRPC delivered 70-87% higher throughput and 37-72% lower latency across all test scenarios. The binary protocol efficiency, HTTP/2 multiplexing, and optimized serialization contribute to these performance gains.

However, the choice between HTTP and gRPC should consider factors beyond raw performance, including team expertise, ecosystem compatibility, and specific use case requirements. For performance-critical applications, especially in microservices architectures, gRPC presents a compelling advantage. For public APIs and web applications requiring broad compatibility, HTTP/REST remains the pragmatic choice.