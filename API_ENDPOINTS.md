make run se# API Endpoints

This document describes the available API endpoints for the Mental Health Counselling Client Management System.

## Base URL

- Local: `http://localhost:8080`
- Port can be configured via `HTTP_PORT` environment variable (default: 8080)

## Endpoints

### Health Check

**GET** `/health`

Check if the server is running.

**Response:**
```json
{
  "status": "ok",
  "message": "Server is running"
}
```

---

### Get All Clients

**GET** `/api/clients`

Retrieves all clients from the database.

**Response:**
```json
[
  {
    "id": "client-001",
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@example.com",
    "phone": "555-0101",
    "date_of_birth": "1985-03-15",
    "address": "123 Main St, Anytown, ST 12345",
    "emergency_contact_name": "Jane Doe",
    "emergency_contact_phone": "555-0102",
    "status": "active",
    "created_at": "2025-11-11T18:00:00Z",
    "updated_at": "2025-11-11T18:00:00Z"
  }
]
```

---

### Get Client by ID

**GET** `/api/clients/{id}`

Retrieves a single client by their ID.

**Parameters:**
- `id` (path parameter) - The client ID (e.g., `client-001`)

**Example:**
```
GET /api/clients/client-001
```

**Response:**
```json
{
  "id": "client-001",
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@example.com",
  "phone": "555-0101",
  "date_of_birth": "1985-03-15",
  "address": "123 Main St, Anytown, ST 12345",
  "emergency_contact_name": "Jane Doe",
  "emergency_contact_phone": "555-0102",
  "status": "active",
  "created_at": "2025-11-11T18:00:00Z",
  "updated_at": "2025-11-11T18:00:00Z"
}
```

**Error Response (404):**
```json
Client not found: client-999
```

---

### Get Active Clients

**GET** `/api/clients/active`

Retrieves all clients with status "active".

**Response:**
```json
[
  {
    "id": "client-001",
    "first_name": "John",
    "last_name": "Doe",
    "status": "active",
    ...
  }
]
```

---

### Get Inactive Clients

**GET** `/api/clients/inactive`

Retrieves all clients with status "inactive".

**Response:**
```json
[
  {
    "id": "client-004",
    "first_name": "Emily",
    "last_name": "Williams",
    "status": "inactive",
    ...
  }
]
```

---

## Testing with Postman

1. **Start the server:**
   ```bash
   make run-server
   ```

2. **Import these requests into Postman:**

   - **Health Check:**
     - Method: `GET`
     - URL: `http://localhost:8080/health`

   - **Get All Clients:**
     - Method: `GET`
     - URL: `http://localhost:8080/api/clients`

   - **Get Client by ID:**
     - Method: `GET`
     - URL: `http://localhost:8080/api/clients/client-001`

   - **Get Active Clients:**
     - Method: `GET`
     - URL: `http://localhost:8080/api/clients/active`

   - **Get Inactive Clients:**
     - Method: `GET`
     - URL: `http://localhost:8080/api/clients/inactive`

3. **All responses are in JSON format** and can be viewed in Postman's response body.

## Error Responses

All endpoints return appropriate HTTP status codes:

- `200 OK` - Success
- `400 Bad Request` - Invalid request (e.g., missing ID)
- `404 Not Found` - Resource not found
- `405 Method Not Allowed` - Wrong HTTP method
- `500 Internal Server Error` - Server error

Error responses include error messages in the response body.

