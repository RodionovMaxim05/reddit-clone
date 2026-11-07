# Reddit Clone Backend

## Description

A Go-based RESTful backend implementation for the Asperitas Reddit clone frontend. The service provides complete API functionality for user authentication, post management (text and link posts), commenting, and voting.

## Architecture

The project follows the [golang-standards/project-layout](https://github.com/golang-standards/project-layout) guidelines:

* **`cmd/redditclone`** — Application entry point (`main.go`). Handles dependency injection, router initialization, and HTTP server startup.
* **`internal/domain`** — Pure domain entities and domain errors, decoupled from transport and storage tags.
* **`internal/repository`** — Thread-safe in-memory storage implementation using `sync.RWMutex` to protect shared state during concurrent HTTP requests.
* **`internal/controller`** — HTTP transport layer handling routing, requests, authentication middleware, and error responses.
* **`internal/controller/dto`** — Request/Response Data Transfer Objects (DTOs) with JSON tags and domain mappers.
* **`internal/pkg/auth`** — JWT token management module using the **HS256** signing algorithm.
* **`static`** — Pre-compiled Single Page Application (SPA) frontend assets.

```text
.
├── api
│   └── openapi.yaml
├── cmd
│   └── redditclone
│       └── main.go
├── internal
│   ├── controller
│   │   ├── auth_middleware.go
│   │   ├── dto
│   │   │   ├── auth_dto.go
│   │   │   ├── comment_dto.go
│   │   │   └── post_dto.go
│   │   ├── handler_utils.go
│   │   ├── post_handler.go
│   │   └── user_handler.go
│   ├── domain
│   │   ├── errors.go
│   │   ├── post.go
│   │   └── user.go
│   ├── pkg
│   │   └── auth
│   │       └── jwt.go
│   └── repository
│       ├── post_memory.go
│       └── user_memory.go
└── static
    ├── css
    ├── html
    └── js
```

## Requirements

* **Go 1.27+**
* Web Browser or `curl`

## Quick Start

1. Clone the repository and navigate to the project directory:

```bash
cd reddit-clone
```

1. Download dependencies:

```bash
go mod download
```

1. Run the application server:

```bash
go run ./cmd/redditclone
```

1. Open your browser and navigate to:

```text
http://localhost:8080/
```

## Configuration

The application can be configured via environment variables:

| Environment Variable | Description | Default Value |
| --- | --- | --- |
| `JWT_SECRET` | Secret key used for signing and verifying JWT tokens | `default_secret_key_change_in_prod` |

Example starting the server with a custom JWT secret:

```bash
JWT_SECRET="$(openssl rand -hex 32)" go run ./cmd/redditclone
```

## API Endpoints

### Summary Table

| Method & Path | Authorization | Description |
| --- | --- | --- |
| `POST /api/register` | None | Register a new user and receive a JWT token. |
| `POST /api/login` | None | Authenticate user credentials and receive a JWT token. |
| `GET /api/posts` | None | Fetch all posts sorted by `score` in descending order. |
| `GET /api/posts/{category}` | None | Fetch posts filtered by category. |
| `GET /api/post/{postId}` | None | Get detailed post information (increments `views` count). |
| `POST /api/posts` | Bearer Token | Create a new post (text or link). |
| `DELETE /api/post/{postId}` | Bearer Token | Delete a post (author or admin only). |
| `POST /api/post/{postId}` | Bearer Token | Add a comment to a specific post. |
| `DELETE /api/post/{postId}/{commentId}` | Bearer Token | Delete a comment (comment author, post author, or admin). |
| `GET /api/post/{postId}/upvote` | Bearer Token | Cast an upvote (+1) on a post. |
| `GET /api/post/{postId}/downvote` | Bearer Token | Cast a downvote (-1) on a post. |
| `GET /api/post/{postId}/unvote` | Bearer Token | Remove user's vote from a post. |
