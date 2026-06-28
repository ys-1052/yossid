# 10. Neon管理・接続設計

## 1. 方針

Neon Postgresを利用する。

MVPでは無料枠を利用し、public endpoint + TLSで接続する。

Neonリソースもコード管理する。

## 2. コード管理方式

確定:

```text
Terraform
```

理由:

- Neon providerでproject / branch / database / role等を管理しやすい
- AWS CDKと責務分離できる
- CDK Custom Resourceよりシンプル

## 3. 管理対象

Neon側でコード管理するもの:

- project
- branch
- database
- role
- compute endpoint
- connection pooling設定

DB schemaはNeon providerではなくmigration toolで管理する。

## 4. Migration

候補:

- golang-migrate
- goose
- Atlas
- Flyway

MVP推奨:

```text
golang-migrate or goose
```

migration管理対象:

- tables
- indexes
- constraints
- enum相当
- grants

## 5. DBユーザー分離

### migration_user

用途:

- schema作成
- migration実行
- DDL

利用場所:

- local
- CI

Lambdaには渡さない。

### app_user

用途:

- runtime application

権限:

- SELECT
- INSERT
- UPDATE
- DELETE

DDL権限は付与しない。

## 6. 接続方式

LambdaからNeonへ接続する。

```text
AWS Lambda
  -> public internet
  -> Neon pooled endpoint
  -> TLS verify-full
```

接続文字列:

```text
sslmode=verify-full
```

## 7. Connection Pooling

Lambdaは同時起動により接続数が増えやすい。

Neonのpooled endpointを使う。

アプリ側でも以下を制限する。

```text
MaxOpenConns
MaxIdleConns
ConnMaxLifetime
```

MVP例:

```text
MaxOpenConns = 5
MaxIdleConns = 1
ConnMaxLifetime = 5分
```

Lambda instanceごとにconnection poolが作られる点に注意する。

## 8. SSM連携

Neon接続文字列はSSM SecureStringへ保存する。

```text
/oidc-provider/prod/database/url
```

Lambda起動時に取得する。

## 9. セキュリティ

無料枠ではPrivateLink / IP Allowlistは前提にしない。

代わりに以下を必須にする。

- TLS verify-full
- app_user最小権限
- migration_user分離
- SQL injection対策
- secrets管理
- query loggingに機微情報を出さない

## 10. 将来移行

外部公開・重要度上昇時の選択肢:

- Neon paid plan + IP Allow
- Neon Private Networking
- AWS RDS / Auroraへ移行
- RDS Proxy利用
- VPC内閉域構成

## 11. Backup

MVPではNeonの標準機能に依存する。

外部公開時は以下を検討。

- 定期dump
- 別branchへのバックアップ
- point-in-time restore
- migration rollback計画

## 12. 環境分離

最低限:

```text
dev
prod
```

Neon branchを使う場合:

```text
main/prod branch
dev branch
```

issuerやclient設定も環境ごとに分ける。

## 13. Terraform構成

NeonはTerraformで管理する。

推奨ディレクトリ:

```text
infra/neon-terraform/
  main.tf
  variables.tf
  outputs.tf
  versions.tf
  envs/
    dev.tfvars
    prod.tfvars
```

管理対象:

- Neon project
- Neon branch
- Neon database
- Neon role
- Neon compute endpoint
- pooled connection endpoint

管理しないもの:

- DB table schema
- migration SQL
- application seed data

これらはmigration toolとseed commandで管理する。

## 14. Terraform State

MVPではlocal stateでも開始可能。

ただし、GitHubにstateをcommitしない。

将来候補:

- Terraform Cloud
- S3 backend
- GitHub Actions OIDC + S3 backend

MVPでは以下を `.gitignore` に入れる。

```text
*.tfstate
*.tfstate.*
.terraform/
.terraform.lock.hcl はcommitしてよい
```

## 15. Terraform Secrets

Neon API Keyは環境変数で渡す。

```bash
export NEON_API_KEY=...
```

tfvarsやGitHubにAPI Keyを置かない。


## 16. Version Pinning

Terraform / Neon providerは以下を基準バージョンとして開始する。

```text
Terraform 1.15.7
kislerdm/neon 0.13.0
```

`versions.tf`:

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

`.terraform.lock.hcl` はcommitする。

`terraform.tfstate` はcommitしない。
