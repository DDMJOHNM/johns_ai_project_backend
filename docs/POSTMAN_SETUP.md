# Postman Setup for API Gateway

## API Gateway URL
```
https://mos5j2g72f.execute-api.us-east-1.amazonaws.com/prod
```

## Endpoints

### 1. Health Check
- **Method:** `GET`
- **URL:** `https://mos5j2g72f.execute-api.us-east-1.amazonaws.com/prod/health`
- **Headers:** None required
- **Body:** None

### 2. Create Client (POST /api/clients/add)
- **Method:** `POST`
- **URL:** `https://mos5j2g72f.execute-api.us-east-1.amazonaws.com/prod/api/clients/add`
- **Headers:**
  - `Content-Type: application/json`
- **Body (raw JSON):**
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@example.com",
  "phone": "555-0101",
  "date_of_birth": "1990-01-15",
  "address": "123 Main St, City, ST 12345",
  "emergency_contact_name": "Jane Doe",
  "emergency_contact_phone": "555-0102"
}
```

**Minimal required fields:**
```json
{
  "first_name": "Test",
  "last_name": "User",
  "email": "test@example.com"
}
```

### 3. Get All Clients
- **Method:** `GET`
- **URL:** `https://mos5j2g72f.execute-api.us-east-1.amazonaws.com/prod/api/clients`
- **Headers:** None required
- **Body:** None

### 4. Get Client by ID
- **Method:** `GET`
- **URL:** `https://mos5j2g72f.execute-api.us-east-1.amazonaws.com/prod/api/clients/{id}`
  - Replace `{id}` with actual client ID (e.g., `client-001`)
- **Headers:** None required
- **Body:** None

### 5. Get Active Clients
- **Method:** `GET`
- **URL:** `https://mos5j2g72f.execute-api.us-east-1.amazonaws.com/prod/api/clients/active`
- **Headers:** None required
- **Body:** None

### 6. Get Inactive Clients
- **Method:** `GET`
- **URL:** `https://mos5j2g72f.execute-api.us-east-1.amazonaws.com/prod/api/clients/inactive`
- **Headers:** None required
- **Body:** None

## Postman Setup Steps

1. **Open Postman** and create a new request

2. **Set the HTTP Method:**
   - Select `POST` from the dropdown

3. **Enter the URL:**
   - `https://mos5j2g72f.execute-api.us-east-1.amazonaws.com/prod/api/clients/add`

4. **Add Headers:**
   - Click on the "Headers" tab
   - Add header:
     - Key: `Content-Type`
     - Value: `application/json`

5. **Add Request Body:**
   - Click on the "Body" tab
   - Select "raw"
   - Select "JSON" from the dropdown (next to "raw")
   - Paste the JSON body:
   ```json
   {
     "first_name": "Test",
     "last_name": "User",
     "email": "test@example.com"
   }
   ```

6. **Send the Request:**
   - Click the "Send" button
   - You should see a response with status `201 Created` and the created client object

## Expected Response

**Success (201 Created):**
```json
{
  "id": "client-20260106042113",
  "first_name": "Test",
  "last_name": "User",
  "email": "test@example.com",
  "phone": "",
  "date_of_birth": "",
  "address": "",
  "emergency_contact_name": "",
  "emergency_contact_phone": "",
  "status": "active",
  "created_at": "2026-01-06T04:21:13Z",
  "updated_at": "2026-01-06T04:21:13Z"
}
```

**Error (400 Bad Request):**
```json
{
  "error": "Validation failed",
  "message": "Email is required"
}
```

## Quick Test Command (cURL)

If you want to test from command line first:
```bash
curl -X POST "https://mos5j2g72f.execute-api.us-east-1.amazonaws.com/prod/api/clients/add" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Test",
    "last_name": "User",
    "email": "test@example.com"
  }'
```

