# 14. Frontend Next.js設計

## 1. 目的

認証UIをNext.jsで作り込み、シンプルかつモダンな体験にする。

OIDC/OAuthのプロトコル処理はGo Lambda + Fositeに残し、Next.jsは以下を担当する。

- 画面表示
- フォームUI
- クライアントサイド補助バリデーション
- エラー表示
- 認証フロー用ページ
- アカウント関連ページの将来拡張

## 2. 基本方針

```text
Next.js = 認証UI
Go Lambda = OIDC/OAuth protocol backend
Neon = 永続化
CloudFront = 同一origin front door
```

Next.jsからDBへ直接接続しない。

Next.jsはToken発行・Authorization Code発行・Refresh Token Rotationを実装しない。

## 3. 採用技術

| 項目 | 採用 |
|---|---|
| Framework | Next.js |
| Router | App Router |
| Language | TypeScript |
| Styling | Tailwind CSS |
| Component | 自前UIコンポーネント |
| Form | React Hook Form候補 |
| Validation | Zod候補 |
| Icons | lucide-react候補 |
| Theme | dark/light対応は将来。MVPはlight/darkどちらか固定でも可 |

## 4. デザインコンセプト

```text
modern
minimal
secure
calm
developer-friendly
```

方向性:

- 余白多め
- 入力欄は大きく読みやすく
- エラーは怖すぎないが明確に
- セキュリティ系サービスらしい落ち着いた配色
- 個人プロジェクト感として "YossID" ブランドを出す

## 5. Branding

Product name:

```text
YossID
```

読み:

```text
ヨッシド
```

意味:

```text
Yoss + ID
Yoss Identity
```

ロゴはMVPではテキストロゴでよい。

例:

```text
YossID
```

## 6. Frontend Routing

```text
/
  Landing / status page

/register
  ユーザー登録

/register/sent
  確認メール送信完了

/email/verify
  メール確認結果

/login
  email/password login

/mfa/email
  メールOTP入力

/logout
  ログアウト

/error
  共通エラー

/account
  将来用アカウントトップ

/account/profile
  将来用プロフィール表示・編集

/account/security
  将来用セキュリティ設定
```

MVP必須:

```text
/register
/register/sent
/email/verify
/login
/mfa/email
/logout
/error
```

## 7. Backend API連携

Next.js画面は同一originのGo APIへPOSTする。

例:

```text
POST /api/register
POST /api/login
POST /api/mfa/email/verify
POST /api/logout
```

または、既存のGo endpointへ直接POSTする。

```text
POST /register
POST /login
POST /mfa/email/verify
POST /logout
```

MVPではGo側endpointに直接POSTでもよい。

ただし、Next.jsのRoute HandlerをBFFとして挟む場合は、責務を以下に限定する。

- CSRF token取り回し
- form data整形
- Go APIへのproxy
- UI向けエラー整形

Token発行・認可コード発行は絶対にNext.js側に置かない。

## 8. Same Origin設計

認証UIとOIDC APIを別originにすると、Cookie / SameSite / redirectが複雑になる。

そのため、CloudFrontで同一origin化する。

```text
https://dxxxxxxxxxxxxx.cloudfront.net/login
https://dxxxxxxxxxxxxx.cloudfront.net/authorize
https://dxxxxxxxxxxxxx.cloudfront.net/token
```

CookieはCloudFrontのhostに対してhost-only cookieとして発行する。

## 9. Cookie設計

Cookie発行は基本的にGo backendが行う。

Next.jsはCookieの中身を読まない。

Cookie例:

```text
op_session
csrf_token
auth_request
mfa_challenge
```

Cookie属性:

```text
HttpOnly
Secure
SameSite=Lax
Path=/
```

CSRF tokenをhidden inputに埋め込む必要がある画面では、Go backendまたはNext.js BFFがCSRF tokenを発行する。

## 10. CSRF設計

HTML formのPOSTにはCSRF対策を入れる。

対象:

- register
- login
- mfa verify
- logout

方式候補:

1. Synchronizer Token
2. Double Submit Cookie

MVP推奨:

```text
Synchronizer Token
```

ただしNext.jsとGo backendを分離するため、CSRF token発行・検証の責務を明確化する。

推奨:

- Go backendがCSRF tokenを発行・検証
- Next.jsは初期表示時にGoからCSRF tokenを取得してformに埋め込む

実装が重い場合、MVPではBFF Route HandlerでCSRFを扱ってからGoへproxyする。

## 11. Form Validation

クライアントサイド:

- 入力漏れ
- email形式
- password confirmation一致
- birthdate形式
- country_code選択

サーバーサイド:

- すべて再検証
- クライアントバリデーションを信用しない

## 12. Error Handling

エラー表示は以下の方針。

- 入力エラーはfield単位で表示
- 認証失敗は汎用メッセージにする
- email存在有無を露出しすぎない
- OAuth errorは仕様形式に従う
- UI上は分かりやすい文言に変換する

例:

```text
メールアドレスまたはパスワードが正しくありません。
```

## 13. Accessibility

最低限対応:

- label / input関連付け
- keyboard navigation
- focus style
- color contrast
- error messageとaria-describedby
- button loading state
- screen reader向けstatus message

## 14. Page Rendering方針

MVPではServer Components中心。

インタラクションが必要なフォーム部分だけClient Componentsにする。

理由:

- JS bundleを小さくする
- セキュリティ系画面のロジックをシンプルにする
- 初期表示を安定させる

## 15. Directory Structure

```text
frontend/
  app/
    layout.tsx
    page.tsx
    register/
      page.tsx
      sent/
        page.tsx
    email/
      verify/
        page.tsx
    login/
      page.tsx
    mfa/
      email/
        page.tsx
    logout/
      page.tsx
    error/
      page.tsx

  components/
    ui/
      button.tsx
      input.tsx
      card.tsx
      alert.tsx
      field-error.tsx
      spinner.tsx
    layout/
      auth-shell.tsx
      brand.tsx

  features/
    auth/
      register-form.tsx
      login-form.tsx
      email-otp-form.tsx
      schemas.ts
      actions.ts

  lib/
    api.ts
    csrf.ts
    errors.ts
    constants.ts

  styles/
    globals.css
```

## 16. UI Component方針

外部UIライブラリはMVPでは必須にしない。

候補:

- Tailwind CSSのみ
- shadcn/ui風の自前コンポーネント
- Headless UI系

MVP推奨:

```text
Tailwind CSS + 自前最小コンポーネント
```

理由:

- 依存を減らす
- 認証画面に必要なUIは少ない
- デザインの一貫性を作りやすい

## 17. Security Boundary

Next.js frontendでやってよいこと:

- UI表示
- 入力補助
- エラー表示
- Go APIへのPOST
- CSRF tokenの受け渡し

Next.js frontendでやらないこと:

- password hash
- token発行
- authorization code発行
- refresh token rotation
- client secret検証
- JWT署名
- DB直接接続

## 18. Deploy方針

AWS上でNext.jsを動かす候補:

1. Amplify Hosting
2. OpenNext on AWS
3. Lambda Web Adapter + Next.js standalone

今回の優先順位:

```text
1. CDKで統一したい -> OpenNext / Lambda Web Adapter系
2. 手軽さ優先 -> Amplify Hosting
```

ただし、issuerとCookieを安定させるため、CloudFront front doorの下に置く。

## 19. MVPで作る画面

- Landing
- Register
- Register Sent
- Email Verify Result
- Login
- Email OTP
- Logout
- Error

## 20. 将来作る画面

- Account top
- Profile edit
- Security settings
- Connected clients
- Consent history
- Active sessions
- Withdrawal
- Password reset
- Passkey management


## 21. Version Pinning

Frontendは以下を基準バージョンとして開始する。

```text
next 16.2.9
react 19.2.7
react-dom 19.2.7
typescript 6.0.3
tailwindcss 4.3.1
@tailwindcss/postcss 4.3.1
react-hook-form 7.80.0
@hookform/resolvers 5.2.2
zod 4.4.3
lucide-react 1.21.0
@types/react 19.2.17
```

`package-lock.json` は必ずcommitする。

バージョン範囲指定ではなく、実装開始時点では固定バージョンで開始する。
