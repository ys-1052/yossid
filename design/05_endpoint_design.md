# 05. Endpoint設計

## 1. Endpoint一覧

| Method | Path | 認証 | 説明 |
|---|---|---|---|
| GET | `/.well-known/openid-configuration` | 不要 | Discovery |
| GET | `/jwks.json` | 不要 | JWKS |
| GET | `/authorize` | OP session | 認可エンドポイント |
| POST | `/token` | client認証 | Tokenエンドポイント |
| GET | `/userinfo` | Bearer token | UserInfo |
| GET | `/register` | 不要 | 登録画面 |
| POST | `/register` | CSRF | 登録開始 |
| GET | `/email/verify` | token | メール確認 |
| GET | `/login` | 不要 | ログイン画面 |
| POST | `/login` | CSRF | password認証 |
| GET | `/mfa/email` | temporary challenge | OTP入力画面 |
| POST | `/mfa/email/verify` | CSRF | OTP検証 |
| POST | `/logout` | OP session | ログアウト |
| GET | `/healthz` | 不要 | ヘルスチェック |

## 2. `GET /.well-known/openid-configuration`

### Response

```json
{
  "issuer": "https://xxxxxxxx.execute-api.ap-northeast-1.amazonaws.com",
  "authorization_endpoint": "https://xxxxxxxx.execute-api.ap-northeast-1.amazonaws.com/authorize",
  "token_endpoint": "https://xxxxxxxx.execute-api.ap-northeast-1.amazonaws.com/token",
  "userinfo_endpoint": "https://xxxxxxxx.execute-api.ap-northeast-1.amazonaws.com/userinfo",
  "jwks_uri": "https://xxxxxxxx.execute-api.ap-northeast-1.amazonaws.com/jwks.json",
  "response_types_supported": ["code"],
  "grant_types_supported": ["authorization_code", "refresh_token"],
  "subject_types_supported": ["public"],
  "id_token_signing_alg_values_supported": ["RS256"],
  "scopes_supported": ["openid", "email", "profile", "address", "offline_access"],
  "token_endpoint_auth_methods_supported": ["client_secret_basic"],
  "code_challenge_methods_supported": ["S256"]
}
```

### Cache

- `Cache-Control: public, max-age=3600`

## 3. `GET /jwks.json`

### Response

```json
{
  "keys": [
    {
      "kty": "RSA",
      "kid": "...",
      "use": "sig",
      "alg": "RS256",
      "n": "...",
      "e": "AQAB"
    }
  ]
}
```

### Cache

- `Cache-Control: public, max-age=3600`

鍵ローテーション時はactive + retired直後の公開鍵を一定期間返す。

## 4. `GET /authorize`

### Query Parameters

必須:

- response_type
- client_id
- redirect_uri
- scope
- state
- nonce
- code_challenge
- code_challenge_method

任意:

- prompt
- max_age
- login_hint

### Validation

- `response_type=code`
- `scope` に `openid` を含む
- `redirect_uri` は登録済みURIと完全一致
- `state` 必須
- `nonce` 必須
- `code_challenge_method=S256`
- clientがactive
- requested scopesがclientに許可済み

### Success

```text
302 Location: {redirect_uri}?code=...&state=...
```

### Error

redirect_uriが検証できる場合のみ、redirect_uriへエラーを返す。

redirect_uriが不正な場合は、エラーページを表示し、redirectしない。

## 5. `POST /token`

### Content-Type

```text
application/x-www-form-urlencoded
```

### Client Auth

MVPでは `client_secret_basic` のみ。

### authorization_code request

```text
grant_type=authorization_code
code=...
redirect_uri=...
code_verifier=...
```

### refresh_token request

```text
grant_type=refresh_token
refresh_token=...
```

### Response

```json
{
  "access_token": "...",
  "token_type": "Bearer",
  "expires_in": 900,
  "refresh_token": "...",
  "id_token": "..."
}
```

### Error

OAuth 2.0形式。

```json
{
  "error": "invalid_grant",
  "error_description": "..."
}
```

## 6. `GET /userinfo`

### Request

```text
Authorization: Bearer <access_token>
```

### Response例

```json
{
  "sub": "user_...",
  "email": "user@example.com",
  "email_verified": true,
  "family_name": "苗字",
  "given_name": "名前",
  "family_name#ja-Kana-JP": "ファミリーネーム",
  "given_name#ja-Kana-JP": "ギブンネーム",
  "gender": "male",
  "birthdate": "1990-01-01",
  "address": {
    "country": "Japan"
  }
}
```

scopeに応じて返すclaimを制御する。

## 7. `GET /register`

登録フォームHTMLを返す。

入力項目:

- email
- password
- password confirmation
- family_name
- given_name
- family_name_kana
- given_name_kana
- gender
- birthdate
- country_code

CSRF tokenを埋め込む。

## 8. `POST /register`

登録情報を受け取り、pending登録を作成する。

### Validation

- email形式
- password policy
- family_name必須
- given_name必須
- kana必須
- gender必須
- birthdate必須
- country_code必須
- password confirmation一致

### Response

- メール確認案内画面
- 既存ユーザー有無は露出しすぎない

## 9. `GET /email/verify`

### Query

```text
token=...
```

### 処理

- token hash化
- pending_user_registrations照合
- expires_at確認
- used_at確認
- users作成
- user_profiles作成
- user_password_credentials作成
- pendingをused_at更新

## 10. `GET /login`

ログイン画面HTMLを返す。

`return_to` または `auth_request_id` により、認可フロー復帰先を保持する。

## 11. `POST /login`

email/passwordを検証し、OTP challengeを作成する。

成功時はOTP入力画面へ。

## 12. `POST /mfa/email/verify`

OTPを検証し、OP sessionを発行する。

Cookie:

```text
Set-Cookie: op_session=...; HttpOnly; Secure; SameSite=Lax; Path=/
```

## 13. `POST /logout`

OP sessionを失効し、cookieを削除する。

## 14. `GET /healthz`

DB接続までは確認しない軽量ヘルスチェックとする。

Response:

```json
{
  "status": "ok"
}
```

必要に応じて `/readyz` を別途追加し、DB接続確認を行う。
