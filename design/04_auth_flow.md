# 04. Auth Flow設計

## 1. ユーザー登録フロー

### 概要

ユーザー登録では、登録情報を `pending_user_registrations` に一時保存し、メール確認完了後に `users` へ本登録する。

`users` にはメール確認済みユーザーのみ保存する。

### 入力項目

必須:

- email
- password
- family_name
- given_name
- family_name_kana
- given_name_kana
- gender
- birthdate
- country_code

### フロー

```text
1. GET /register
2. ユーザーが登録フォームを入力
3. POST /register
4. 入力バリデーション
5. password hash作成
6. email verification token生成
7. pending_user_registrations に保存
8. SESで確認メール送信
9. ユーザーがメール内リンクをクリック
10. GET /email/verify?token=...
11. token hash照合
12. pending_user_registrations が有効か確認
13. users / user_profiles / user_password_credentials に本登録
14. pending_user_registrations を used_at 更新
15. 登録完了画面表示
```

### セキュリティ要件

- passwordは平文保存しない
- verification tokenはDBにhash保存
- verification tokenは短命
- verification tokenは1回限り
- 同一emailのpendingがある場合は再送または上書き方針を決める
- 登録フォームにCSRF対策を入れる
- 登録回数にrate limitを入れる

## 2. ログインフロー

### 概要

ログインは email/password + メールOTP の2段階認証必須。

### フロー

```text
1. GET /login
2. email/password入力
3. POST /login
4. users.email でユーザー取得
5. users.status = active を確認
6. password hash検証
7. email OTP challenge作成
8. SESでOTP送信
9. OTP入力画面表示
10. POST /mfa/email/verify
11. OTP検証
12. login_sessions 作成
13. OP session cookie発行
14. return_to があれば元の /authorize に戻す
```

### OTP仕様

- 6桁数字
- 有効期限 5分
- 最大試行回数 5回
- DBにはOTP hashのみ保存
- 成功後は used_at を設定
- 連続送信制限を設ける

### ID Token反映

認証成功時、ID Tokenに以下を入れる。

```json
"amr": ["pwd", "email"]
```

`auth_time` はemail/password + OTPの認証完了時刻とする。

## 3. Authorization Code Flow

### 前提

Clientは confidential client のみ。

PKCE S256必須。

### フロー

```text
1. RP -> GET /authorize
2. request parameter検証
   - response_type=code
   - client_id
   - redirect_uri完全一致
   - scopeにopenidを含む
   - state必須
   - nonce必須
   - code_challenge必須
   - code_challenge_method=S256
3. OP session cookie確認
4. 未ログインならauthorization requestを保存して/loginへ
5. ログイン済みなら認可コード発行
6. redirect_uriへ code + state を付与してリダイレクト
```

### 同意

MVPでは first-party confidential client のみのため、consent画面は省略する。

ただし、将来外部公開に備え、内部処理は以下の構造にする。

```text
requested_scopes
allowed_client_scopes
granted_scopes
```

MVPでは `granted_scopes = requested_scopes` とする。

## 4. Token交換フロー

### authorization_code grant

```text
1. Client -> POST /token
2. client_secret_basic でClient認証
3. grant_type=authorization_code確認
4. code hashでauthorization_codesを取得
5. 未使用・未期限切れを確認
6. redirect_uri一致確認
7. PKCE code_verifier検証
8. authorization codeをused_at更新
9. ID Token / Access Token / Refresh Token発行
10. refresh_tokensにhash保存
11. response返却
```

### 認可コード一回限り利用

Postgresで以下をatomicに実行する。

```sql
UPDATE authorization_codes
SET used_at = now()
WHERE code_hash = $1
  AND used_at IS NULL
  AND expires_at > now()
RETURNING *;
```

RETURNINGが0件なら失敗。

## 5. Refresh Token Rotation

```text
1. Client -> POST /token grant_type=refresh_token
2. client認証
3. refresh_token hashでDB検索
4. active / not expired / not revoked を確認
5. 古いrefresh tokenをrevoked_at更新
6. 新しいrefresh tokenを発行
7. rotated_to_idを設定
8. 新しいID Token / Access Token / Refresh Token返却
```

### 再利用検知

すでにrevoked/rotated済みのrefresh tokenが利用された場合:

- reuse_detected_at を設定
- 同一familyのrefresh tokenを全失効
- audit_logsに記録
- token発行を拒否

## 6. UserInfoフロー

```text
1. RP -> GET /userinfo
2. Authorization: Bearer access_token
3. access token署名検証
4. exp / iss / aud / scope検証
5. subでユーザー取得
6. scopeに応じてclaim生成
7. JSON返却
```

## 7. Logoutフロー

MVPではOP session削除のみ対応する。

```text
1. POST /logout
2. OP session cookie確認
3. login_sessions.revoked_at を設定
4. Cookie削除
5. ログアウト完了
```

MVPではRP-Initiated Logoutは実装しない。

将来対応時に検討するもの:

- `end_session_endpoint`
- `id_token_hint`
- `post_logout_redirect_uri`
- RPごとの登録済みlogout URI
