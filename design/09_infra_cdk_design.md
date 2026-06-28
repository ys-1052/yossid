# 09. AWS CDK TypeScript設計

## 1. 方針

AWSリソースは CDK TypeScript で管理する。

NeonリソースはCDKではなく、Terraform / OpenTofu または Neon APIで管理する。

理由:

- AWS CDKはAWSリソース管理に集中させる
- NeonはAWSリソースではない
- CDK Custom ResourceでNeon APIを叩く構成はMVPでは複雑

## 2. AWSリソース一覧

MVPで作成するAWSリソース:

- Lambda Function
- API Gateway HTTP API
- IAM Role / Policy
- SSM Parameter references
- CloudWatch Log Group
- SES identity 参照または作成
- EventBridge cleanup rule optional

ドメインを取得しないため、以下はMVPでは作成しない。

- Route 53 Hosted Zone
- ACM Certificate
- API Gateway Custom Domain

## 3. CDK構成

```text
infra/aws-cdk/
  bin/
    app.ts
  lib/
    oidc-provider-stack.ts
  package.json
  tsconfig.json
  cdk.json
```

## 4. Lambda

### Runtime

Go Lambda。

デプロイ方式は以下のどちらか。

#### 案A: Go binary zip

- `go build` でLinux向けbinary作成
- CDKでassetとしてzip

#### 案B: Docker bundling

- CDK bundlingでGo build
- 環境差分が少ない

MVPでは Docker bundling 推奨。

### Environment Variables

Lambda環境変数にはParameter名を入れる。

```text
ENV=prod
ISSUER=https://xxxxxxxx.execute-api.ap-northeast-1.amazonaws.com
DATABASE_URL_PARAMETER=/oidc-provider/prod/database/url
JWT_PRIVATE_KEY_PARAMETER=/oidc-provider/prod/jwt/private-key/kid-...
COOKIE_SIGNING_KEY_PARAMETER=/oidc-provider/prod/cookie/signing-key
TOKEN_PEPPER_PARAMETER=/oidc-provider/prod/token/pepper
OTP_PEPPER_PARAMETER=/oidc-provider/prod/otp/pepper
SES_FROM_EMAIL=no-reply@example.com
```

実際の秘密値はSSMから取得する。

## 5. API Gateway HTTP API

### Routing

すべてのrouteをLambdaにproxyする。

```text
ANY /{proxy+}
```

または明示route。

MVPではproxy integrationでよい。

### Stage

default stageを利用し、issuer URLを短くする。

注意:

- API Gateway URLがissuerになる
- APIを再作成するとissuerが変わる
- stack削除に注意

## 6. IAM

Lambda Roleに許可する。

### SSM

```text
ssm:GetParameter
```

対象Parameterのみ。

### SES

```text
ses:SendEmail
ses:SendRawEmail
```

送信元Identityに限定できるなら限定する。

### CloudWatch Logs

Lambda基本権限。

## 7. SSM Parameter Store SecureString

CDKでParameter名を参照するが、SecureStringの値は手動投入または別手順で管理する。

MVPではCDKで秘密値そのものは作成しない。

例:

```bash
aws ssm put-parameter \
  --name /oidc-provider/prod/database/url \
  --type SecureString \
  --value "postgres://..."
```

## 8. CloudWatch Logs

Log Groupを明示作成し、保持期間を設定する。

MVP推奨:

```text
retention: 30 days
```

外部公開時は延長を検討。

## 9. SES

MVPではメール確認・OTP送信に利用する。

必要な設定:

- 送信元メールアドレスまたはドメインの検証
- sandboxの場合、送信先制限に注意
- 本番外部公開時はsandbox解除申請

ドメインを取得しない場合、送信元メールアドレスの選定が必要。

## 10. Cleanup

期限切れデータの削除はMVPでは必須ではないが、DB肥大化防止のため定期処理を検討する。

対象:

- expired pending_user_registrations
- expired authorization_requests
- expired authorization_codes
- expired email_otp_challenges
- expired login_sessions

実装候補:

- EventBridge Scheduler
- cleanup用Lambda
- 通常リクエスト時のlazy cleanup

MVPではlazy cleanupでも可。

## 11. CDK Stack出力

Outputs:

- ApiEndpoint
- Issuer
- SsmParameterNames

## 13. Next.js Frontendリソース

認証UIをNext.jsで作り込むため、MVPにFrontend hostingを追加する。

### 方針

- Next.js App Router
- TypeScript
- Tailwind CSS
- CloudFrontをfront doorにする
- Go OIDC APIとNext.js UIを同一originで公開する
- OIDC/OAuth protocol endpointはGo Lambdaが担当する
- UI rendering / form / client-side validationはNext.jsが担当する

### AWS構成候補

#### 推奨案: CloudFront + Path Routing

```text
CloudFront
  ├─ /authorize, /token, /userinfo, /.well-known/*, /jwks.json
  │    -> API Gateway HTTP API -> Go Lambda
  │
  └─ /register, /login, /mfa/*, /email/verify, /logout, /_next/*
       -> Next.js hosting origin
```

Next.js hosting originの候補:

1. AWS Amplify Hosting
2. OpenNext on AWS
3. Lambda Web Adapter + Next.js standalone

MVPでは、CDKで統一管理したい場合は OpenNext / Lambda Web Adapter 系を検討する。
運用の簡単さを優先する場合は Amplify Hosting も候補。

### CDK管理方針

Go OIDC API:

- CDKで管理

CloudFront:

- CDKで管理

Next.js Hosting:

- CDK管理を優先するならOpenNext系
- まず簡単に始めるならAmplify Hostingも可
- ただしAmplifyを使う場合、CDK管理の範囲が分かれる点に注意

### Issuer

CloudFront distribution domainをissuerとする。

```text
ISSUER=https://dxxxxxxxxxxxxx.cloudfront.net
```

API Gateway URLを直接issuerにしない。

理由:

- Next.js UIとGo APIを同一origin化するため
- Cookie管理を単純化するため
- 将来CloudFrontでWAFやcache policyを入れやすくするため

### Cache Policy

OIDC API endpointは原則cacheしない。

No cache:

- `/authorize`
- `/token`
- `/userinfo`
- `/register`
- `/login`
- `/mfa/*`
- `/email/verify`
- `/logout`

Cache可能:

- `/.well-known/openid-configuration`
- `/jwks.json`
- `/_next/static/*`
- static assets

### Cookie Forwarding

CloudFrontはOP session cookie / CSRF cookieを必要なoriginへ転送する。

対象cookie例:

```text
op_session
csrf_token
auth_request
mfa_challenge
```


## 14. Version Pinning

AWS CDKは以下を基準バージョンとして開始する。

```text
aws-cdk-lib 2.260.0
constructs 10.6.0
typescript 6.0.3
```

Next.js on AWSでOpenNextを採用する場合:

```text
@opennextjs/aws 4.0.2
```

CDK側の `package-lock.json` は必ずcommitする。
