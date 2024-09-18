# Zephyrix Development

This document outlines the comprehensive development roadmap for the Zephyrix backend web framework. Below is a list of features to be implemented over time, along with their current status.

## things to do before

- install `air` for hot reloading and easier development

```bash
go install github.com/air-verse/air@latest
```

then use `air` to run a test server (./examples/simple-test.go)

```bash
make dev
```

## TODO

1. !! A very important feature: MOVE EVERY **SYNC / MUTEX LOCK** TO A **REDIS LOCK** for multi-instance deployments.
2. 

## thigs that are working (implemented features)

> here is a list of things that are implemented, and working in the Zephyrix web framework.
> they're not fully tested or functonal, but they've been implemented and work fine.

- [x] Cobra CLI integration (globally)
- [x] Configuration Management using `viper`
- [x] Logger integration [`zephyrix.Logger`](zephyrix_logger.go#Logger) is usable and is of type [`zephyrix.ZephyrixLogger`](zephyrix_logger.go#ZephyrixLogger)
- [x] a working HTTP Server using `gin-gonic` as the underlying engine
- [x] Dependency Injection using [`fx`](https://github.com/uber-go/fx) as the underlying DI container. (might be replaced with [`wire`](https://github.com/google/wire) in the future)
  - [x] Dependency Injection
  - [x] Service container
  - [x] Lifecycle management
- [x] Beeorm implementation for ORM and database management (Should be replaced with a custom ORM in the future)
  - [x] Database Integration
  - [x] ORM or query builder
  - [x] Connection pooling
  - [x] Migration support
- [x] Dependency Injected Router / Route Manager
- [x] Custom context support in routes (WIP: supports the context, but the context is not yet fully functional)
- [x] CORS handling -> handled by the `gin-gonic` engine (uses zephyrix configuration to easily enable/disable - handle)
- [x] Middleware Support
  - [x] Custom Middlewares
  - [x] Named Middlewares
        // todo: implement custom middlewares from string
        // ? e.g: "auth:(some args)" example: "auth:jwt" or "auth:jwt,admin"
        // ? or "ratelimit:10:1m" (10 requests per minute)
  - [x] Custom error handling (WIP: not fully implemented)
- [ ] Authentication
  - [ ] User authentication
  - [ ] Role-based access control
  - [ ] JWT support
  - [ ] Session creation and storage
  - [ ] Cookie handling
  - [ ] Administration Ui
  - [ ] User management
  - [ ] Full OAuth2 support
- [ ] API Documentation
  - [ ] Automatic API docs generation
  - [ ] Swagger integration
  - [ ] Postman collection generation

## TODO

- [ ] Named routes (WIP: partially implemented)

## Core Features

- [ ] Authentication

- [ ] Request Parsing

  - [ ] Query parameters
  - [ ] Form data
  - [ ] JSON body
  - [ ] File uploads

- [ ] Response Handling

  - [ ] JSON responses
  - [ ] Streaming responses

- [ ] Caching

  - [ ] In-memory caching
  - [ ] Distributed caching support

- [ ] Logging and Error Handling

  - [ ] Structured logging
  - [ ] Error reporting

- [ ] Configuration Management

  - [ ] Environment-based configs
  - [ ] Configuration file parsing

- [ ] RESTful API Support

  - [ ] Resource-based routing
  - [ ] Content negotiation

- [ ] Security Features

  - [ ] CSRF protection
  - [ ] XSS prevention
  - [ ] SQL injection prevention

- [ ] Validation

  - [ ] Input validation
  - [ ] Custom validation rules

- [ ] Testing Utilities

  - [ ] Unit testing helpers
  - [ ] Integration testing support

- [ ] CLI Tools

  - [ ] Project scaffolding
  - [ ] Code generation

- [ ] Static File Serving

  - [ ] Efficient serving of assets
  - [ ] Cache control

- [ ] Rate Limiting

  - [ ] Request throttling
  - [ ] Custom rate limit rules

- [ ] Background Job Processing

  - [ ] Task queues
  - [ ] Scheduled jobs

- [ ] API Documentation

  - [ ] Automatic API docs generation

- [ ] Monitoring and Metrics
  - [ ] Performance tracking
  - [ ] Health checks

## Advanced Backend Features

- [ ] Advanced Database Features

  - [ ] Multi-database support (SQL and NoSQL)
  - [ ] Database migration versioning
  - [ ] Query optimization tools

- [ ] Caching Strategies

  - [ ] Distributed caching (e.g., Redis integration)
  - [ ] Cache invalidation patterns

- [ ] Message Queue Integration

  - [ ] Support for various message brokers (e.g., RabbitMQ, Apache Kafka)
  - [ ] Pub/Sub pattern implementation

- [ ] Microservices Toolkit

  - [ ] Service discovery
  - [ ] Circuit breaker implementation
  - [ ] Distributed tracing

- [ ] API Gateway Functionality

  - [ ] Request/Response transformation
  - [ ] API composition

- [ ] Advanced Logging and Monitoring

  - [ ] Integration with popular monitoring tools (e.g., Prometheus, Grafana)

- [ ] Serverless Deployment Support

  - [ ] Adapters for major serverless platforms (AWS Lambda, Google Cloud Functions)

- [ ] gRPC Support

  - [ ] Protocol buffer integration
  - [ ] gRPC-Web support for browser clients

- [ ] GraphQL Server Implementation

  - [ ] Schema definition tools
  - [ ] Resolver framework

- [ ] Webhook Management

  - [ ] Webhook creation and management
  - [ ] Retry mechanisms for failed webhook deliveries

- [ ] Task Scheduling

  - [ ] Cron-like job scheduling
  - [ ] Distributed task execution

- [ ] Data Streaming

  - [ ] Support for real-time data processing
  - [ ] Integration with streaming platforms (e.g., Apache Kafka)

- [ ] Cryptography and Security Tools

  - [ ] Encryption utilities
  - [ ] Token-based authentication (JWT, OAuth)

- [ ] Containerization Support

  - [x] Docker integration
  - [ ] Kubernetes deployment helpers

- [ ] Event Sourcing and CQRS
  - [ ] Event store implementation
  - [ ] Command and query segregation utilities

## Contributing

We welcome contributions to the Zephyrix web framework. Please refer to our CONTRIBUTING.md(CONTRIBUTING.MD) file for guidelines on how to contribute.
