# 06. Data Model設計

## 1. 方針

- Postgresを利用する
- Neon Postgresに保存する
- DBスキーマはmigration toolで管理する
- runtime userとmigration userを分離する
- token / code / session id / OTP はhash保存
- usersにはメール確認済みユーザーのみ保存する
- メール確認前は pending_user_registrations で管理する
- statusでアカウント利用可否を管理する

## 2. ER概要

```text
users
  ├─ user_profiles
  ├─ user_password_credentials
  ├─ login_sessions
  ├─ refresh_tokens
  ├─ authorization_codes
  ├─ user_withdrawals
  └─ audit_logs

clients
  ├─ client_redirect_uris
  ├─ client_allowed_scopes
  ├─ authorization_codes
  └─ refresh_tokens

pending_user_registrations
email_otp_challenges
signing_keys
```

## 3. Enum

### user_status

```text
active
withdrawn
disabled
locked
```

MVPで使う:

```text
active
withdrawn
```

### gender

```text
male
female
other
```

### client_type

```text
confidential
public
```

MVPでは `confidential` のみ。

### token_endpoint_auth_method

```text
client_secret_basic
client_secret_post
private_key_jwt
none
```

MVPでは `client_secret_basic` のみ。

## 4. users

```sql
CREATE TABLE users (
  id UUID PRIMARY KEY,
  sub TEXT UNIQUE NOT NULL,
  email TEXT UNIQUE NOT NULL,
  email_verified_at TIMESTAMPTZ NOT NULL,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### 備考

- `sub` はOIDCの安定識別子
- emailはログインIDとして使う
- `email_verified` claimは `email_verified_at IS NOT NULL` から生成する
- statusはアカウント利用可否

## 5. user_profiles

```sql
CREATE TABLE user_profiles (
  user_id UUID PRIMARY KEY REFERENCES users(id),
  family_name TEXT,
  given_name TEXT,
  family_name_kana TEXT,
  given_name_kana TEXT,
  gender TEXT,
  birthdate DATE,
  country_code CHAR(2),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### 備考

今回のID直登録では全項目必須入力とする。

ただし、DB上は将来のscope差分・外部IdP連携・部分更新に備えてnullableとする。

### Claim mapping

| DB | OIDC Claim |
|---|---|
| family_name | family_name |
| given_name | given_name |
| family_name_kana | family_name#ja-Kana-JP |
| given_name_kana | given_name#ja-Kana-JP |
| gender | gender |
| birthdate | birthdate |
| country_code | address.country |

## 6. user_password_credentials

```sql
CREATE TABLE user_password_credentials (
  user_id UUID PRIMARY KEY REFERENCES users(id),
  password_hash TEXT NOT NULL,
  password_changed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

## 7. pending_user_registrations

```sql
CREATE TABLE pending_user_registrations (
  id UUID PRIMARY KEY,
  email TEXT NOT NULL,
  password_hash TEXT NOT NULL,
  family_name TEXT NOT NULL,
  given_name TEXT NOT NULL,
  family_name_kana TEXT NOT NULL,
  given_name_kana TEXT NOT NULL,
  gender TEXT NOT NULL,
  birthdate DATE NOT NULL,
  country_code CHAR(2) NOT NULL,
  verification_token_hash TEXT UNIQUE NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_pending_user_registrations_email
ON pending_user_registrations(email);

CREATE INDEX idx_pending_user_registrations_expires_at
ON pending_user_registrations(expires_at);
```

## 8. clients

```sql
CREATE TABLE clients (
  id UUID PRIMARY KEY,
  client_id TEXT UNIQUE NOT NULL,
  client_name TEXT NOT NULL,
  client_type TEXT NOT NULL DEFAULT 'confidential',
  client_secret_hash TEXT,
  token_endpoint_auth_method TEXT NOT NULL DEFAULT 'client_secret_basic',
  require_pkce BOOLEAN NOT NULL DEFAULT true,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

## 9. client_redirect_uris

```sql
CREATE TABLE client_redirect_uris (
  id UUID PRIMARY KEY,
  client_id UUID NOT NULL REFERENCES clients(id),
  redirect_uri TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(client_id, redirect_uri)
);
```

## 10. client_allowed_scopes

```sql
CREATE TABLE client_allowed_scopes (
  client_id UUID NOT NULL REFERENCES clients(id),
  scope TEXT NOT NULL,
  PRIMARY KEY(client_id, scope)
);
```

## 11. authorization_requests

```sql
CREATE TABLE authorization_requests (
  id UUID PRIMARY KEY,
  request_id_hash TEXT UNIQUE NOT NULL,
  client_id UUID NOT NULL REFERENCES clients(id),
  redirect_uri TEXT NOT NULL,
  scope TEXT NOT NULL,
  state TEXT NOT NULL,
  nonce TEXT NOT NULL,
  code_challenge TEXT NOT NULL,
  code_challenge_method TEXT NOT NULL,
  requested_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at TIMESTAMPTZ NOT NULL,
  user_id UUID REFERENCES users(id),
  completed_at TIMESTAMPTZ
);
```

### 備考

未ログイン時の `/authorize` request を保持し、ログイン後に復元する。

## 12. authorization_codes

```sql
CREATE TABLE authorization_codes (
  id UUID PRIMARY KEY,
  code_hash TEXT UNIQUE NOT NULL,
  client_id UUID NOT NULL REFERENCES clients(id),
  user_id UUID NOT NULL REFERENCES users(id),
  redirect_uri TEXT NOT NULL,
  scope TEXT NOT NULL,
  nonce TEXT NOT NULL,
  code_challenge TEXT NOT NULL,
  code_challenge_method TEXT NOT NULL,
  auth_time TIMESTAMPTZ NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_authorization_codes_expires_at
ON authorization_codes(expires_at);
```

### 一回限り利用

```sql
UPDATE authorization_codes
SET used_at = now()
WHERE code_hash = $1
  AND used_at IS NULL
  AND expires_at > now()
RETURNING *;
```

## 13. refresh_tokens

```sql
CREATE TABLE refresh_tokens (
  id UUID PRIMARY KEY,
  token_hash TEXT UNIQUE NOT NULL,
  token_family_id UUID NOT NULL,
  client_id UUID NOT NULL REFERENCES clients(id),
  user_id UUID NOT NULL REFERENCES users(id),
  scope TEXT NOT NULL,
  parent_id UUID REFERENCES refresh_tokens(id),
  rotated_to_id UUID REFERENCES refresh_tokens(id),
  issued_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_used_at TIMESTAMPTZ,
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ,
  reuse_detected_at TIMESTAMPTZ
);

CREATE INDEX idx_refresh_tokens_family
ON refresh_tokens(token_family_id);

CREATE INDEX idx_refresh_tokens_user_client
ON refresh_tokens(user_id, client_id);
```

## 14. login_sessions

```sql
CREATE TABLE login_sessions (
  id UUID PRIMARY KEY,
  session_hash TEXT UNIQUE NOT NULL,
  user_id UUID NOT NULL REFERENCES users(id),
  auth_time TIMESTAMPTZ NOT NULL,
  amr TEXT NOT NULL,
  ip_address TEXT,
  user_agent TEXT,
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

## 15. email_otp_challenges

```sql
CREATE TABLE email_otp_challenges (
  id UUID PRIMARY KEY,
  challenge_id_hash TEXT UNIQUE NOT NULL,
  user_id UUID NOT NULL REFERENCES users(id),
  otp_hash TEXT NOT NULL,
  attempts INT NOT NULL DEFAULT 0,
  max_attempts INT NOT NULL DEFAULT 5,
  expires_at TIMESTAMPTZ NOT NULL,
  used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

## 16. signing_keys

```sql
CREATE TABLE signing_keys (
  id UUID PRIMARY KEY,
  kid TEXT UNIQUE NOT NULL,
  alg TEXT NOT NULL,
  public_jwk JSONB NOT NULL,
  private_key_parameter_name TEXT NOT NULL,
  status TEXT NOT NULL,
  not_before TIMESTAMPTZ NOT NULL,
  not_after TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### status

```text
active
standby
retired
```

秘密鍵本体はSSM SecureStringに保存し、DBにはParameter名を保存する。

## 17. user_withdrawals

```sql
CREATE TABLE user_withdrawals (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id),
  reason TEXT,
  ip_address TEXT,
  user_agent TEXT,
  withdrawn_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

退会時は以下を行う。

- users.status = withdrawn
- refresh token 全失効
- login session 全失効
- audit log記録
- user_withdrawals作成

## 18. audit_logs

```sql
CREATE TABLE audit_logs (
  id UUID PRIMARY KEY,
  event_type TEXT NOT NULL,
  user_id UUID REFERENCES users(id),
  client_id UUID REFERENCES clients(id),
  result TEXT NOT NULL,
  ip_address TEXT,
  user_agent TEXT,
  metadata JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_logs_user_created
ON audit_logs(user_id, created_at DESC);

CREATE INDEX idx_audit_logs_event_created
ON audit_logs(event_type, created_at DESC);
```

### event_type例

```text
user_registered
email_verified
login_password_success
login_password_failure
email_otp_sent
email_otp_success
email_otp_failure
op_session_created
authorization_code_issued
authorization_code_reused
token_issued
refresh_token_rotated
refresh_token_reuse_detected
logout
user_withdrawn
```

## 19. consents 将来用

MVPではconsent画面を実装しないが、外部公開に備えて将来追加する。

```sql
CREATE TABLE consents (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id),
  client_id UUID NOT NULL REFERENCES clients(id),
  granted_scopes TEXT NOT NULL,
  granted_claims JSONB,
  granted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  revoked_at TIMESTAMPTZ,
  UNIQUE(user_id, client_id)
);
```
