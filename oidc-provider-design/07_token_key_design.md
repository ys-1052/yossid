# 07. Token / Key設計

## 1. Token一覧

| Token | 形式 | DB保存 | 有効期限 |
|---|---|---|---|
| Authorization Code | opaque random | hash保存 | 1〜5分 |
| ID Token | JWT | 保存しない | 5〜15分 |
| Access Token | JWT | 原則保存しない | 5〜15分 |
| Refresh Token | opaque random | hash保存 | 7〜30日 |
| OP Session ID | opaque random | hash保存 | 12〜24時間 |
| Email Verification Token | opaque random | hash保存 | 15〜60分 |
| Email OTP | 6桁数字 | hash保存 | 5分 |

## 2. Authorization Code

### 形式

- 256bit以上のランダム値
- URL safe base64

### 保存

DBには以下のみ保存。

```text
sha256(code + pepper)
```

### 有効期限

MVP推奨:

```text
5分
```

### 利用条件

- 未使用
- 期限内
- client一致
- redirect_uri一致
- PKCE検証成功

## 3. ID Token

### 形式

JWT。

署名アルゴリズム:

```text
RS256
```

将来候補:

```text
ES256
```

### 必須Claim

```text
iss
sub
aud
exp
iat
```

### 条件付きClaim

```text
nonce
auth_time
amr
azp
```

### scopeに応じたClaim

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

### 有効期限

MVP推奨:

```text
15分
```

## 4. Access Token

### 形式

MVPではJWT。

### Claim

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

### 有効期限

MVP推奨:

```text
15分
```

### 保存

MVPでは保存しない。

失効が必要になった場合は以下を追加検討。

- access_token_jti blacklist
- introspection endpoint
- opaque access token化

## 5. Refresh Token

### 形式

opaque random token。

### 保存

DBにはhash保存。

```text
sha256(refresh_token + pepper)
```

### 有効期限

MVP推奨:

```text
30日
```

### Rotation

Refresh Token利用時に必ず新しいRefresh Tokenを発行する。

```text
old refresh token
  -> revoked_at set
  -> new refresh token issued
  -> old.rotated_to_id = new.id
```

### Reuse Detection

既にrevoked / rotated済みのrefresh tokenが利用された場合:

- `reuse_detected_at` を設定
- 同一 `token_family_id` のtokenを全失効
- audit log記録
- token発行拒否

## 6. OP Session

### Cookie

```text
op_session=<random>; HttpOnly; Secure; SameSite=Lax; Path=/
```

### 保存

DBにはsession hashのみ保存。

### 有効期限

MVP推奨:

```text
12時間
```

### セッション更新

MVPではsliding sessionは実装しない。

将来検討:

- session refresh
- max_age対応
- prompt=login対応
- remember me

## 7. Email Verification Token

### 形式

opaque random token。

### 有効期限

MVP推奨:

```text
30分
```

### 保存

DBにはhash保存。

### 利用

1回限り。

## 8. Email OTP

### 形式

6桁数字。

### 有効期限

MVP推奨:

```text
5分
```

### 最大試行回数

```text
5回
```

### 保存

DBにはhash保存。

## 9. Signing Key

### 方針

- RS256
- kid必須
- active keyで署名
- JWKSで公開鍵を公開
- 秘密鍵はSSM SecureString
- DBには秘密鍵のParameter名を保存する

### signing_keys

```text
kid
alg
public_jwk
private_key_parameter_name
status
not_before
not_after
```

### status

```text
standby
active
retired
```

## 10. 鍵ローテーション

### 初期MVP

手動ローテーションでよい。

### 推奨手順

```text
1. 新しい鍵を生成
2. signing_keysにstandbyとして登録
3. JWKSにstandby公開鍵も表示
4. 一定時間後activeへ切替
5. 古い鍵をretiredへ
6. 既存token期限切れまでJWKSに残す
7. 期限経過後JWKSから除外
```

## 11. SSM Parameter設計

### 例

```text
/oidc-provider/prod/database/url
/oidc-provider/prod/jwt/private-key/kid-2026-01
/oidc-provider/prod/cookie/signing-key
/oidc-provider/prod/cookie/encryption-key
/oidc-provider/prod/token/pepper
/oidc-provider/prod/otp/pepper
```

### Lambda IAM

Lambdaには必要なParameterだけ `ssm:GetParameter` を許可する。

## 12. Token Response例

```json
{
  "access_token": "eyJ...",
  "token_type": "Bearer",
  "expires_in": 900,
  "refresh_token": "rt_...",
  "id_token": "eyJ..."
}
```

## 13. ログ出力禁止値

以下は絶対にログ出力しない。

- authorization code
- access token
- refresh token
- id token
- client_secret
- password
- OTP
- session cookie
- verification token
