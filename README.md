# Daydev Utils

Small, focused Go module with three leaf packages:

- crypto — secure token utilities
- jwt — thin wrapper for signing JWTs
- date — time helpers

Module path: `github.com/Daydev-Org/utils` (targets `go 1.24.3`).

## Install

```bash
go get github.com/Daydev-Org/utils@latest
```

Import subpackages as needed:

- `github.com/Daydev-Org/utils/crypto`
- `github.com/Daydev-Org/utils/jwt`
- `github.com/Daydev-Org/utils/date`

## Quick start

### crypto

```go
package main

import (
    "fmt"
    "github.com/Daydev-Org/utils/crypto"
)

func main() {
    tok, err := crypto.GenerateRefreshToken()
    if err != nil { panic(err) }
    fmt.Println("refresh token:", tok) // 86 chars, base64url (no padding)

    hash := crypto.HashToken(tok)
    fmt.Println("sha256(token):", hash)
}
```

### jwt

```go
package main

import (
    stdjwt "github.com/golang-jwt/jwt/v5"
    utiljwt "github.com/Daydev-Org/utils/jwt"
)

func main() {
    secret := []byte("your-hs256-secret")
    claims := stdjwt.MapClaims{
        "sub": "user-123",
        "role": "tester",
    }
    token, err := utiljwt.GenerateToken(secret, claims)
    if err != nil { panic(err) }
    _ = token
}
```

Notes:
- Always signs with HS256. If you need RS/ES algorithms, add a separate function in your codebase.
- Validation (exp/nbf/aud/iss) is not performed by this helper; validate on parse in your app.

### date

```go
package main

import (
    "fmt"
    "time"
    "github.com/Daydev-Org/utils/date"
)

func main() {
    fmt.Println("now (unix):", date.NowUnix())
    fmt.Println("in 5 minutes:", date.AddTime(5*time.Minute))
}
```

## Development

- Toolchain: Go 1.24.x (module targets `1.24.3`).
- Dependencies: only `github.com/golang-jwt/jwt/v5` (used by the `jwt` package).

Common commands:

```bash
go mod tidy
go vet ./...
go build ./...
go test ./...
```

With race detector and coverage:

```bash
go test -race -covermode=atomic -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## Versioning

Semantic versioning. This is a leaf utility module; avoid breaking API changes without a major version bump.

## License

Copyright (c) 2022–2025 Daydev, Inc. All Rights Reserved.