# 11. Test Plan

## 1. 方針

OIDC Providerはセキュリティ影響が大きいため、unit test / integration test / security test を分けて実施する。

将来、OpenID Foundation Conformance Suiteでの検証を目指す。

## 2. Unit Test

### Validation

- email validation
- password policy
- kana validation
- gender validation
- birthdate validation
- country_code validation

### OAuth/OIDC

- redirect_uri完全一致
- scope validation
- PKCE S256 validation
- response_type=codeのみ許可
- nonce必須
- state必須
- openid scope必須

### Token

- ID Token claim生成
- Access Token claim生成
- JWKS kid selection
- exp / iat / auth_time
- amr

### Security

- password hash verify
- token hash
- OTP hash
- CSRF token
- cookie生成

## 3. Integration Test

### Registration

- 正常登録
- メール確認正常
- 期限切れverification token
- 使用済みverification token
- 重複email
- 不正入力

### Login

- 正常login
- password不一致
- 存在しないemail
- withdrawn user
- OTP正常
- OTP不一致
- OTP期限切れ
- OTP試行回数超過

### Authorization Code Flow

- /authorize 正常
- 未ログイン時loginへ
- login後authorize復帰
- code発行
- /tokenでtoken発行
- ID Token検証
- Access Token検証
- UserInfo取得

### Error Case

- invalid client_id
- invalid redirect_uri
- invalid scope
- missing state
- missing nonce
- missing code_challenge
- code_challenge_method=plain
- reused authorization code
- expired authorization code
- invalid code_verifier
- invalid client_secret

### Refresh Token

- 正常rotation
- 古いrefresh token再利用
- expired refresh token
- revoked refresh token
- token family失効

## 4. Security Test

- open redirectがないこと
- CSRF対策
- session fixation対策
- SQL injection対策
- token/code/password/OTPがログに出ないこと
- cookie属性確認
- UserInfoがscope外claimを返さないこと
- withdrawn userがtoken発行できないこと
- refresh token reuse detection

## 5. Conformance Test 将来

将来的にOpenID Provider conformance testを実施する。

MVPでは自己テストまで。

目標:

- Basic OP相当
- Authorization Code Flow
- Discovery
- JWKS
- UserInfo

## 6. Manual Test

### RP連携

テスト用RPを用意する。

候補:

- 自作簡易RP
- oauth2-proxy
- Next.js RP
- oidc-client系ライブラリ

確認項目:

- discovery読み込み
- login redirect
- code exchange
- id_token validation
- userinfo取得

## 7. Load / Cost Test

MVPでは軽量確認のみ。

- Lambda cold start
- Neon scale-to-zero後の初回接続
- token endpoint latency
- concurrent login
- DB connection数

## 8. Test Data

- active user
- withdrawn user
- confidential client
- invalid client
- expired auth code
- reused refresh token
