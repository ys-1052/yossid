# 08. Security設計

## 1. 基本方針

無料・低コスト構成のためDB閉域接続は行わない。

ただし、以下により個人開発MVPとして十分堅い構成を目指す。

- TLS verify-full
- DB runtime user最小権限
- secrets管理
- token/code/session hash保存
- PKCE S256必須
- Authorization Code Flowのみ
- Refresh Token Rotation
- 2FA必須
- 監査ログ
- rate limit
- secure cookie

## 2. OAuth/OIDCセキュリティ

### 許可するフロー

```text
Authorization Code Flowのみ
```

### 禁止するフロー

```text
Implicit Flow
Hybrid Flow
Resource Owner Password Credentials
Client Credentials
```

### 必須パラメータ

`/authorize` では以下を必須とする。

- response_type=code
- client_id
- redirect_uri
- scope includes openid
- state
- nonce
- code_challenge
- code_challenge_method=S256

### redirect_uri

- 完全一致のみ
- wildcard禁止
- query込みで登録値と一致させる
- 不正なredirect_uriにはredirectしない

### PKCE

- S256のみ許可
- plain禁止
- 未指定禁止
- confidential clientでも必須

## 3. 認証セキュリティ

### Password

- Argon2idを推奨
- bcryptでも可
- password policyを設ける
- passwordはログ出力しない
- password hashのみ保存

### Email OTP 2FA

- 必須
- 6桁
- 有効期限5分
- 最大試行回数5回
- hash保存
- 再送制限
- OTP成功後にOP sessionを発行

### Email Verification

- 登録時必須
- tokenはhash保存
- 短命
- 1回限り

## 4. Cookieセキュリティ

OP session cookie:

```text
HttpOnly
Secure
SameSite=Lax
Path=/
```

MVPではDomain属性は設定しない。

CSRF用Cookieを使う場合はセッションCookieと分離する。

## 5. CSRF対策

HTML form endpointにはCSRF tokenを入れる。

対象:

- POST /register
- POST /login
- POST /mfa/email/verify
- POST /logout

`/authorize` の `state` はRP側CSRF対策であり、OP自身のform CSRF対策とは別に扱う。

## 6. Token保存

DBには生値を保存しない。

hash保存対象:

- authorization code
- refresh token
- OP session id
- email verification token
- OTP
- password reset token 将来用

hashにはpepperを加える。

pepperはSSM SecureStringで管理する。

## 7. DB接続セキュリティ

### 接続

- Neon public endpoint
- `sslmode=verify-full`
- pooled connection endpoint

### DBユーザー

分離する。

```text
migration_user
  - DDL権限あり
  - CI/ローカルから利用
  - Lambdaには渡さない

runtime_user
  - DML権限のみ
  - 必要テーブルのみ
  - Lambdaで利用
```

### 権限例

```sql
REVOKE ALL ON SCHEMA public FROM PUBLIC;

GRANT USAGE ON SCHEMA oidc TO app_user;

GRANT SELECT, INSERT, UPDATE, DELETE
ON ALL TABLES IN SCHEMA oidc
TO app_user;
```

## 8. Secrets管理

SSM Parameter Store SecureStringで管理する。

保持対象:

- DATABASE_URL
- JWT private key
- cookie signing key
- token pepper
- OTP pepper

禁止:

- GitHubにcommit
- Lambda環境変数に平文で直接書く
- ログ出力

Lambda起動時にSSMから読み込む。

必要に応じてメモリキャッシュする。

## 9. Rate Limit

MVPでも最低限入れる。

対象:

- POST /register
- POST /login
- POST /mfa/email/verify
- POST /mfa/email/send
- GET /email/verify
- POST /token

Lambda単体で厳格な分散rate limitは難しいため、MVPではDBベースまたは簡易制限とする。

将来候補:

- WAF rate-based rule
- API Gateway usage plan
- DynamoDB / Redisベースrate limit

## 10. 監査ログ

以下を記録する。

- user_registered
- email_verified
- login_password_success
- login_password_failure
- email_otp_sent
- email_otp_success
- email_otp_failure
- authorization_code_issued
- token_issued
- refresh_token_rotated
- refresh_token_reuse_detected
- logout
- user_withdrawn

metadataには機微情報を入れない。

## 11. ログ禁止事項

ログ出力禁止:

- password
- OTP
- authorization code
- access token
- refresh token
- id token
- client secret
- session cookie
- email verification token

## 12. Header

レスポンスに付与する。

```text
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
Referrer-Policy: no-referrer
Cache-Control: no-store
```

Discovery / JWKSのみpublic cache可。

## 13. Error設計

ユーザー列挙を避ける。

ログイン失敗時:

```text
メールアドレスまたはパスワードが正しくありません
```

登録時も、既存emailかどうかを過度に露出しない。

OAuth errorは仕様形式に従う。

## 14. 外部公開前に必要な追加対策

- Consent画面
- Client審査
- Client secret rotation
- WAF
- より強いrate limit
- 独自ドメイン
- security.txt
- privacy policy
- terms
- OpenID Conformance Suite
- pentest
- DB接続のIP制限またはPrivateLink検討
