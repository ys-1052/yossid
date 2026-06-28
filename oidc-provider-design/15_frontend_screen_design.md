# 15. Frontend Screen設計

## 1. 画面一覧

| 画面 | Path | MVP | 説明 |
|---|---|---|---|
| Landing | `/` | 任意 | YossIDの簡易説明 |
| Register | `/register` | 必須 | ユーザー登録 |
| Register Sent | `/register/sent` | 必須 | 確認メール送信完了 |
| Email Verify | `/email/verify` | 必須 | メール確認結果 |
| Login | `/login` | 必須 | email/password |
| Email OTP | `/mfa/email` | 必須 | メールOTP入力 |
| Logout | `/logout` | 必須 | ログアウト処理 |
| Error | `/error` | 必須 | 共通エラー |
| Account | `/account` | 将来 | アカウントトップ |
| Profile | `/account/profile` | 将来 | プロフィール |
| Security | `/account/security` | 将来 | セキュリティ設定 |

## 2. 共通レイアウト

### Auth Shell

中央寄せカードレイアウト。

```text
+--------------------------------+
| YossID                         |
| Secure identity for your apps. |
|                                |
| [ Card ]                       |
|                                |
| small footer                   |
+--------------------------------+
```

### Footer

MVP:

```text
YossID
```

将来:

- Terms
- Privacy
- Contact
- Security

## 3. Register画面

### 入力項目

必須:

- email
- password
- password confirmation
- family_name
- given_name
- family_name_kana
- given_name_kana
- gender
- birthdate
- country_code

### UI

セクション分けする。

```text
Account
- email
- password
- password confirmation

Profile
- family name
- given name
- family name kana
- given name kana
- gender
- birthdate
- country
```

### Validation

- email形式
- password policy
- password confirmation一致
- family_name必須
- given_name必須
- kana必須
- gender必須
- birthdate必須
- country_code必須

### 成功

`/register/sent` へ遷移。

## 4. Register Sent画面

表示:

```text
確認メールを送信しました。
メール内のリンクを開いて登録を完了してください。
```

再送ボタンはMVPでは任意。

## 5. Email Verify画面

### 成功

```text
メールアドレスを確認しました。
ログインできます。
```

CTA:

```text
ログインへ
```

### 失敗

```text
確認リンクが無効または期限切れです。
```

## 6. Login画面

### 入力項目

- email
- password

### 表示

```text
Sign in to YossID
```

### 成功

`/mfa/email` へ遷移。

### 失敗

汎用エラー:

```text
メールアドレスまたはパスワードが正しくありません。
```

## 7. Email OTP画面

### 入力項目

- 6桁OTP

### 表示

```text
メールに送信された6桁のコードを入力してください。
```

### 成功

return_toがあれば復帰。

例:

```text
/authorize?... に戻る
```

### 失敗

```text
コードが正しくないか、期限切れです。
```

## 8. Logout画面

POST logoutを実行し、完了表示。

```text
ログアウトしました。
```

## 9. Error画面

共通エラー表示。

表示項目:

- title
- message
- error_code optional
- return link

OAuth errorの詳細を出しすぎない。

## 10. Interaction

### Loading state

- button disabled
- spinner
- 二重送信防止

### Error state

- field error
- form-level error

### Success state

- clear next action

## 11. Design Tokens

### Color

MVPでは落ち着いた配色。

```text
background: near-white or near-black
surface: card background
primary: blue / indigo / emerald系
text: high contrast
muted: gray
danger: red
```

具体色は実装時にTailwind themeで定義する。

### Typography

- system font
- headingは太め
- bodyは読みやすさ優先

### Radius

- card: large
- input: medium
- button: medium

### Spacing

- generous spacing
- mobile first
- max-width 400〜480px

## 12. Responsive

最低対応:

- mobile width 360px
- desktop center card
- form fields full width

## 13. Accessibility

- すべてのinputにlabel
- errorにaria-describedby
- focus-visible
- 色だけでエラーを伝えない
- OTP入力もpaste対応
- semantic HTML

## 14. Security UX

- password policyを事前表示
- OTPの有効期限を表示
- login failureは詳細を出しすぎない
- register時に既存emailの有無を露出しすぎない
- email verification完了後に自動ログインしない
- 2FA完了後にsession発行

## 15. Copy案

### Product tagline

```text
Simple identity for personal apps.
```

または

```text
A small, secure identity provider by Yoss.
```

### Login

```text
Sign in to YossID
```

### Register

```text
Create your YossID
```

### OTP

```text
Check your email
```

## 16. 将来画面

### Account

- profile summary
- security status
- connected clients

### Profile

- family_name
- given_name
- kana
- gender
- birthdate
- country

### Security

- password change
- email OTP setting
- passkey management
- active sessions

### Connected Clients

- client name
- granted scopes
- last used
- revoke

### Withdrawal

- warning
- refresh token revoke
- status withdrawn
