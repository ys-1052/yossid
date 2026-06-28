# 02. Protocol Scope設計

## 1. 対応するフロー

MVPでは以下のみ対応する。

- OAuth 2.0 Authorization Code Flow
- PKCE S256
- OIDC ID Token
- Refresh Token Rotation

対応しない。

- Implicit Flow
- Hybrid Flow
- Resource Owner Password Credentials
- Client Credentials Grant
- Device Flow
- CIBA
- Dynamic Client Registration
- FAPI
- PAR
- RAR
- Federation

## 2. response_type

対応:

```text
response_type=code
```

非対応:

```text
token
id_token
code id_token
code token
id_token token
code id_token token
```

## 3. grant_type

対応:

```text
authorization_code
refresh_token
```

非対応:

```text
client_credentials
password
urn:ietf:params:oauth:grant-type:device_code
```

## 4. PKCE

MVPではすべてのClientにPKCE S256を必須とする。

許可:

```text
code_challenge_method=S256
```

拒否:

```text
plain
未指定
```

## 5. Client認証

MVPでは confidential client のみ対応する。

対応:

```text
client_secret_basic
```

将来候補:

```text
client_secret_post
private_key_jwt
tls_client_auth
self_signed_tls_client_auth
none
```

MVPでは public client は対応しないが、将来拡張できるよう `clients.client_type` は持つ。

## 6. Scope

MVPで対応するscope:

```text
openid
email
profile
address
offline_access
```

### openid

必須scope。OIDC requestには必ず含める。

返却Claim:

```text
sub
```

### email

返却Claim:

```text
email
email_verified
```

### profile

返却Claim:

```text
family_name
given_name
family_name#ja-Kana-JP
given_name#ja-Kana-JP
gender
birthdate
```

### address

返却Claim:

```json
{
  "address": {
    "country": "Japan"
  }
}
```

DBでは `country_code` を保持し、レスポンス時に表示名へ変換する。

### offline_access

Refresh Token発行を許可するscope。

MVPでは first-party confidential client のみ利用を許可する。

## 7. Claims方針

### 標準Claim

- sub
- email
- email_verified
- family_name
- given_name
- gender
- birthdate
- address

### 仮名Claim

OIDCの言語タグ付きClaimとして返す。

```text
family_name#ja-Kana-JP
given_name#ja-Kana-JP
```

DB上は以下で保持する。

```text
family_name_kana
given_name_kana
```

### address

MVPでは国のみ。

```json
{
  "address": {
    "country": "Japan"
  }
}
```

DB上はISO 3166-1 alpha-2の `country_code` を保持する。

例:

```text
JP
```

## 8. ID Token Claim

MVPのID Tokenには以下を含める。

必須:

```text
iss
sub
aud
exp
iat
```

条件付き:

```text
nonce
auth_time
amr
azp
```

scopeに応じて追加:

```text
email
email_verified
family_name
given_name
family_name#ja-Kana-JP
given_name#ja-Kana-JP
gender
birthdate
address
```

MVPでは認証方式として以下を設定する。

```json
"amr": ["pwd", "email"]
```

## 9. Access Token

MVPではJWT Access Tokenとする。

含めるClaim:

```text
iss
sub
aud
exp
iat
jti
client_id
scope
```

API側はJWKSで署名検証する。

## 10. Refresh Token

Refresh Tokenはopaque random tokenとする。

DBにはハッシュのみ保存する。

仕様:

- rotation必須
- 1回利用したら古いtokenは無効化
- 再利用検知時は関連tokenを失効する
- 有効期限を持つ
- `offline_access` scopeが許可された場合のみ発行する

## 11. Discovery Metadata

`/.well-known/openid-configuration` で返すmetadata例。

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
  "code_challenge_methods_supported": ["S256"],
  "claims_supported": [
    "sub",
    "email",
    "email_verified",
    "family_name",
    "given_name",
    "family_name#ja-Kana-JP",
    "given_name#ja-Kana-JP",
    "gender",
    "birthdate",
    "address",
    "auth_time",
    "amr"
  ]
}
```

## 12. Issuer方針

MVPでは独自ドメインを取得しないため、API GatewayのURLをissuerとする。

注意点:

- issuerは後から変更しづらい
- issuerが変わるとClient設定・Token検証に影響する
- 将来外部公開時は独自ドメインへの移行を検討する
- MVP段階ではAPI Gatewayの削除・再作成を避ける
