# Authentication specification

## Type

Use **server-side sessions**. Do **not** use JWT for the web interface.

## Session cookie

Production-oriented shape:

```go
http.Cookie{
    Name:     "__Host-forgeboard_session",
    Value:    rawToken,
    Path:     "/",
    HttpOnly: true,
    Secure:   true, // production
    SameSite: http.SameSiteLaxMode,
    MaxAge:   60 * 60 * 24 * 7,
}
```

`__Host-` prefix requires: `Secure`, `Path=/`, no `Domain` attribute.

For local HTTP development, use a simpler name:

```text
forgeboard_session
```

## Session database fields

```text
id
user_id
token_hash
created_at
last_seen_at
expires_at
user_agent
ip_address
revoked_at
```

Never store the raw session token. Store `sha256(rawToken)`.

## Password hashing

Use **Argon2id**.

The application should:

1. generate a random salt
2. hash the password using Argon2id
3. store algorithm parameters with the hash
4. compare in constant time
5. permit future rehashing when parameters change

Do not implement Argon2 itself — use Go’s cryptographic library and wrap it in a password service.

## Login messaging

Login errors must not reveal whether an account exists:

```text
Invalid email or password.
```

## Related

- Login flow: [../examples/flows/login.md](../examples/flows/login.md)
- Registration flow: [../examples/flows/registration.md](../examples/flows/registration.md)
- Middleware: [../architecture/middleware.md](../architecture/middleware.md)
