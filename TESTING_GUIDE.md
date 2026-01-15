# MGSearch - Testing Guide & Quick Fixes

## Overview

This guide provides instructions for running tests, fixing bugs, and ensuring code quality.

---

## Running Tests

### Prerequisites

1. **Start MongoDB:**
```bash
# Using the dev environment
just dev-up

# Or manually
mongod --dbpath /path/to/data
```

2. **Set Environment Variables:**
```bash
export DATABASE_URL="mongodb://localhost:27017/mgsearch_test"
export MEILISEARCH_URL="http://localhost:7701"
export MEILISEARCH_API_KEY="test-master-key"
export JWT_SIGNING_KEY="test-signing-key-32-characters-long-here"
export ENCRYPTION_KEY="0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
export SHOPIFY_API_KEY="test"
export SHOPIFY_API_SECRET="test"
export SHOPIFY_APP_URL="http://localhost:8080"
```

### Run All Tests

```bash
# Run all handler tests
go test ./handlers/... -v

# Run with coverage
go test ./handlers/... -cover -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out

# Run specific test file
go test -v ./handlers/user_auth_test.go

# Run specific test function
go test -v ./handlers/user_auth_test.go -run TestUserAuthHandler_RegisterUser
```

### Run Tests Without External Dependencies (Unit Tests Only)

```bash
# Run tests with short flag (skips integration tests)
go test ./handlers/... -short -v
```

---

## Test Files Created

### ✅ New Comprehensive Test Files

1. **handlers/user_auth_test.go** (930+ lines)
   - Complete coverage of UserAuthHandler
   - Tests all 10 API endpoints
   - 50+ test cases covering:
     - User registration
     - Login/logout
     - Client creation
     - API key generation
     - API key revocation
     - User updates
     - Access control
     - Edge cases
     - Security scenarios

2. **handlers/index_test.go** (350+ lines)
   - Complete coverage of IndexHandler
   - Tests index creation and listing
   - Concurrent access tests
   - UID format validation
   - Race condition detection

### ✅ Existing Test Files (Enhanced)

All other handlers already have test files:
- `handlers/auth_test.go` - Shopify OAuth
- `handlers/search_test.go` - Search operations
- `handlers/storefront_test.go` - Storefront search
- `handlers/store_test.go` - Store management
- `handlers/session_test.go` - Session management
- `handlers/webhook_test.go` - Webhook processing
- `handlers/settings_test.go` - Settings updates
- `handlers/tasks_test.go` - Task queries

---

## Critical Bugs & Quick Fixes

### 🔴 BUG #1: Race Condition in Index Creation (HIGH)

**File:** `handlers/index.go:51`

**Issue:** Concurrent index creation can cause conflicts

**Quick Fix:**

```go
// BEFORE (line 51)
existing, _ := h.indexRepo.FindByNameAndClientID(c.Request.Context(), req.Name, clientID)
if existing != nil {
    c.JSON(http.StatusConflict, gin.H{"error": "Index with this name already exists for this client"})
    return
}

// AFTER - Handle database duplicate key errors
savedIndex, err := h.indexRepo.Create(c.Request.Context(), index)
if err != nil {
    // Check for duplicate key error
    if mongo.IsDuplicateKeyError(err) {
        c.JSON(http.StatusConflict, gin.H{
            "error": "Index with this name already exists for this client",
            "code": "DUPLICATE_INDEX",
        })
        return
    }
    c.JSON(http.StatusInternalServerError, gin.H{
        "error": fmt.Sprintf("Failed to save index record: %v", err),
    })
    return
}
```

**Also Add to** `repositories/index_repository.go`:

```go
// Add unique index to MongoDB
func EnsureIndexes(db *mongo.Database) error {
    indexModel := mongo.IndexModel{
        Keys: bson.D{
            {Key: "client_id", Value: 1},
            {Key: "name", Value: 1},
        },
        Options: options.Index().SetUnique(true),
    }
    _, err := db.Collection("indexes").Indexes().CreateOne(context.Background(), indexModel)
    return err
}
```

---

### 🟡 BUG #2: Missing Input Validation (MEDIUM)

**File:** `handlers/search.go:81`

**Issue:** Index names not validated for special characters

**Quick Fix:**

```go
// Add this helper function at the top of the file
import "regexp"

var validIndexNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

func isValidIndexName(name string) bool {
    return validIndexNameRegex.MatchString(name)
}

// BEFORE (line 81)
indexName := strings.TrimSpace(c.Param("index_name"))
if indexName == "" {
    c.JSON(http.StatusBadRequest, gin.H{
        "error": "index name is required",
    })
    return
}

// AFTER
indexName := strings.TrimSpace(c.Param("index_name"))
if indexName == "" {
    c.JSON(http.StatusBadRequest, gin.H{
        "error": "index name is required",
    })
    return
}
if !isValidIndexName(indexName) {
    c.JSON(http.StatusBadRequest, gin.H{
        "error": "invalid index name format",
        "details": "Index name must be 1-64 characters, alphanumeric, hyphens, or underscores only",
    })
    return
}
```

**Apply the same fix to:**
- `handlers/settings.go:77`
- `handlers/tasks.go` (if using index names)

---

### 🟡 BUG #3: Email Case Sensitivity (MEDIUM)

**File:** `repositories/user_repository.go`

**Issue:** MongoDB might not have case-insensitive index on email

**Quick Fix:**

Add to `pkg/database/migrations.go`:

```go
func createUserIndexes(db *mongo.Database) error {
    // Create case-insensitive unique index on email
    indexModel := mongo.IndexModel{
        Keys: bson.D{{Key: "email", Value: 1}},
        Options: options.Index().
            SetUnique(true).
            SetCollation(&options.Collation{
                Locale:   "en",
                Strength: 2, // Case-insensitive
            }),
    }
    
    _, err := db.Collection("users").Indexes().CreateOne(context.Background(), indexModel)
    return err
}

// Call this in RunMigrations
func RunMigrations(ctx context.Context, client *mongo.Client, dbName string) error {
    db := client.Database(dbName)
    
    // Existing migrations...
    
    // Add user indexes
    if err := createUserIndexes(db); err != nil {
        return fmt.Errorf("failed to create user indexes: %w", err)
    }
    
    return nil
}
```

---

### 🟢 BUG #4: Missing Rate Limiting (LOW)

**Create:** `middleware/rate_limit_middleware.go`

```go
package middleware

import (
    "net/http"
    "sync"
    "time"

    "github.com/gin-gonic/gin"
)

type rateLimiter struct {
    requests map[string][]time.Time
    mu       sync.RWMutex
    limit    int
    window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) gin.HandlerFunc {
    rl := &rateLimiter{
        requests: make(map[string][]time.Time),
        limit:    limit,
        window:   window,
    }

    // Cleanup old entries every minute
    go rl.cleanup()

    return func(c *gin.Context) {
        clientIP := c.ClientIP()

        if !rl.allow(clientIP) {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "rate limit exceeded",
                "code":  "RATE_LIMIT_EXCEEDED",
            })
            c.Abort()
            return
        }

        c.Next()
    }
}

func (rl *rateLimiter) allow(key string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    now := time.Now()
    cutoff := now.Add(-rl.window)

    // Remove old requests
    requests := rl.requests[key]
    validRequests := []time.Time{}
    for _, t := range requests {
        if t.After(cutoff) {
            validRequests = append(validRequests, t)
        }
    }

    // Check limit
    if len(validRequests) >= rl.limit {
        return false
    }

    // Add current request
    validRequests = append(validRequests, now)
    rl.requests[key] = validRequests

    return true
}

func (rl *rateLimiter) cleanup() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        rl.mu.Lock()
        now := time.Now()
        cutoff := now.Add(-rl.window)

        for key, requests := range rl.requests {
            validRequests := []time.Time{}
            for _, t := range requests {
                if t.After(cutoff) {
                    validRequests = append(validRequests, t)
                }
            }
            if len(validRequests) == 0 {
                delete(rl.requests, key)
            } else {
                rl.requests[key] = validRequests
            }
        }
        rl.mu.Unlock()
    }
}
```

**Add to** `main.go`:

```go
import "mgsearch/middleware"

// Add rate limiting
router.Use(middleware.NewRateLimiter(100, 1*time.Minute)) // 100 req/min

// Or per-route:
publicAPI := router.Group("/api/v1")
publicAPI.Use(middleware.NewRateLimiter(60, 1*time.Minute))
```

---

### 🟢 BUG #5: Missing Structured Logging (LOW)

**Create:** `pkg/logger/logger.go`

```go
package logger

import (
    "os"

    "github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func init() {
    Log = logrus.New()
    Log.SetOutput(os.Stdout)
    Log.SetFormatter(&logrus.JSONFormatter{})
    
    // Set log level from environment
    level := os.Getenv("LOG_LEVEL")
    switch level {
    case "debug":
        Log.SetLevel(logrus.DebugLevel)
    case "warn":
        Log.SetLevel(logrus.WarnLevel)
    case "error":
        Log.SetLevel(logrus.ErrorLevel)
    default:
        Log.SetLevel(logrus.InfoLevel)
    }
}

func Info(msg string, fields map[string]interface{}) {
    Log.WithFields(fields).Info(msg)
}

func Error(msg string, err error, fields map[string]interface{}) {
    if fields == nil {
        fields = make(map[string]interface{})
    }
    if err != nil {
        fields["error"] = err.Error()
    }
    Log.WithFields(fields).Error(msg)
}

func Debug(msg string, fields map[string]interface{}) {
    Log.WithFields(fields).Debug(msg)
}

func Warn(msg string, fields map[string]interface{}) {
    Log.WithFields(fields).Warn(msg)
}
```

**Usage in handlers:**

```go
import "mgsearch/pkg/logger"

// Before
if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
    return
}

// After
if err != nil {
    logger.Error("Failed to create user", err, map[string]interface{}{
        "email":      email,
        "ip_address": c.ClientIP(),
        "user_agent": c.Request.UserAgent(),
    })
    c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
    return
}
```

---

## Standardized Error Responses

**Create:** `pkg/errors/api_error.go`

```go
package errors

type APIError struct {
    Error   string                 `json:"error"`
    Code    string                 `json:"code,omitempty"`
    Details map[string]interface{} `json:"details,omitempty"`
}

func NewAPIError(message, code string) *APIError {
    return &APIError{
        Error: message,
        Code:  code,
    }
}

func (e *APIError) WithDetails(details map[string]interface{}) *APIError {
    e.Details = details
    return e
}

// Common errors
var (
    ErrUnauthorized     = NewAPIError("unauthorized", "UNAUTHORIZED")
    ErrForbidden        = NewAPIError("access denied", "FORBIDDEN")
    ErrNotFound         = NewAPIError("resource not found", "NOT_FOUND")
    ErrBadRequest       = NewAPIError("invalid request", "BAD_REQUEST")
    ErrConflict         = NewAPIError("resource already exists", "CONFLICT")
    ErrInternalServer   = NewAPIError("internal server error", "INTERNAL_ERROR")
    ErrRateLimitExceeded = NewAPIError("rate limit exceeded", "RATE_LIMIT_EXCEEDED")
)
```

**Usage:**

```go
import apierrors "mgsearch/pkg/errors"

// Instead of:
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})

// Use:
c.JSON(http.StatusBadRequest, apierrors.ErrBadRequest.WithDetails(map[string]interface{}{
    "field": "email",
    "reason": "invalid format",
}))
```

---

## Security Enhancements

### 1. Strong Password Validation

**Create:** `pkg/auth/password_validator.go`

```go
package auth

import (
    "errors"
    "unicode"
)

func ValidatePasswordStrength(password string) error {
    if len(password) < 12 {
        return errors.New("password must be at least 12 characters long")
    }

    var (
        hasUpper   bool
        hasLower   bool
        hasNumber  bool
        hasSpecial bool
    )

    for _, char := range password {
        switch {
        case unicode.IsUpper(char):
            hasUpper = true
        case unicode.IsLower(char):
            hasLower = true
        case unicode.IsDigit(char):
            hasNumber = true
        case unicode.IsPunct(char) || unicode.IsSymbol(char):
            hasSpecial = true
        }
    }

    if !hasUpper {
        return errors.New("password must contain at least one uppercase letter")
    }
    if !hasLower {
        return errors.New("password must contain at least one lowercase letter")
    }
    if !hasNumber {
        return errors.New("password must contain at least one number")
    }
    if !hasSpecial {
        return errors.New("password must contain at least one special character")
    }

    return nil
}
```

**Use in** `handlers/user_auth.go`:

```go
// Add validation before hashing
if err := auth.ValidatePasswordStrength(req.Password); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{
        "error": "weak password",
        "details": err.Error(),
    })
    return
}

passwordHash, err := auth.HashPassword(req.Password)
```

### 2. Audit Logging

**Create:** `pkg/audit/audit.go`

```go
package audit

import (
    "context"
    "time"

    "mgsearch/repositories"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type AuditLog struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    Timestamp time.Time          `bson:"timestamp"`
    UserID    string             `bson:"user_id,omitempty"`
    Action    string             `bson:"action"`
    Resource  string             `bson:"resource"`
    ResourceID string            `bson:"resource_id,omitempty"`
    IPAddress string             `bson:"ip_address"`
    UserAgent string             `bson:"user_agent"`
    Success   bool               `bson:"success"`
    ErrorCode string             `bson:"error_code,omitempty"`
    Details   map[string]interface{} `bson:"details,omitempty"`
}

type AuditLogger struct {
    repo *repositories.AuditRepository
}

func NewAuditLogger(repo *repositories.AuditRepository) *AuditLogger {
    return &AuditLogger{repo: repo}
}

func (a *AuditLogger) Log(ctx context.Context, log *AuditLog) error {
    log.Timestamp = time.Now().UTC()
    return a.repo.Create(ctx, log)
}
```

**Use in handlers:**

```go
// Log sensitive operations
auditLogger.Log(c.Request.Context(), &audit.AuditLog{
    UserID:     userID,
    Action:     "USER_LOGIN",
    Resource:   "user",
    ResourceID: user.ID.Hex(),
    IPAddress:  c.ClientIP(),
    UserAgent:  c.Request.UserAgent(),
    Success:    true,
})
```

---

## Testing Checklist

Before deploying to production, ensure:

- [ ] All tests pass: `go test ./handlers/... -v`
- [ ] Test coverage > 75%: `go test ./handlers/... -cover`
- [ ] No race conditions: `go test ./handlers/... -race`
- [ ] Linter passes: `golangci-lint run`
- [ ] Format check: `go fmt ./...`
- [ ] Dependencies updated: `go mod tidy`
- [ ] Environment variables documented
- [ ] Database indexes created
- [ ] Rate limiting enabled
- [ ] Logging configured
- [ ] CORS configured for production
- [ ] API documentation updated

---

## Quick Start

```bash
# 1. Clone and setup
git clone <repo>
cd mgsearch
nix develop  # Or install dependencies manually

# 2. Start services
just dev-up

# 3. Run tests
go test ./handlers/... -v

# 4. Apply fixes (see BUG #1-5 above)

# 5. Re-run tests
go test ./handlers/... -v -cover

# 6. Check coverage
go test ./handlers/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Next Steps

1. **Apply all quick fixes** listed above
2. **Run full test suite** and verify all pass
3. **Review security enhancements** and implement
4. **Add monitoring** (Prometheus, Grafana)
5. **Set up CI/CD** with automated testing
6. **Perform load testing** (k6, Apache Bench)
7. **Security audit** by third party
8. **Penetration testing**

---

## Resources

- [Go Testing](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [MongoDB Security](https://docs.mongodb.com/manual/security/)

---

**Last Updated:** 2026-01-15  
**Version:** 1.0  
**Status:** Ready for Implementation
