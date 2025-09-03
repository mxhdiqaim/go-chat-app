# Go Chat Application

This is a real-time chat application built with Go. It features a backend API for user authentication and a WebSocket-based system for handling real-time, bidirectional communication between clients.

The project is designed with a clean, layered architecture, separating concerns between API handling, business logic, and database access.

## Features

- **User Authentication**: Secure user registration and login endpoints.
- **JWT-based Security**: Protected routes are secured using JSON Web Tokens.
- **Password Hashing**: User passwords are securely hashed using `bcrypt`.
- **Real-time Chat**: Concurrent, real-time messaging via WebSockets in dedicated rooms.
- **Type-Safe Database Access**: Uses `sqlc` to generate fully type-safe Go code from raw SQL.
- **Database Migrations**: Uses `goose` for managing database schema changes.
- **Live API Documentation**: Provides an interactive Swagger UI for all endpoints.

## Technologies Used

- **Go**: Backend language
- **Chi**: HTTP router
- **PostgreSQL**: Database
- **pgx**: Go driver for PostgreSQL
- **sqlc**: Type-safe SQL code generation
- **Gorilla WebSocket**: WebSocket implementation
- **goose**: Database migrations
- **swag**: Automatic Swagger/OpenAPI documentation

## Getting Started (Running Locally)

### 1. Prerequisites

- **Go**: Version 1.21 or later.
- **PostgreSQL**: A running instance. Using [Docker](https://www.docker.com/) is recommended.
- **`goose` CLI**: For database migrations.
- **`swag` CLI**: For generating API documentation.

### 2. Installation & Setup

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/mxhdiqaim/go-chat-app.git
    cd go-chat-app
    ```

2.  **Install Go dependencies:**

    ```bash
    go mod tidy
    ```

3.  **Start a PostgreSQL database (using Docker):**

    ```bash
    docker run --name go-chat-db -p 5432:5432 -e POSTGRES_PASSWORD=password -d postgres
    ```

4.  **Set up environment variables:**
    Copy the example `.env` file and edit it if your database credentials are different.

    ```bash
    cp .env.example .env
    ```

    The default `.env` file is already configured for the Docker command above.

5.  **Install CLI tools:**

    ```bash
    # Install goose for migrations
    go install github.com/pressly/goose/v3/cmd/goose@latest

    # Install swag for API docs
    go install github.com/swaggo/swag/cmd/swag@latest
    ```

6.  **Run database migrations:**
    This command reads your `DATABASE_URL` from the `.env` file (if you are using a tool that loads it) or you can paste the string directly.
    ```bash
    goose -dir ./migrations postgres "postgresql://postgres:password@localhost:5432/postgres" up
    ```

### 3. Running the Application

1.  **Generate API Documentation:**
    The `swag` tool reads the comments in your code to generate the documentation files.

    ```bash
    swag init -g cmd/api/main.go
    ```

2.  **Start the server:**
    ```bash
    go run cmd/api/main.go
    ```
    The server will start on `http://localhost:8080`.

## API Documentation

Once the server is running, you can view the interactive Swagger API documentation by navigating to:

**[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)**

## Deployment Note

The `render_build.sh` script is included for automated deployments on platforms like Render **Deployment in progress.**
