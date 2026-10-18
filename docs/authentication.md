# Authentication Guide

This guide covers how Orla Numis signs users in — username and password, JWT access tokens and refresh tokens — with code examples and production security recommendations. There is no other sign-in method.

## JWT Authentication (Primary)

JWT is the primary authentication method. Users log in with a username and password to receive a short-lived access token.

### Registration

The first user to register is automatically assigned the **admin** role. All subsequent users are regular users.

After the first user, new accounts are refused while the `RegistrationMode` setting is `closed` (the default); set it to `open` to let others register.

Registration requires a **username**, **password**, and **email** (validated as a proper email format).

```sh
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "password": "s3cur3Pa$$word", "email": "alice@example.com"}'
```

### Login

```sh
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "password": "s3cur3Pa$$word"}'
```

**Response:**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "refreshToken": "rt_a1b2c3d4e5f6...",
  "user": {
    "id": 1,
    "username": "alice",
    "role": "admin",
    "email": "alice@example.com",
    "avatarPath": "/uploads/avatars/alice.jpg",
    "isPublic": false,
    "bio": ""
  }
}
```

The registration endpoint returns the same response shape.

### Using the Access Token

Pass the token in the `Authorization` header on all authenticated requests:

```sh
curl http://localhost:8080/api/coins \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

### Current User

```sh
curl http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Response:**

```json
{
  "id": 1,
  "username": "alice",
  "role": "admin",
  "email": "alice@example.com",
  "avatarPath": "/uploads/avatars/alice.jpg",
  "isPublic": false,
  "bio": "",
  "emailMissing": false,
  "createdAt": "2024-01-15T08:30:00Z"
}
```

#### Email Migration for Legacy Users

Users who registered before email was required will have `emailMissing: true` in the response. The frontend displays a dismissible modal prompting them to add their email address. The prompt can be dismissed for 7 days (tracked via `localStorage`), after which it reappears.

### Token Details

| Property | Value |
| -------- | ----- |
| **Expiry** | 15 minutes |
| **Signing secret** | `JWT_SECRET` environment variable |
| **Header** | `Authorization: Bearer <token>` |
| **Middleware** | `src/api/middleware/auth.go` — validates the JWT and populates `userId` and `userRole` in the Gin context |

---

## Refresh Tokens

Refresh tokens allow the frontend to silently obtain new access tokens without requiring the user to log in again.

### How It Works

1. On login, the server returns both an **access token** (15 min) and a **refresh token** (30 days).
2. When the access token expires, the client sends the refresh token to get a new pair.
3. The old refresh token is **revoked** on each refresh — this is a rolling window. Each refresh token can only be used once.

### Refreshing Tokens

```sh
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refreshToken": "rt_a1b2c3d4e5f6..."}'
```

**Response:**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...(new access token)",
  "refreshToken": "rt_f6e5d4c3b2a1...(new refresh token)"
}
```

### Token Format & Storage

| Property | Value |
| -------- | ----- |
| **Format** | `rt_` prefix + 32 random hex bytes |
| **Server storage** | SHA-256 hash in the database (`RefreshToken` model) |
| **Client storage** | `localStorage` on the frontend |
| **Lifetime** | 30 days (rolling — resets on each refresh) |

### Frontend Behavior

The frontend uses an **axios interceptor** that automatically handles token refresh:

- When a request returns **401 Unauthorized**, the interceptor sends the refresh token to `/api/auth/refresh`.
- While the refresh is in flight, any **concurrent requests** are queued and replayed with the new token once the refresh completes. This avoids race conditions when multiple API calls fail at the same time.
- If the refresh itself fails (e.g., the refresh token has expired), the user is redirected to the login page.

---

## How Requests Are Checked

Protected routes require the `Authorization: Bearer <access token>` header (`src/api/middleware/auth.go`). The middleware checks the signature and expiry and puts `userId` and `userRole` in the Gin context. There is no API-key or query-parameter alternative: photos and other uploads are fetched by the web app with the same header.

## Sign-in Protection

- **Lockout** — repeated failed sign-ins lock the account for a while. Administrators see locked users in **Administration → Security** and can unlock them.
- **IP blocking** — repeated password failures from one IP block that IP for a while, automatically.
- **Rate limits** — sign-in and registration are rate limited; a `429` response carries `Retry-After`, and the login page counts down.
- **Generic errors** — a failed sign-in never says whether the username exists.

### JWT Claim Fields

The access token contains these claims (set in `src/api/services/auth_service.go`):

| Claim | Type | Description |
|-------|------|-------------|
| `userId` | number | Database user ID |
| `username` | string | Login username |
| `role` | string | `admin` or `user` |
| `exp` | number | Expiry (Unix timestamp, 15 minutes from issue) |
| `iat` | number | Issued-at (Unix timestamp) |

---

## Production Security Checklist

Before deploying to production, review these settings:

### JWT Secret

Change the `JWT_SECRET` environment variable to a strong, random value (minimum 32 characters):

```sh
export JWT_SECRET=$(openssl rand -base64 48)
```

### Registration

Keep `RegistrationMode` at `closed` unless you are inviting someone, and close it again afterwards.

### Refresh Token Storage

Refresh tokens are currently stored in `localStorage` on the frontend. This is acceptable for a PWA but be aware that `localStorage` is accessible to any JavaScript running on the same origin.

### Summary

| Setting | Environment Variable | Example Value |
| ------- | -------------------- | ------------- |
| JWT signing secret | `JWT_SECRET` | Output of `openssl rand -base64 48` |
| Allowed browser origins | `CORS_ORIGINS` | `https://coins.example.com` |
| Reverse proxy addresses | `TRUSTED_PROXIES` | `10.0.0.0/8` |
