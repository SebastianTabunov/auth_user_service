# Auth User Service

Microservice for user authentication, management, and order processing.

## Quick Start

```bash
docker-compose up -d
```
Application will be available at http://localhost:8080

## Configuration

Main environment variables:

```bash
PORT=8080
DB_HOST=db
DB_USER=user
DB_PASSWORD=password
DB_NAME=auth_service
JWT_SECRET=your-jwt-secret-key
CORS_ALLOWED_ORIGINS=*
```

## API Endpoints

### Authentication

POST /auth/register - User registration
POST /auth/login - User login
POST /auth/refresh - Token refresh
POST /auth/logout - User logout
Users

GET /api/user/profile - Get user profile
PUT /api/user/profile - Update user profile
Orders

GET /api/orders - Get user orders
GET /api/orders/{id} - Get order details
POST /api/orders - Create new order
System

GET /health - Health check

## Request Examples

### Registration

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

### Login

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'
```

### Get Profile

```bash
curl -X GET http://localhost:8080/api/user/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Technologies

Go • PostgreSQL • Redis • Docker • JWT

## Migrations

Database migrations run automatically when container starts.
