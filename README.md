# Token Authentication Middleware for Traefik

A Traefik middleware plugin that provides token-based authentication with secure session management.

## Features

- Token-based authentication via query parameters
- Secure session cookies with SHA-256 hashed tokens
- Token parameter takes priority over stored cookies
- Automatic URL cleanup (removes token from URL after authentication)
- HttpOnly, Secure, and SameSite cookie protection

## Installation

### Static Configuration

Add the plugin to your Traefik static configuration:

```yaml
experimental:
  plugins:
    tokenauth:
      moduleName: github.com/flohoss/tokenauth
      version: v0.1.0
```

Or with Docker labels:

```yaml
labels:
  - 'traefik.experimental.plugins.tokenauth.modulename=github.com/flohoss/tokenauth'
  - 'traefik.experimental.plugins.tokenauth.version=v0.1.0'
```

### Dynamic Configuration

Configure the middleware in your dynamic configuration:

```yaml
# config.yml
http:
  middlewares:
    my-tokenauth:
      plugin:
        tokenauth:
          tokenParam: 'token'
          cookie:
            name: 'auth_session'
            httpOnly: true
            secure: true
            sameSite: 'Strict'
            maxAge: 0
          allowedTokens:
            - 'your-secret-token-1'
            - 'your-secret-token-2'
```

Or with Docker labels:

```yaml
labels:
  - 'traefik.http.middlewares.tokenauth.plugin.tokenauth.tokenParam=token'
  - 'traefik.http.middlewares.tokenauth.plugin.tokenauth.cookie.name=auth_session'
  - 'traefik.http.middlewares.tokenauth.plugin.tokenauth.cookie.httpOnly=true'
  - 'traefik.http.middlewares.tokenauth.plugin.tokenauth.cookie.secure=true'
  - 'traefik.http.middlewares.tokenauth.plugin.tokenauth.cookie.sameSite=Strict'
  - 'traefik.http.middlewares.tokenauth.plugin.tokenauth.cookie.maxAge=0'
  - 'traefik.http.middlewares.tokenauth.plugin.tokenauth.allowedTokens[0]=your-secret-token'
```

### Apply to Routes

```yaml
http:
  routers:
    my-router:
      rule: 'Host(`example.com`)'
      service: my-service
      middlewares:
        - my-tokenauth
```

Or with Docker:

```yaml
labels:
  - 'traefik.http.routers.my-router.rule=Host(`example.com`)'
  - 'traefik.http.routers.my-router.middlewares=tokenauth'
```

## Configuration Options

| Parameter              | Type     | Default          | Description                                         |
| ---------------------- | -------- | ---------------- | --------------------------------------------------- |
| `tokenParam`           | string   | `"token"`        | Query parameter name for the authentication token   |
| `cookie.name`          | string   | `"auth_session"` | Name of the session cookie                          |
| `cookie.httpOnly`      | bool     | `true`           | Set HttpOnly flag on cookies                        |
| `cookie.secure`        | bool     | `true`           | Set Secure flag on cookies (requires HTTPS)         |
| `cookie.sameSite`      | string   | `"Strict"`       | SameSite attribute: "Strict", "Lax", or "None"     |
| `cookie.maxAge`        | int      | `0`              | Cookie max age in seconds (0 = session cookie)      |
| `allowedTokens`        | []string | `[]`             | List of valid authentication tokens                 |

## Usage

### First-Time Authentication

1. User visits: `https://example.com?token=your-secret-token-1`
2. Middleware validates the token
3. Token is hashed (SHA-256) and stored in a secure cookie
4. User is redirected to: `https://example.com` (clean URL)

### Subsequent Requests

- The middleware checks for the session cookie
- If valid, access is granted immediately
- No token in URL is needed for authenticated sessions

### Session Cookie Details

- **Duration**: Configurable via `cookieMaxAge` (default: session cookie that expires when browser closes)
- **Security**: Configurable via `cookieHttpOnly`, `cookieSecure`, `cookieSameSite` (defaults: HttpOnly=true, Secure=true, SameSite=Strict)
- **Storage**: SHA-256 hash of the token (not plaintext)

**Example: Persistent cookie for 30 days**
```yaml
middlewares:
  my-tokenauth:
    plugin:
      tokenauth:
        cookie:
          maxAge: 2592000  # 30 days in seconds
        # ... other config
```

## Authentication Flow

The middleware implements the following authentication priority:

1. **Token Parameter (Highest Priority)**: If a token is provided in the query parameter:
   - Validates the token against the allowed tokens list
   - On success: Sets the session cookie with hashed token and redirects (clean URL)
   - On failure: Returns HTTP 401 Unauthorized

2. **Session Cookie**: If no token is in the query:
   - Checks for a valid session cookie
   - On success: Grants access to the protected resource
   - On failure: Returns HTTP 401 Unauthorized

This ensures that providing a token always validates it, even if a previous valid cookie exists.

## Security Considerations

1. **Use HTTPS**: The middleware sets the `Secure` flag on cookies, requiring HTTPS
2. **Strong Tokens**: Use cryptographically random tokens (e.g., 32+ characters)
3. **Token Storage**: Never expose `allowedTokens` in public repositories
4. **Environment Variables**: Consider loading tokens from environment variables
5. **Token Rotation**: Regularly rotate authentication tokens
