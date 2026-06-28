# 13. Open Decisions / 将来検討

## 1. Issuer / 独自ドメイン

MVPではAPI Gateway URLをissuerとする。

将来外部公開時は独自ドメインを取得するか再検討する。

懸念:

- issuer変更はClientへの影響が大きい
- 外部公開前に独自ドメインへ移行した方がよい

## 2. Password Hash

候補:

- Argon2id
- bcrypt

推奨:

```text
Argon2id
```

ただし、LambdaでのCPUコスト・実装容易性も考慮する。

## 3. JWT署名アルゴリズム

MVP:

```text
RS256
```

将来:

```text
ES256
```

## 4. Access Token形式

MVP:

```text
JWT Access Token
```

将来:

- opaque token
- introspection endpoint
- jti blacklist

## 5. Consent画面

MVPではfirst-party clientのみのため省略。

外部公開前には必須。

## 6. Client管理

MVPではseed管理。

将来:

- 管理画面
- client secret rotation
- client status management
- third-party onboarding

## 7. パスワードリセット

MVPでは実装しない。

外部公開前には必要。

## 8. Passkey

後回し。

将来検討:

- WebAuthn
- usernameless対応
- amrへの反映

## 9. FAPI / PAR / RAR

MVPでは実装しない。

金融APIや高セキュリティRPが必要になった場合に再検討。

## 10. DBネットワーク

MVP:

```text
Lambda -> Neon public endpoint + TLS verify-full
```

将来:

- Neon IP Allow
- Neon Private Networking
- AWS RDS / Aurora
- VPC内閉域化

## 11. Rate Limit

MVPでは簡易実装。

将来:

- WAF
- DynamoDB-based rate limit
- API Gateway usage plan
- Bot対策

## 12. SES送信元

ドメインを取得しないため、送信元メールアドレスをどうするか決める必要がある。

候補:

- 既存ドメインのメールアドレス
- SES verified email address
- 将来ドメイン取得

## 13. Neon IaC

MVP推奨:

```text
Terraform / OpenTofu
```

未決:

- Terraform state管理場所
- Neon provider version
- dev/prod分離方法

## 14. OpenID Conformance Suite

MVPでは必須にしない。

外部公開前に実施する。
