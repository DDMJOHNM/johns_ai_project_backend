# Authentication System

This document describes the JWT-based authentication system implemented for the John AI Project API.

## Overview

The API uses **JWT (JSON Web Tokens)** for authentication. Users must register an account, login to receive a token, and include that token in subsequent requests to access protected endpoints.

## Features

✅ **User Registration** - Create new user accounts  
✅ **Login with JWT** - Authenticate and receive a 24-hour token  
✅ **Password Hashing** - bcrypt encryption for secure password storage  
✅ **Protected Routes** - All client endpoints require authentication  
✅ **Username or Email Login** - Users can login with either credential  
✅ **Role-based Access** - Support for user roles (admin, user)  

## Database Schema

The authentication system uses the existing `users` DynamoDB table with the following structure:

```
Table: users
Primary Key: id (String)

Attributes:
- id: String (UUID)
- username: String (unique)
- email: String (unique)
- password_hash: String (bcrypt hash)
- first_name: String
- last_name: String
- role: String ("admin" | "user")
- is_active: Boolean
- created_at: String (ISO 8601)
- updated_at: String (ISO 8601)

Global Secondary Indexes:
- username-index: Query by username
- email-index: Query by email
- role-index: Query by role
```

## API Endpoints

### Public Endpoints (No Authentication Required)

#### 1. Register New User

**POST** `/api/auth/register`

Create a new user account.

**Request Body:**
```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "SecurePass123!",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Response (201 Created):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "johndoe",
    "email": "john@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "role": "user",
    "is_active": true,
    "created_at": "2026-01-27T12:00:00Z",
    "updated_at": "2026-01-27T12:00:00Z"
  }
}
```

**Validation:**
- All fields are required
- Password must be at least 8 characters
- Username and email must be unique

**Error Responses:**
- `400` - Invalid request or validation error
- `400` - User with email already exists
- `400` - Username is already taken

---

#### 2. Login

**POST** `/api/auth/login`

Authenticate and receive a JWT token.

**Request Body:**
```json
{
  "login": "johndoe",
  "password": "SecurePass123!"
}
```

**Note:** The `login` field accepts either username or email.

**Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "johndoe",
    "email": "john@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "role": "user",
    "is_active": true,
    "created_at": "2026-01-27T12:00:00Z",
    "updated_at": "2026-01-27T12:00:00Z"
  }
}
```

**Error Responses:**
- `400` - Missing login or password
- `401` - Invalid credentials
- `401` - Account is disabled

---

### Protected Endpoints (Authentication Required)

All protected endpoints require the `Authorization` header with a Bearer token:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### 3. Get Current User

**GET** `/api/auth/me`

Get information about the currently authenticated user.

**Headers:**
```
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "johndoe",
  "email": "john@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "role": "user",
  "is_active": true,
  "created_at": "2026-01-27T12:00:00Z",
  "updated_at": "2026-01-27T12:00:00Z"
}
```

**Error Responses:**
- `401` - Missing or invalid token
- `404` - User not found

---

#### 4. Client Endpoints (All Protected)

All client management endpoints now require authentication:

- `GET /api/clients` - Get all clients
- `GET /api/clients/{id}` - Get client by ID
- `GET /api/clients/active` - Get active clients
- `GET /api/clients/inactive` - Get inactive clients
- `POST /api/clients/add` - Create new client

See the main API documentation for details on these endpoints.

---

## Usage Examples

### Example 1: Register and Access Protected Endpoint

```bash
# 1. Register a new user
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "email": "john@example.com",
    "password": "SecurePass123!",
    "first_name": "John",
    "last_name": "Doe"
  }'

# Response includes token:
# {"token": "eyJhbGc...", "user": {...}}

# 2. Use the token to access protected endpoint
export TOKEN="eyJhbGc..."

curl http://localhost:8080/api/clients \
  -H "Authorization: Bearer $TOKEN"
```

### Example 2: Login with Username

```bash
# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "login": "johndoe",
    "password": "SecurePass123!"
  }'

# Extract and use token
export TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"login":"johndoe","password":"SecurePass123!"}' \
  | jq -r '.token')

curl http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

### Example 3: Login with Email

```bash
# You can also login with email
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "login": "john@example.com",
    "password": "SecurePass123!"
  }'
```

---

## JWT Token Details

### Token Contents (Claims)

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "johndoe",
  "email": "john@example.com",
  "role": "user",
  "exp": 1706443200,
  "iat": 1706356800,
  "iss": "john-ai-project"
}
```

### Token Properties

- **Algorithm:** HS256 (HMAC SHA-256)
- **Expiration:** 24 hours from issue time
- **Issuer:** john-ai-project

### Security Considerations

1. **Secrets:** The JWT secret is configured via the `JWT_SECRET` environment variable
2. **HTTPS:** In production, always use HTTPS to prevent token interception
3. **Storage:** Store tokens securely on the client side (secure cookies or localStorage)
4. **Expiration:** Tokens expire after 24 hours - users need to login again
5. **Password Requirements:** Minimum 8 characters (add more validation as needed)

---

## Configuration

### Environment Variables

```bash
# Required for production
JWT_SECRET=your-super-secret-key-here

# DynamoDB configuration
AWS_REGION=us-east-1
DYNAMODB_ENDPOINT=  # Leave empty for AWS, set for local development

# Server configuration
HTTP_PORT=8080
```

### Setting JWT Secret

**Development:**
```bash
export JWT_SECRET="dev-secret-key-change-in-production"
./bin/server
```

**Production (EC2):**

Add to the systemd service file:
```ini
Environment="JWT_SECRET=your-production-secret-key-here"
```

Or set via SSM Parameter Store:
```bash
aws ssm put-parameter \
  --name "/john-ai-project/jwt-secret" \
  --value "your-production-secret-key-here" \
  --type "SecureString"
```

---

## Error Responses

All authentication errors follow this format:

```json
{
  "error": "Error type",
  "message": "Detailed error message"
}
```

### Common Error Codes

| Status Code | Description |
|-------------|-------------|
| 400 | Bad Request - Invalid input or validation error |
| 401 | Unauthorized - Missing, invalid, or expired token |
| 404 | Not Found - Resource doesn't exist |
| 500 | Internal Server Error - Server-side error |

---

## Testing Authentication

### Using Postman

1. **Register/Login:**
   - Create a POST request to `/api/auth/register` or `/api/auth/login`
   - Copy the `token` from the response

2. **Access Protected Endpoints:**
   - Create a new request to any protected endpoint
   - Go to **Authorization** tab
   - Select **Bearer Token**
   - Paste your token

3. **Save Token as Environment Variable:**
   - In Postman, add a test script to your login request:
   ```javascript
   pm.environment.set("auth_token", pm.response.json().token);
   ```
   - Use `{{auth_token}}` in subsequent requests

### Using curl

```bash
# Save token to file
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"login":"johndoe","password":"SecurePass123!"}' \
  | jq -r '.token' > token.txt

# Use token from file
curl http://localhost:8080/api/clients \
  -H "Authorization: Bearer $(cat token.txt)"
```

---

## Migration Guide

If you have existing API clients, update them to:

1. **Add Registration Flow:**
   - Users must create an account before accessing the API

2. **Add Login Flow:**
   - Call `/api/auth/login` to get a token
   - Store the token securely

3. **Update API Calls:**
   - Add `Authorization: Bearer <token>` header to all `/api/clients/*` requests

4. **Handle 401 Errors:**
   - Token expired → Prompt user to login again
   - Invalid token → Redirect to login page

---

## Future Enhancements

Potential improvements to consider:

- [ ] Password reset functionality
- [ ] Email verification
- [ ] Refresh tokens (longer-lived sessions)
- [ ] Two-factor authentication (2FA)
- [ ] OAuth integration (Google, GitHub, etc.)
- [ ] Rate limiting on login attempts
- [ ] Account lockout after failed attempts
- [ ] Password strength requirements
- [ ] Role-based permissions (admin-only endpoints)
- [ ] User management endpoints (admin features)

---

## Troubleshooting

### "Authorization header required"

Make sure you're including the header in your request:
```bash
curl http://localhost:8080/api/clients \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### "Invalid or expired token"

Your token has expired (24 hours). Login again to get a new token.

### "User with this email already exists"

The email is already registered. Use `/api/auth/login` instead or register with a different email.

### "Invalid credentials"

Check your username/email and password are correct. Note that passwords are case-sensitive.

### "Table not found" errors

Make sure the users table exists in DynamoDB:
```bash
make create-db
# or
go run cmd/create-db/main.go
```

---

## Architecture

### Components

```
┌─────────────────────────────────────────────┐
│              Client Application              │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│              API Router                      │
│  - Public routes (register, login)          │
│  - Protected routes (with AuthMiddleware)   │
└──────────────────┬──────────────────────────┘
                   │
      ┌────────────┴────────────┐
      ▼                         ▼
┌─────────────┐         ┌──────────────┐
│ AuthHandler │         │ClientHandler │
└──────┬──────┘         └──────┬───────┘
       │                       │
       ▼                       ▼
┌─────────────┐         ┌──────────────┐
│ AuthService │         │ClientService │
└──────┬──────┘         └──────┬───────┘
       │                       │
       ▼                       ▼
┌────────────────┐    ┌──────────────────┐
│ UserRepository │    │ ClientRepository │
└────────┬───────┘    └────────┬─────────┘
         │                     │
         └──────────┬──────────┘
                    ▼
           ┌─────────────────┐
           │  DynamoDB        │
           │  - users table   │
           │  - clients table │
           └──────────────────┘
```

### Request Flow

1. **Registration/Login:**
   - Client → AuthHandler → AuthService → UserRepository → DynamoDB
   - Response includes JWT token

2. **Protected Endpoint:**
   - Client sends request with Bearer token
   - AuthMiddleware validates token
   - If valid, adds user context and forwards to handler
   - Handler processes request normally

---

For more information, see:
- [Main README](../README.md)
- [Postman Setup](./POSTMAN_SETUP.md)
- [Debugging Guide](./DEBUGGING.md)

