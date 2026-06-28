# 17. Dependency Versions

## 1. 方針

このプロジェクトでは、実装開始時点で確認した公開最新版をピン留めする。

認証基盤のため、`latest` 指定のまま運用しない。

```text
OK:
- go.mod に具体バージョンを記録
- package-lock.json をcommit
- .terraform.lock.hcl をcommit
- Renovate / Dependabot で更新PRを出す

NG:
- package.jsonで "*" を使う
- 本番buildで `go get ...@latest` を実行する
- Terraform providerをversion未指定にする
```

## 2. Version確認日

```text
2026-06-28
```

## 3. Backend / Go

### Go

```text
Go 1.26
```

Lambda runtime:

```text
provided.al2023
```

### Go modules

```text
github.com/labstack/echo/v4 v4.15.4
github.com/ory/fosite v0.49.0
github.com/aws/aws-lambda-go v1.54.0
github.com/awslabs/aws-lambda-go-api-proxy v0.16.2
github.com/jackc/pgx/v5 v5.10.0
github.com/aws/aws-sdk-go-v2/config v1.32.25
github.com/aws/aws-sdk-go-v2/service/ssm v1.69.3
github.com/aws/aws-sdk-go-v2/service/sesv2 v1.62.4
golang.org/x/crypto v0.53.0
```

### go.mod例

```go
module github.com/ys-1052/yossid/backend

go 1.26

require (
    github.com/aws/aws-lambda-go v1.54.0
    github.com/aws/aws-sdk-go-v2/config v1.32.25
    github.com/aws/aws-sdk-go-v2/service/sesv2 v1.62.4
    github.com/aws/aws-sdk-go-v2/service/ssm v1.69.3
    github.com/awslabs/aws-lambda-go-api-proxy v0.16.2
    github.com/jackc/pgx/v5 v5.10.0
    github.com/labstack/echo/v4 v4.15.4
    github.com/ory/fosite v0.49.0
    golang.org/x/crypto v0.53.0
)
```

### Tool

```text
sqlc v1.31.1
```

## 4. Frontend / Next.js

### Runtime / Framework

```text
next 16.2.9
react 19.2.7
react-dom 19.2.7
typescript 6.0.3
```

### UI / Form

```text
tailwindcss 4.3.1
@tailwindcss/postcss 4.3.1
react-hook-form 7.80.0
@hookform/resolvers 5.2.2
zod 4.4.3
lucide-react 1.21.0
@types/react 19.2.17
```

### package.json例

```json
{
  "dependencies": {
    "@hookform/resolvers": "5.2.2",
    "@tailwindcss/postcss": "4.3.1",
    "lucide-react": "1.21.0",
    "next": "16.2.9",
    "react": "19.2.7",
    "react-dom": "19.2.7",
    "react-hook-form": "7.80.0",
    "tailwindcss": "4.3.1",
    "zod": "4.4.3"
  },
  "devDependencies": {
    "@types/react": "19.2.17",
    "typescript": "6.0.3"
  }
}
```

## 5. AWS CDK TypeScript

```text
aws-cdk-lib 2.260.0
constructs 10.6.0
typescript 6.0.3
```

### package.json例

```json
{
  "dependencies": {
    "aws-cdk-lib": "2.260.0",
    "constructs": "10.6.0"
  },
  "devDependencies": {
    "typescript": "6.0.3"
  }
}
```

## 6. Next.js on AWS

CDK統合でNext.jsをAWSに載せる場合の候補。

```text
@opennextjs/aws 4.0.2
```

採用する場合は、Next.jsの対応バージョンとOpenNextの制約を実装前に再確認する。

## 7. Terraform / Neon

### Terraform CLI

最新stableとして以下を採用する。

```text
Terraform 1.15.7
```

`1.16.0-alpha...` はalphaのため採用しない。

### Neon Terraform Provider

```text
kislerdm/neon 0.13.0
```

### versions.tf例

```hcl
terraform {
  required_version = "= 1.15.7"

  required_providers {
    neon = {
      source  = "kislerdm/neon"
      version = "= 0.13.0"
    }
  }
}
```

## 8. Lock file方針

必ずcommitする。

```text
backend/go.sum
frontend/package-lock.json
infra/aws-cdk/package-lock.json
infra/neon-terraform/.terraform.lock.hcl
```

commitしない。

```text
terraform.tfstate
terraform.tfstate.*
.terraform/
node_modules/
```

## 9. 更新運用

認証基盤のため、自動マージはしない。

推奨:

- Renovate or DependabotでPR作成
- security updateは優先
- minor/patchはテスト通過後に手動マージ
- major updateは設計影響を確認
- Fosite / Echo / pgx / Next.js / React / Terraform provider は特に慎重に確認

## 10. 実装開始時の再確認

このファイルは2026-06-28時点の最新版を前提とする。

実装開始時に以下を再実行して確認する。

```bash
go list -m -versions github.com/labstack/echo/v4
go list -m -versions github.com/ory/fosite
go list -m -versions github.com/jackc/pgx/v5
npm view next version
npm view react version
npm view aws-cdk-lib version
terraform version
```
