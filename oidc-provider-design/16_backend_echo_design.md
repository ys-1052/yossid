# 16. Backend Echo設計

## 1. 方針

Go backendは Echo を利用する。

OIDC/OAuthのプロトコル処理は Ory Fosite を使う。

AWS Lambda上では Echo を Lambda Adapter 経由で実行する。

## 2. 構成

```text
API Gateway HTTP API
  -> Lambda
    -> Echo Adapter
      -> Echo
        -> Middleware
        -> Handlers
        -> Services
        -> Repositories
        -> Neon Postgres
```

## 3. 採用ライブラリ候補

```text
github.com/labstack/echo/v4
github.com/awslabs/aws-lambda-go-api-proxy/echo
github.com/aws/aws-lambda-go/lambda
github.com/ory/fosite
github.com/jackc/pgx/v5
```

DBアクセスは以下のどちらかを推奨する。

```text
sqlc + pgx
```

または

```text
pgx直書きRepository
```

MVPでは型安全性・保守性のため `sqlc + pgx` を推奨する。

## 4. Entry Point

```text
cmd/lambda/main.go
```

local実行もできるようにする。

```text
cmd/server/main.go
```

または環境変数で切り替える。

```text
RUN_MODE=lambda
RUN_MODE=http
```

## 5. Echo Router

```text
GET  /.well-known/openid-configuration
GET  /jwks.json

GET  /authorize
POST /token
GET  /userinfo

GET  /register
POST /register
GET  /email/verify

GET  /login
POST /login
GET  /mfa/email
POST /mfa/email/verify

POST /logout
GET  /healthz
```

Next.js frontendを使う場合、Go側はAPI endpointのみ提供する構成に寄せてもよい。

その場合:

```text
POST /api/register
POST /api/login
POST /api/mfa/email/verify
POST /api/logout
```

ただし、OIDC endpointはGo側が必ず担当する。

## 6. Middleware

MVPで実装するMiddleware:

- Request ID
- Recover
- Secure Headers
- No-store for auth endpoints
- Access log
- Error log
- CSRF for form endpoints
- OP session loader
- Rate limit optional

注意:

- access logにtoken/code/password/OTPを出さない
- query string全体をログに出さない
- Authorization headerをログに出さない
- Cookie headerをログに出さない

## 7. Error Handler

Echo default error handlerをそのまま使わず、用途別に分ける。

### HTML endpoint

- ユーザー向けエラー画面
- 詳細を出しすぎない

### OAuth endpoint

OAuth 2.0 error responseに合わせる。

```json
{
  "error": "invalid_request",
  "error_description": "..."
}
```

### OIDC redirect error

redirect_uri検証済みの場合のみredirect_uriへerrorを返す。

redirect_uri不正時はredirectしない。

## 8. Handler設計

```text
internal/http/handler/
  authorize.go
  token.go
  userinfo.go
  discovery.go
  jwks.go
  register.go
  login.go
  mfa.go
  logout.go
  health.go
```

Handlerは薄くする。

責務:

- request parse
- validation呼び出し
- service呼び出し
- response生成

## 9. Service設計

```text
internal/service/
  registration_service.go
  login_service.go
  mfa_service.go
  oidc_service.go
  token_service.go
  userinfo_service.go
  key_service.go
  audit_service.go
```

## 10. Repository設計

```text
internal/repository/
  user_repository.go
  client_repository.go
  auth_code_repository.go
  refresh_token_repository.go
  session_repository.go
  registration_repository.go
  otp_repository.go
  signing_key_repository.go
  audit_repository.go
```

RepositoryはDBトランザクションを明示的に扱う。

特に以下はトランザクション必須。

- メール確認完了 -> users本登録
- authorization code使用
- refresh token rotation
- refresh token reuse detection
- withdrawal

## 11. Fosite連携

Echo handlerでHTTP request/responseをFositeに渡す。

方針:

- Fosite provider生成は起動時
- StorageはPostgres実装
- Client storageはclients table
- Authorize request生成時にOP sessionと紐付ける
- Claim mappingは独自Serviceで行う

## 12. Local Development

localでは通常のHTTP serverとして起動できるようにする。

```bash
go run ./cmd/server
```

Lambda向けbuild:

```bash
GOOS=linux GOARCH=arm64 go build -o bootstrap ./cmd/lambda
```

## 13. Lambda Adapter

Lambda entrypoint例:

```go
package main

import (
    "github.com/aws/aws-lambda-go/lambda"
    echoadapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
)

func main() {
    e := NewEchoApp()
    adapter := echoadapter.New(e)
    lambda.Start(adapter.ProxyWithContext)
}
```

実装時はpackage構成に合わせて調整する。

## 14. HTML Rendering

Next.js frontendを採用するため、Go側でHTML renderingは極力持たない。

ただし、以下はGo側で直接返してもよい。

- OAuth error minimal page
- healthz
- fallback error page

UIはNext.js側に寄せる。

## 15. Security Headers

全レスポンスに基本headerを付与する。

```text
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
Referrer-Policy: no-referrer
```

認証系endpoint:

```text
Cache-Control: no-store
```

Discovery / JWKS:

```text
Cache-Control: public, max-age=3600
```

## 16. Request Logging

ログ出力項目:

- request_id
- method
- path
- status
- latency
- user_id optional
- client_id optional
- event_type optional

出力しない:

- query string raw
- authorization header
- cookie
- password
- OTP
- token
- code
- client_secret

## 17. Recommended Directory

```text
backend/
  cmd/
    lambda/
      main.go
    server/
      main.go

  internal/
    app/
      echo.go
      routes.go
      middleware.go

    handler/
    service/
    repository/
    oidc/
    authn/
    security/
    mail/
    audit/
    config/

  migrations/
  sql/
    queries/
    schema/
```

## 18. MVP実装順

1. Echo skeleton
2. Config / SSM loading
3. DB connection
4. healthz
5. registration API
6. email verification
7. login
8. email OTP
9. Fosite authorize
10. token
11. userinfo
12. refresh token rotation
13. logout
14. audit log


## 19. Version Pinning

Backendは以下を基準バージョンとして開始する。

```text
Go 1.26
github.com/labstack/echo/v4 v4.15.4
github.com/ory/fosite v0.49.0
github.com/aws/aws-lambda-go v1.54.0
github.com/awslabs/aws-lambda-go-api-proxy v0.16.2
github.com/jackc/pgx/v5 v5.10.0
github.com/aws/aws-sdk-go-v2/config v1.32.25
github.com/aws/aws-sdk-go-v2/service/ssm v1.69.3
github.com/aws/aws-sdk-go-v2/service/sesv2 v1.62.4
golang.org/x/crypto v0.53.0
```

`go.mod` と `go.sum` は必ずcommitする。

Lambda runtimeは `provided.al2023` を使う。
