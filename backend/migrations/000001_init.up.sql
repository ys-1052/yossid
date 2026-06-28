-- 000001_init.up.sql

-- 1. users table
CREATE TABLE users (
  id UUID PRIMARY KEY,
  sub TEXT UNIQUE NOT NULL,
  email TEXT UNIQUE NOT NULL,
  email_verified_at TIMESTAMPTZ NOT NULL,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 2. user_profiles table
CREATE TABLE user_profiles (
  user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  family_name TEXT,
  given_name TEXT,
  family_name_kana TEXT,
  given_name_kana TEXT,
  gender TEXT,
  birthdate DATE,
  country_code CHAR(2),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 3. user_password_credentials table
CREATE TABLE user_password_credentials (
  user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  password_hash TEXT NOT NULL,
  password_changed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 4. pending_user_registrations table
CREATE TABLE pending_user_registrations (
  id UUID PRIMARY KEY,
  email TEXT NOT NULL,
  password_hash TEXT NOT NULL,
  family_name TEXT NOT NULL,
  given_name TEXT NOT NULL,
  family_name_kana TEXT NOT NULL,
  given_name_kana TEXT NOT NULL,
  gender TEXT NOT NULL,
  birthdate DATE NOT NULL,
  country_code CHAR(2) NOT NULL,
  verification_token_hash TEXT UNIQUE NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_pending_user_registrations_email ON pending_user_registrations(email);
CREATE INDEX idx_pending_user_registrations_expires_at ON pending_user_registrations(expires_at);

-- 5. clients table
CREATE TABLE clients (
  id UUID PRIMARY KEY,
  client_id TEXT UNIQUE NOT NULL,
  client_name TEXT NOT NULL,
  client_type TEXT NOT NULL DEFAULT 'confidential',
  client_secret_hash TEXT,
  token_endpoint_auth_method TEXT NOT NULL DEFAULT 'client_secret_basic',
  require_pkce BOOLEAN NOT NULL DEFAULT true,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 6. client_redirect_uris table
CREATE TABLE client_redirect_uris (
  id UUID PRIMARY KEY,
  client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
  redirect_uri TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(client_id, redirect_uri)
);

-- 7. client_allowed_scopes table
CREATE TABLE client_allowed_scopes (
  client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
  scope TEXT NOT NULL,
  PRIMARY KEY(client_id, scope)
);

-- 8. authorization_requests table
CREATE TABLE authorization_requests (
  id UUID PRIMARY KEY,
  request_id_hash TEXT UNIQUE NOT NULL,
  client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
  redirect_uri TEXT NOT NULL,
  scope TEXT NOT NULL,
  state TEXT NOT NULL,
  nonce TEXT NOT NULL,
  code_challenge TEXT NOT NULL,
  code_challenge_method TEXT NOT NULL,
  requested_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at TIMESTAMPTZ NOT NULL,
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  completed_at TIMESTAMPTZ
);

-- 9. authorization_codes table
CREATE TABLE authorization_codes (
  id UUID PRIMARY KEY,
  code_hash TEXT UNIQUE NOT NULL,
  client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  redirect_uri TEXT NOT NULL,
  scope TEXT NOT NULL,
  nonce TEXT NOT NULL,
  code_challenge TEXT NOT NULL,
  code_challenge_method TEXT NOT NULL,
  auth_time TIMESTAMPTZ NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_authorization_codes_expires_at ON authorization_codes(expires_at);

-- 10. refresh_tokens table
CREATE TABLE refresh_tokens (
  id UUID PRIMARY KEY,
  token_hash TEXT UNIQUE NOT NULL,
  token_family_id UUID NOT NULL,
  client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  scope TEXT NOT NULL,
  parent_id UUID REFERENCES refresh_tokens(id) ON DELETE SET NULL,
  rotated_to_id UUID REFERENCES refresh_tokens(id) ON DELETE SET NULL,
  issued_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_used_at TIMESTAMPTZ,
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ,
  reuse_detected_at TIMESTAMPTZ
);

CREATE INDEX idx_refresh_tokens_family ON refresh_tokens(token_family_id);
CREATE INDEX idx_refresh_tokens_user_client ON refresh_tokens(user_id, client_id);

-- 11. login_sessions table
CREATE TABLE login_sessions (
  id UUID PRIMARY KEY,
  session_hash TEXT UNIQUE NOT NULL,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  auth_time TIMESTAMPTZ NOT NULL,
  amr TEXT NOT NULL,
  ip_address TEXT,
  user_agent TEXT,
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 12. email_otp_challenges table
CREATE TABLE email_otp_challenges (
  id UUID PRIMARY KEY,
  challenge_id_hash TEXT UNIQUE NOT NULL,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  otp_hash TEXT NOT NULL,
  attempts INT NOT NULL DEFAULT 0,
  max_attempts INT NOT NULL DEFAULT 5,
  expires_at TIMESTAMPTZ NOT NULL,
  used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 13. signing_keys table
CREATE TABLE signing_keys (
  id UUID PRIMARY KEY,
  kid TEXT UNIQUE NOT NULL,
  alg TEXT NOT NULL,
  public_jwk JSONB NOT NULL,
  private_key_parameter_name TEXT NOT NULL,
  status TEXT NOT NULL,
  not_before TIMESTAMPTZ NOT NULL,
  not_after TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 14. user_withdrawals table
CREATE TABLE user_withdrawals (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  reason TEXT,
  ip_address TEXT,
  user_agent TEXT,
  withdrawn_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 15. audit_logs table
CREATE TABLE audit_logs (
  id UUID PRIMARY KEY,
  event_type TEXT NOT NULL,
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  client_id UUID REFERENCES clients(id) ON DELETE SET NULL,
  result TEXT NOT NULL,
  ip_address TEXT,
  user_agent TEXT,
  metadata JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_logs_user_created ON audit_logs(user_id, created_at DESC);
CREATE INDEX idx_audit_logs_event_created ON audit_logs(event_type, created_at DESC);

-- Setup local test role if running locally, and grant permissions
DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'app_user') THEN
    CREATE ROLE app_user WITH LOGIN PASSWORD 'app_password';
  END IF;
END
$$;

GRANT USAGE ON SCHEMA public TO app_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_user;
