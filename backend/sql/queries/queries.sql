-- queries.sql

-- Users
-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserBySub :one
SELECT * FROM users WHERE sub = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (id, sub, email, email_verified_at, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateUserStatus :one
UPDATE users
SET status = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: WithdrawUser :exec
UPDATE users
SET status = 'withdrawn', updated_at = now()
WHERE id = $1;


-- User Profiles
-- name: GetUserProfile :one
SELECT * FROM user_profiles WHERE user_id = $1;

-- name: CreateUserProfile :one
INSERT INTO user_profiles (user_id, family_name, given_name, family_name_kana, given_name_kana, gender, birthdate, country_code)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateUserProfile :one
UPDATE user_profiles
SET family_name = $2, given_name = $3, family_name_kana = $4, given_name_kana = $5, gender = $6, birthdate = $7, country_code = $8, updated_at = now()
WHERE user_id = $1
RETURNING *;


-- User Password Credentials
-- name: GetPasswordCredential :one
SELECT * FROM user_password_credentials WHERE user_id = $1;

-- name: CreatePasswordCredential :one
INSERT INTO user_password_credentials (user_id, password_hash)
VALUES ($1, $2)
RETURNING *;

-- name: UpdatePasswordCredential :one
UPDATE user_password_credentials
SET password_hash = $2, password_changed_at = now()
WHERE user_id = $1
RETURNING *;


-- Pending User Registrations
-- name: GetPendingRegistrationByToken :one
SELECT * FROM pending_user_registrations WHERE verification_token_hash = $1;

-- name: CreatePendingRegistration :one
INSERT INTO pending_user_registrations (id, email, password_hash, family_name, given_name, family_name_kana, given_name_kana, gender, birthdate, country_code, verification_token_hash, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: MarkPendingRegistrationUsed :exec
UPDATE pending_user_registrations
SET used_at = now()
WHERE id = $1;


-- Clients
-- name: GetClientByID :one
SELECT * FROM clients WHERE id = $1;

-- name: GetClientByClientID :one
SELECT * FROM clients WHERE client_id = $1;

-- name: CreateClient :one
INSERT INTO clients (id, client_id, client_name, client_type, client_secret_hash, token_endpoint_auth_method, require_pkce, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;


-- Client Redirect URIs
-- name: GetClientRedirectURIs :many
SELECT * FROM client_redirect_uris WHERE client_id = $1;

-- name: CreateClientRedirectURI :one
INSERT INTO client_redirect_uris (id, client_id, redirect_uri)
VALUES ($1, $2, $3)
RETURNING *;


-- Client Allowed Scopes
-- name: GetClientAllowedScopes :many
SELECT scope FROM client_allowed_scopes WHERE client_id = $1;

-- name: CreateClientAllowedScope :exec
INSERT INTO client_allowed_scopes (client_id, scope)
VALUES ($1, $2);


-- Authorization Requests
-- name: GetAuthorizationRequest :one
SELECT * FROM authorization_requests WHERE request_id_hash = $1;

-- name: CreateAuthorizationRequest :one
INSERT INTO authorization_requests (id, request_id_hash, client_id, redirect_uri, scope, state, nonce, code_challenge, code_challenge_method, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: CompleteAuthorizationRequest :one
UPDATE authorization_requests
SET user_id = $2, completed_at = now()
WHERE id = $1
RETURNING *;


-- Authorization Codes
-- name: GetAuthorizationCode :one
SELECT * FROM authorization_codes WHERE code_hash = $1;

-- name: CreateAuthorizationCode :one
INSERT INTO authorization_codes (id, code_hash, client_id, user_id, redirect_uri, scope, nonce, code_challenge, code_challenge_method, auth_time, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: UseAuthorizationCode :one
UPDATE authorization_codes
SET used_at = now()
WHERE code_hash = $1 AND used_at IS NULL AND expires_at > now()
RETURNING *;


-- Refresh Tokens
-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens WHERE token_hash = $1;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (id, token_hash, token_family_id, client_id, user_id, scope, parent_id, rotated_to_id, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: RotateRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = now(), rotated_to_id = $2
WHERE id = $1;

-- name: RevokeRefreshTokenFamily :exec
UPDATE refresh_tokens
SET revoked_at = now()
WHERE token_family_id = $1;

-- name: MarkRefreshTokenReuse :exec
UPDATE refresh_tokens
SET reuse_detected_at = now()
WHERE id = $1;


-- Login Sessions
-- name: GetLoginSession :one
SELECT * FROM login_sessions WHERE session_hash = $1;

-- name: CreateLoginSession :one
INSERT INTO login_sessions (id, session_hash, user_id, auth_time, amr, ip_address, user_agent, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: RevokeLoginSession :exec
UPDATE login_sessions
SET revoked_at = now()
WHERE id = $1;


-- Email OTP Challenges
-- name: GetEmailOTPChallenge :one
SELECT * FROM email_otp_challenges WHERE challenge_id_hash = $1;

-- name: CreateEmailOTPChallenge :one
INSERT INTO email_otp_challenges (id, challenge_id_hash, user_id, otp_hash, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: IncrementOTPAttempts :one
UPDATE email_otp_challenges
SET attempts = attempts + 1
WHERE id = $1
RETURNING *;

-- name: MarkOTPUsed :exec
UPDATE email_otp_challenges
SET used_at = now()
WHERE id = $1;


-- Signing Keys
-- name: GetSigningKey :one
SELECT * FROM signing_keys WHERE kid = $1;

-- name: GetActiveSigningKeys :many
SELECT * FROM signing_keys WHERE status = 'active' OR status = 'standby' ORDER BY created_at DESC;

-- name: CreateSigningKey :one
INSERT INTO signing_keys (id, kid, alg, public_jwk, private_key_parameter_name, status, not_before, not_after)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;


-- User Withdrawals
-- name: CreateUserWithdrawal :one
INSERT INTO user_withdrawals (id, user_id, reason, ip_address, user_agent)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;


-- Audit Logs
-- name: CreateAuditLog :one
INSERT INTO audit_logs (id, event_type, user_id, client_id, result, ip_address, user_agent, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;
