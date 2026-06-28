# 03. Architecture設計

## 1. 全体構成

```text
[Browser / RP]
    |
    | HTTPS
    v
[API Gateway HTTP API]
    |
    v
[AWS Lambda: Go Echo OIDC Provider]
    |
    | TLS verify-full
    v
[Neon Postgres]

[AWS Lambda]
    |
    | SES SendEmail
    v
[Amazon SES]

[AWS Lambda]
    |
    | GetParameter
    v
[SSM Parameter Store SecureString]

[CloudWatch Logs]
```

## 2. AWS構成

### API Gateway HTTP API

公開エンドポイントを提供する。

主なパス:

- `GET /.well-known/openid-configuration`
- `GET /jwks.json`
- `GET /authorize`
- `POST /token`
- `GET /userinfo`
- `GET /register`
- `POST /register`
- `GET /email/verify`
- `GET /login`
- `POST /login`
- `POST /mfa/email/send`
- `POST /mfa/email/verify`
- `POST /logout`
- `GET /healthz`

### Lambda

Go + Echoで実装するOIDC Provider本体。

役割:

- OIDC/OAuth endpoint
- 登録画面
- ログイン画面
- メール確認
- メールOTP
- Token発行
- UserInfo
- 監査ログ出力

### Neon Postgres

永続化データを保持する。

主なデータ:

- users
- user_profiles
- password_credentials
- clients
- redirect_uris
- authorization_requests
- authorization_codes
- refresh_tokens
- login_sessions
- email_verification_tokens
- email_otp_challenges
- signing_keys metadata
- audit_logs
- user_withdrawals

### SES

以下のメールを送信する。

- メールアドレス確認
- メールOTP

MVPでは送信元メールアドレス・ドメイン検証が必要。

### SSM Parameter Store SecureString

保持する秘密情報:

- Neon DATABASE_URL runtime
- Neon DATABASE_URL migration はCI/ローカル用に別管理
- JWT署名秘密鍵
- cookie signing key
- cookie encryption key
- email token pepper
- OTP token pepper

### CloudWatch Logs

以下を出力する。

- application logs
- error logs
- security event logs

注意:

- token
- authorization code
- refresh token
- password
- OTP
- session cookie

は絶対にログに出力しない。

## 3. Lambda構成方針

MVPでは Echo アプリケーションを単一Lambdaで実装する。

理由:

- Cookie / session / OIDC flow の取り回しが簡単
- Fosite provider 初期化を共通化しやすい
- 小規模MVPでは運用が楽

将来的な分割候補:

- public endpoint Lambda
- email worker Lambda
- cleanup worker Lambda
- admin Lambda

## 4. パッケージ構成

```text
cmd/lambda/main.go

internal/config
internal/http
  authorize_handler.go
  token_handler.go
  userinfo_handler.go
  discovery_handler.go
  jwks_handler.go
  register_handler.go
  email_verify_handler.go
  login_handler.go
  mfa_handler.go
  logout_handler.go
  health_handler.go

internal/oidc
  provider.go
  fosite_factory.go
  claims.go
  token.go
  jwks.go

internal/authn
  password.go
  session.go
  email_otp.go
  email_verification.go

internal/storage
  storage.go
  postgres/
    users.go
    profiles.go
    clients.go
    auth_requests.go
    auth_codes.go
    refresh_tokens.go
    sessions.go
    otp.go
    signing_keys.go
    audit_logs.go

internal/security
  random.go
  hash.go
  password_hash.go
  cookie.go
  csrf.go
  rate_limit.go

internal/mail
  ses.go
  templates.go

internal/audit
  logger.go

migrations/
infra/aws-cdk/
infra/neon/
docs/
```

## 5. Fosite利用方針

Fositeには以下を任せる。

- authorize request検証
- token request検証
- PKCE検証
- client検証
- scope処理
- token発行フロー補助
- OIDC handler

自前実装するもの:

- ユーザー登録
- email/password認証
- メールOTP
- OP session
- consent省略ポリシー
- DB storage実装
- Claims mapping
- JWK管理
- メール送信
- 監査ログ
- rate limit

## 6. Storage設計方針

Postgresを使う。

理由:

- authorization code の一回限り利用をトランザクションで保証しやすい
- refresh token rotation を実装しやすい
- client / scope / consent の関係を扱いやすい
- 将来外部公開に向けた管理がしやすい

## 7. 通信方式

### Client / Browser -> API Gateway

- HTTPS
- API Gateway標準ドメインを使用
- MVPでは独自ドメインなし

### Lambda -> Neon

- public endpoint
- TLS必須
- `sslmode=verify-full`
- pooled connection endpointを利用
- runtime DB userは最小権限

### Lambda -> SES

- AWS SDK
- IAM Roleで権限付与

### Lambda -> SSM

- AWS SDK
- 必要なParameterだけ読み取り許可

## 8. 独自ドメインなしの注意

MVPではissuerがAPI GatewayのURLになる。

例:

```text
https://xxxxxxxx.execute-api.ap-northeast-1.amazonaws.com
```

注意:

- API Gatewayを作り直すとissuerが変わる
- issuerが変わるとClient側の検証が失敗する
- 将来外部公開時は独自ドメイン移行計画が必要
- MVP中はAPI Gateway IDを安定させる

## 10. Next.js Frontend追加構成

認証UIをNext.jsで作り込む場合、OIDC/OAuthプロトコル処理はGo Lambda/Fositeに残し、UIをNext.jsに分離する。

### 推奨構成

```text
[Browser / RP]
    |
    | HTTPS
    v
[CloudFront Distribution]
    |
    | path routing
    |
    ├─ /authorize
    ├─ /token
    ├─ /userinfo
    ├─ /.well-known/*
    ├─ /jwks.json
    |    -> API Gateway -> Go Lambda OIDC API
    |
    ├─ /register
    ├─ /login
    ├─ /mfa/*
    ├─ /email/verify
    ├─ /logout
    ├─ /account/*
    ├─ /_next/*
    |    -> Next.js Frontend on AWS
```

### なぜCloudFrontをfront doorにするか

独自ドメインを取得しない場合でも、CloudFrontのdistribution domainをissuer originとして利用できる。

```text
https://dxxxxxxxxxxxxx.cloudfront.net
```

Next.js UIとGo OIDC APIを同一originに見せることで、以下をシンプルにできる。

- OP session cookie
- CSRF cookie
- redirect_uri後の画面遷移
- SameSite=Lax cookie
- `/authorize` から `/login` への遷移

### 注意

CloudFront Distributionを削除・再作成するとdomainが変わるため、issuerが変わる。

外部公開前には独自ドメイン取得を再検討する。

## 11. Echo on Lambda構成

Go backendは Echo を利用する。

```text
API Gateway HTTP API
  -> Lambda
    -> aws-lambda-go-api-proxy/echo
      -> Echo Router
        -> handlers
          -> services
            -> repositories
              -> Neon Postgres
```

### 採用理由

- GoでのHTTP routingが書きやすい
- Middlewareを使いやすい
- HTML rendering / JSON response / form handlingが扱いやすい
- Lambda移行前にローカルHTTP serverとして動かしやすい
- OIDC endpointと認証UI/BFF endpointを同一アプリ内で整理しやすい

### 注意点

- Lambda上ではlong-running serverではなくAdapter経由でrequestを処理する
- DB connection poolはLambda実行環境ごとに作られる
- Echo middlewareで機微情報をログに出さない
- Echo default error handlerはOAuth error response要件に合わせて調整する
