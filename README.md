# Chirpy API Documentation

Chirpy is a social media API that allows users to create accounts, post short messages (chirps), and manage their profiles. This documentation covers all available endpoints and their usage.

## Base URL
```
http://localhost:8080
```

## Authentication

The API uses JWT (JSON Web Tokens) for authentication. Most endpoints require authentication via Bearer tokens in the Authorization header.

### Authorization Header Format
```
Authorization: Bearer <your-jwt-token>
```

## Data Models

### User
```json
{
  "id": "uuid",
  "created_at": "timestamp",
  "updated_at": "timestamp", 
  "email": "string",
  "is_chirpy_red": "boolean"
}
```

### Chirp
```json
{
  "id": "uuid",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "body": "string",
  "user_id": "uuid"
}
```

## API Endpoints

### Health Check

#### GET /api/healthz
Check if the API is running.

**Response:**
- **200 OK**: Returns "OK" if the service is healthy

**Example:**
```bash
curl -X GET http://localhost:8080/api/healthz
```

---

### User Management

#### POST /api/users
Create a new user account.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "your-password"
}
```

**Response:**
- **201 Created**: User created successfully
- **400 Bad Request**: Invalid request payload
- **500 Internal Server Error**: Failed to create user

**Example:**
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securepassword123"
  }'
```

#### PUT /api/users
Update user information (requires authentication).

**Headers:**
```
Authorization: Bearer <jwt-token>
```

**Request Body:**
```json
{
  "email": "newemail@example.com",
  "password": "newpassword"
}
```

**Response:**
- **200 OK**: User updated successfully
- **400 Bad Request**: Invalid request payload
- **401 Unauthorized**: Invalid or missing token
- **500 Internal Server Error**: Failed to update user

**Example:**
```bash
curl -X PUT http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-jwt-token>" \
  -d '{
    "email": "newemail@example.com",
    "password": "newpassword123"
  }'
```

---

### Authentication

#### POST /api/login
Authenticate a user and receive JWT tokens.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "your-password"
}
```

**Response:**
- **200 OK**: Login successful, returns user data with tokens
- **400 Bad Request**: Invalid request payload
- **401 Unauthorized**: Incorrect email or password
- **500 Internal Server Error**: Server error

**Response Body:**
```json
{
  "id": "uuid",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "email": "user@example.com",
  "token": "jwt-access-token",
  "refresh_token": "refresh-token",
  "is_chirpy_red": false
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securepassword123"
  }'
```

#### POST /api/refresh
Refresh an access token using a refresh token.

**Headers:**
```
Authorization: Bearer <refresh-token>
```

**Response:**
- **200 OK**: New access token generated
- **401 Unauthorized**: Invalid, expired, or revoked refresh token
- **500 Internal Server Error**: Failed to create JWT token

**Response Body:**
```json
{
  "token": "new-jwt-access-token"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/refresh \
  -H "Authorization: Bearer <your-refresh-token>"
```

#### POST /api/revoke
Revoke a refresh token.

**Headers:**
```
Authorization: Bearer <refresh-token>
```

**Response:**
- **204 No Content**: Token revoked successfully
- **401 Unauthorized**: Invalid refresh token
- **500 Internal Server Error**: Failed to revoke token

**Example:**
```bash
curl -X POST http://localhost:8080/api/revoke \
  -H "Authorization: Bearer <your-refresh-token>"
```

---

### Chirps (Posts)

#### GET /api/chirps
Get all chirps or chirps by a specific author.

**Query Parameters:**
- `author_id` (optional): UUID of the author to filter chirps by

**Response:**
- **200 OK**: Returns array of chirps
- **400 Bad Request**: Invalid author_id format
- **404 Not Found**: No chirps found for the specified author
- **500 Internal Server Error**: Failed to retrieve chirps

**Example:**
```bash
# Get all chirps
curl -X GET http://localhost:8080/api/chirps

# Get chirps by specific author
curl -X GET "http://localhost:8080/api/chirps?author_id=550e8400-e29b-41d4-a716-446655440000"
```

#### GET /api/chirps/{chirpID}
Get a specific chirp by ID.

**Path Parameters:**
- `chirpID`: UUID of the chirp

**Response:**
- **200 OK**: Returns the chirp
- **400 Bad Request**: Invalid chirp ID format
- **404 Not Found**: Chirp not found
- **500 Internal Server Error**: Failed to retrieve chirp

**Example:**
```bash
curl -X GET http://localhost:8080/api/chirps/550e8400-e29b-41d4-a716-446655440000
```

#### POST /api/chirps
Create a new chirp (requires authentication).

**Headers:**
```
Authorization: Bearer <jwt-token>
```

**Request Body:**
```json
{
  "body": "This is my chirp message!"
}
```

**Constraints:**
- Body must be 140 characters or less
- Banned words will be replaced with "****"

**Response:**
- **201 Created**: Chirp created successfully
- **400 Bad Request**: Invalid request payload or body too long
- **401 Unauthorized**: Invalid or missing token
- **500 Internal Server Error**: Failed to create chirp

**Example:**
```bash
curl -X POST http://localhost:8080/api/chirps \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-jwt-token>" \
  -d '{
    "body": "Hello, world! This is my first chirp."
  }'
```

#### DELETE /api/chirps/{chirpID}
Delete a chirp (requires authentication and ownership).

**Headers:**
```
Authorization: Bearer <jwt-token>
```

**Path Parameters:**
- `chirpID`: UUID of the chirp to delete

**Response:**
- **204 No Content**: Chirp deleted successfully
- **400 Bad Request**: Invalid chirp ID format
- **401 Unauthorized**: Invalid or missing token
- **403 Forbidden**: User doesn't own the chirp
- **404 Not Found**: Chirp not found
- **500 Internal Server Error**: Failed to delete chirp

**Example:**
```bash
curl -X DELETE http://localhost:8080/api/chirps/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer <your-jwt-token>"
```

---

### Webhooks

#### POST /api/polka/webhooks
Handle user upgrade events from Polka payment processor.

**Headers:**
```
Authorization: ApiKey <polka-api-key>
```

**Request Body:**
```json
{
  "event": "user.upgraded",
  "data": {
    "user_id": "uuid"
  }
}
```

**Response:**
- **204 No Content**: Event processed successfully
- **400 Bad Request**: Invalid request payload
- **401 Unauthorized**: Missing API key
- **403 Forbidden**: Invalid API key
- **404 Not Found**: User not found
- **500 Internal Server Error**: Failed to upgrade user

---

## Error Responses

All error responses follow this format:
```json
{
  "error": "Error description"
}
```

## Common HTTP Status Codes

- **200 OK**: Request successful
- **201 Created**: Resource created successfully
- **204 No Content**: Request successful, no content to return
- **400 Bad Request**: Invalid request data
- **401 Unauthorized**: Authentication required or invalid
- **403 Forbidden**: Access denied
- **404 Not Found**: Resource not found
- **500 Internal Server Error**: Server error

## Content Word Filtering

Chirps are automatically filtered for inappropriate content. The following words are replaced with "****":
- kerfuffle
- sharbert
- fornax

## Rate Limiting

Currently, there are no rate limits implemented, but this may change in future versions.

## Development Setup

1. Set up your environment variables in `.env`:
   ```
   DB_URL="postgres://postgres:postgres@localhost:5432/chirpy?sslmode=disable"
   PLATFORM="dev"
   JWT_SECRET="your-jwt-secret"
   POLKA_KEY="your-polka-api-key"
   ```

2. Run database migrations
3. Start the server: `go run .`
4. The API will be available at `http://localhost:8080`

## Testing

You can test the API endpoints using the provided `test.http` file with tools like REST Client for VS Code, or use curl commands as shown in the examples above.