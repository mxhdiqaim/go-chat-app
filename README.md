# Go Chat Application

This is a real-time chat application built with Go. It features a backend API for user authentication and a WebSocket-based system for handling real-time, bidirectional communication between clients.

The project is designed with a clean, layered architecture, separating concerns between API handling, business logic, and database access.

## Features

- **User Authentication**: Secure user registration and login endpoints.
- **JWT-based Security**: Protected routes are secured using JSON Web Tokens.
- **Password Hashing**: User passwords are securely hashed using `bcrypt`.
- **Real-time Chat**: Concurrent, real-time messaging via WebSockets.
- **Public and Private Messaging**: Supports broadcasting messages to all users and sending private messages to specific users.
- **Type-Safe Database Access**: Uses `sqlc` to generate fully type-safe Go
