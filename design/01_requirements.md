# 01. 要件定義

## 1. 目的

個人開発アプリで共通利用できる自作 OIDC Provider を実装する。

初期利用は個人・少人数に限定するが、将来的に外部 RP / Client に対して OIDC Provider として公開できるよう、仕様に沿った設計を行う。

## 2. ゴール

### MVPゴール

- OIDC Providerとして最低限のフローが動作する
- Authorization Code Flow + PKCE S256 に対応する
- email/password + メールOTP 2FA による認証ができる
- ID Token / Access Token / Refresh Token を発行できる
- Refresh Token Rotation を実装する
- Discovery / JWKS / UserInfo を提供する
- Neon Postgres に認証・認可データを永続化する
- AWS Lambda上で動作する
- AWSリソースは CDK TypeScript で管理する
- 秘密情報は SSM Parameter Store SecureString で管理する

### 将来ゴール

- 外部Client登録
- Consent画面
- Client管理画面
- パスワードリセット
- Passkey
- FAPI / PAR / RAR
- 独自ドメイン
- OpenID Provider conformance test 通過
- より厳格なネットワーク制御

## 3. 非ゴール

MVPでは以下を実装しない。

- FAPI対応
- PAR
- RAR
- Dynamic Client Registration
- Device Flow
- CIBA
- Federation
- Implicit Flow
- Hybrid Flow
- Resource Owner Password Credentials
- Client Credentials Grant
- Passkey
- 外部IdPログイン
- 管理画面
- パスワードリセット
- Consent画面
- 独自ドメイン
- 商用レベルの高可用構成
- DB閉域接続

## 4. 利用者

### MVP

- 開発者本人
- 少人数のテストユーザー
- first-party client のみ

### 将来

- 外部 RP / Client
- 複数アプリケーション
- third-party client

## 5. Client方針

MVPでは confidential client のみ対応する。

- `client_secret_basic` を基本とする
- PKCE S256 は confidential client でも必須
- redirect_uri は完全一致
- Client設定はコード管理・seed管理とする
- Dynamic Client Registration は実装しない

将来に備え、DB上は以下を持てるようにする。

- client_type
- token_endpoint_auth_method
- require_pkce
- allowed_grant_types
- allowed_response_types
- allowed_scopes
- jwks_uri
- client_secret_hash
- status

## 6. ユーザー登録方針

今回実装する自前ID登録では以下をすべて必須入力とする。

- email
- password
- family_name
- given_name
- family_name_kana
- given_name_kana
- gender
- birthdate
- country_code

ただし、DB上のプロフィール項目は将来のRP・scope差分に対応できるよう optional として扱える設計にする。

## 7. ユーザーステータス方針

メール確認前ユーザーは `users` には登録せず、`pending_user_registrations` で管理する。

`users` はメール確認済みユーザーのみを保持する。

`users.status` はアカウント利用可否を管理する。

MVPで利用するstatus:

- `active`
- `withdrawn`

将来追加候補:

- `disabled`
- `locked`

`email_verified_at` はメール確認状態を表し、`status` はアカウント利用可否を表す。

## 8. インフラ前提

- AWSに寄せる
- 実行環境は AWS Lambda
- API Gateway HTTP API を利用
- 独自ドメインはMVPでは使わない
- issuer は API Gateway のURLを固定して利用する
- DBは Neon Postgres
- Neonへの通信は public endpoint + TLS `verify-full`
- 秘密情報は SSM Parameter Store SecureString
- メール送信は SES
- AWSリソースは CDK TypeScript
- Neonリソースは Terraform でコード管理する

## 9. セキュリティ基本方針

- Authorization Code Flow のみ
- PKCE S256 必須
- redirect_uri 完全一致
- state 必須
- nonce 必須
- refresh token rotation 必須
- token/code/session id はハッシュ保存
- password は Argon2id または bcrypt でハッシュ化
- DBユーザーは runtime / migration で分離
- ログに token / code / password / OTP を出さない
- メールOTP 2FA 必須
- OP session cookie は HttpOnly / Secure / SameSite=Lax
- 監査ログを保存する

## 10. 外部公開時の追加条件

外部公開する場合は以下を追加検討する。

- 独自ドメイン
- Consent画面
- Client管理画面
- Clientごとのscope制御
- 利用規約・プライバシーポリシー
- Abuse対策
- WAF
- Rate limit強化
- IP allowlist / PrivateLink / RDS等のDB移行
- OpenID Conformance Suiteによる検証

## 11. 確定事項

以下を確定とする。

- Neonのコード管理は Terraform
- Go HTTP Frameworkは Echo
- Lambda上では Echo を aws-lambda-go-api-proxy/echo 等のAdapterで実行する
