# Zephyrix

Zephyrix is a high-performance web framework for Go, designed to provide developers with a powerful and flexible toolkit for building modern web applications.

## Table of Contents

- [Zephyrix](#zephyrix)
  - [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Installation](#installation)
  - [Quick Start](#quick-start)
  - [Usage Examples](#usage-examples)
    - [Defining Routes](#defining-routes)
    - [Using Middleware](#using-middleware)
    - [Database Operations](#database-operations)
  - [Configuration](#configuration)
  - [Database Integration](#database-integration)
  - [Middleware](#middleware)
  - [Routing](#routing)
  - [SSL/TLS Support](#ssltls-support)
  - [Logging](#logging)
  - [Testing](#testing)
  - [Development](#development)
  - [Contributing](#contributing)
  - [License](#license)
  - [Donation](#donation)

## Features

Zephyrix offers a rich set of features to streamline your web development process:

- **High Performance**: Built on top of the fast and efficient Gin framework
- **Dependency Injection**: Utilizes uber-go/fx for flexible and testable code structure
- **Database Integration**: Seamless integration with MySQL using BeeORM
- **Redis Caching**: Built-in support for Redis caching to boost performance
- **Auto SSL**: Automatic SSL certificate management with Let's Encrypt and ZeroSSL
- **Middleware Support**: Easy-to-use middleware system for request/response processing
- **Flexible Routing**: Intuitive API for defining routes and handlers
- **Configuration Management**: YAML-based configuration with environment variable support
- **Logging**: Configurable logging with multiple output options
- **CORS Support**: Built-in Cross-Origin Resource Sharing (CORS) configuration
- **Graceful Shutdown**: Handles shutdown signals for graceful application termination
- **CLI Commands**: Includes built-in CLI commands for common tasks like database migrations

## Installation

To install Zephyrix, use the following command:

```bash
go get -u go.mamad.dev/zephyrix
```

## Quick Start

Here's a minimal example to get your Zephyrix application up and running:

```go
package main

import (
    "context"
    "go.mamad.dev/zephyrix"
)

func main() {
    app := zephyrix.NewApplication()

    app.Router().GET("/", func(c zephyrix.Context) {
        c.JSON(200, "Hello, Zephyrix!")
    })

    if err := app.Start(context.Background()); err != nil {
        panic(err)
    }
}
```

## Usage Examples

### Defining Routes

```go
app.Router().Group(func(router zephyrix.Router) {
    router.GET("/users", GetUsers)
    router.POST("/users", CreateUser)
    router.GET("/users/:id", GetUser)
}, "/api/v1")
```

### Using Middleware

```go
app.RegisterMiddleware(AuthMiddleware)

app.Router().GET("/protected", ProtectedHandler, "auth")
```

### Database Operations

```go
type User struct {
    ID   uint64 `orm:"table=users;redisCache;localCache"`
    Name string
}

func init() {
    app.Database().RegisterEntity(&User{})
}

func GetUser(c zephyrix.Context) {
    // Use the database to fetch a user
    // Implementation details depend on your specific use case
}
```

## Configuration

Zephyrix uses a YAML configuration file. Here's a sample configuration:

```yaml
log:
  level: "debug"
  outputs: ["stdout"]

server:
  address: ":8000"
  cors:
    enabled: true
    allowed_origins: ["*"]

database:
  pools:
    - name: "default"
      dsn: "root:password@tcp(127.0.0.1:3306)/zephyrix?charset=utf8mb4&parseTime=True&loc=Local"
      max_open_conns: 100
      max_idle_conns: 10
      conn_max_lifetime: "1h"
      cache:
        enabled: true
        size: 10000
      redis:
        enabled: true
        address: "127.0.0.1:6379"
```

## Database Integration

Zephyrix uses BeeORM for database operations, providing an easy-to-use interface for working with MySQL databases and Redis caching.

- BeeORM documents are available here: [BeeORM Documentation](https://beeorm.io/)
- BeeORM's domain has been expired, use the [BeeORM Documentation](https://beeorm-doc-v3-mhwi4tuqq-latolukasz.vercel.app/) from the [repo](https://github.com/latolukasz/beeorm-doc/tree/v3) instead.

## Middleware

Middleware in Zephyrix can be easily created and applied to routes:

```go
func LoggingMiddleware() zephyrix.ZephyrixMiddleware {
    return &customMiddleware{
        name: "logging",
    }
}

type customMiddleware struct {
    name string
}

func (m *customMiddleware) Name() string {
    return m.name
}

func (m *customMiddleware) Handler(args ...any) any {
    return func(c zephyrix.Context) {
        // Middleware logic here
    }
}
```

## Routing

Zephyrix provides a flexible routing system that supports grouping, parameterized routes, and various HTTP methods.

## SSL/TLS Support

Zephyrix includes built-in support for SSL/TLS, including automatic certificate management with Let's Encrypt and ZeroSSL.

## Logging

The framework offers configurable logging with support for multiple outputs, including file, stdout, and potentially database and remote logging.

## Testing

Zephyrix is designed with testability in mind. It includes test utilities and supports dependency injection for easy mocking and testing of components.

## Development

For information on setting up a development environment and contributing to Zephyrix, please refer to the [DEVELOPMENT.md](DEVELOPMENT.md) file in the repository.

## Contributing

We welcome contributions to Zephyrix! Please see our [CONTRIBUTING.md](CONTRIBUTING.md) file for details on how to contribute, our code of conduct, and the process for submitting pull requests.

## License

Zephyrix is released under the MIT License. See the [LICENSE](LICENSE) file for more details.

## Donation

If you find Zephyrix useful and would like to support its development, consider making a donation. Your support helps maintain and improve the project.

- Bitcoin:
- Ethereum:

Thank you for your support!

<!-- this is still partially ai generated, i will be writing an actual readme when i finish the project . :P -->
