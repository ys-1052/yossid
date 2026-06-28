# 12. Implementation Plan

## Phase 0. リポジトリ準備

- Go module作成
- Echo導入
- CDK TypeScript project作成
- migration tool導入
- local Postgres docker compose作成
- lint / test / build設定
- GitHub Actions任意

## Phase 1. DB基盤

- migrations作成
- users
- user_profiles
- user_password_credentials
- pending_user_registrations
- clients
- client_redirect_uris
- client_allowed_scopes
- authorization_requests
- authorization_codes
- refresh_tokens
- login_sessions
- email_otp_challenges
- signing_keys
- audit_logs
- user_withdrawals

## Phase 2. AWS基盤

- CDK stack作成
- Lambda
- API Gateway
- IAM
- CloudWatch Logs
- SSM Parameter参照
- SES権限

## Phase 3. Config / Secrets

- SSM読み込み実装
- DATABASE_URL読み込み
- JWT private key読み込み
- pepper読み込み
- cookie key読み込み

## Phase 4. User Registration

- GET /register
- POST /register
- pending_user_registrations保存
- email verification token発行
- SES送信
- GET /email/verify
- users本登録

## Phase 5. Login / MFA

- GET /login
- POST /login
- password検証
- email OTP生成
- SES送信
- GET /mfa/email
- POST /mfa/email/verify
- OP session cookie発行

## Phase 6. Fosite / OIDC Core

- Fosite provider初期化
- client storage実装
- authorize handler
- token handler
- PKCE
- authorization code保存
- ID Token発行
- Access Token発行
- Refresh Token発行

## Phase 7. Refresh Token Rotation

- refresh token grant
- rotation
- reuse detection
- token family revoke
- audit log

## Phase 8. Discovery / JWKS / UserInfo

- /.well-known/openid-configuration
- /jwks.json
- /userinfo
- scope-based claims mapping

## Phase 9. Logout / Withdrawal

- POST /logout
- OP session revoke
- user withdrawal function
- refresh token revoke
- audit log

## Phase 10. Test

- unit tests
- integration tests
- security tests
- test RP連携

## Phase 11. Deploy

- TerraformでNeon project作成
- Neon role/database作成
- migration実行
- SSM Parameter投入
- CDK deploy
- SES検証
- Client seed投入
- end-to-end test

## 推奨実装順序

最初の動くゴール:

```text
/register
/email/verify
/login
/mfa/email/verify
/authorize
/token
/userinfo
```

その後:

```text
refresh token rotation
logout
audit logs
rate limit
conformance test準備
```

## Phase 6.5. Next.js Frontend

- frontend workspace作成
- Next.js App Router導入
- Tailwind CSS導入
- Auth Shell作成
- UI components作成
- Register画面
- Register Sent画面
- Email Verify画面
- Login画面
- Email OTP画面
- Logout画面
- Error画面
- Go API連携
- CSRF連携
- CloudFront path routing設計反映

## Phase 2.5. Echo Backend Skeleton

- Echo導入
- Lambda Adapter導入
- local server mode実装
- Lambda handler mode実装
- routing定義
- middleware定義
- error handler定義
- request id
- security headers
- no-store headers
- form parsing
- JSON response helper
- OAuth error response helper
