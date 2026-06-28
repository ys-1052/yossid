#!/usr/bin/env bash
set -euo pipefail

# Configuration
REGION="ap-northeast-1"
PREFIX="/yossid"

# Verify profile environment variable
if [ -z "${AWS_PROFILE:-}" ]; then
    echo "Warning: AWS_PROFILE environment variable is not set. Using default AWS profile."
else
    echo "Using AWS Profile: $AWS_PROFILE"
fi

echo "=========================================================="
echo " YossID OIDC Provider — SSM Parameter Setup"
echo "=========================================================="

# Check AWS CLI
if ! command -v aws &> /dev/null; then
    echo "Error: aws CLI is not installed" >&2
    exit 1
fi

# Check OpenSSL (for key generation)
if ! command -v openssl &> /dev/null; then
    echo "Error: openssl is not installed (required to generate JWT RSA private key)" >&2
    exit 1
fi

put_param() {
    local name="$1"
    local desc="$2"
    local val="$3"

    echo "Putting Parameter: $name..."
    aws ssm put-parameter \
        --region "$REGION" \
        --name "$name" \
        --description "$desc" \
        --type "SecureString" \
        --value "$val" \
        --overwrite >/dev/null
}

# 1. Generate Cookie Signing Key (64-byte random hex)
COOKIE_KEY=$(openssl rand -hex 64)
put_param "$PREFIX/cookie/signing-key" "YossID session cookie signing key" "$COOKIE_KEY"

# 2. Generate Token Pepper (32-byte random hex)
TOKEN_PEPPER=$(openssl rand -hex 32)
put_param "$PREFIX/token/pepper" "YossID token pepper for hashing values" "$TOKEN_PEPPER"

# 3. Generate OTP Pepper (32-byte random hex)
OTP_PEPPER=$(openssl rand -hex 32)
put_param "$PREFIX/otp/pepper" "YossID MFA OTP pepper for hashing values" "$OTP_PEPPER"

# 4. Generate RSA Private Key PEM (2048-bit)
echo "Generating RSA Private Key..."
RSA_PEM=$(openssl genrsa 2048 2>/dev/null)
put_param "$PREFIX/jwt/private-key" "YossID OIDC JWT signing RSA private key (RS256)" "$RSA_PEM"

echo "----------------------------------------------------------"
echo "Generated and uploaded core secrets successfully!"
echo "----------------------------------------------------------"
echo ""
echo "Now you need to set up the following parameters manually:"
echo ""
echo "1. Database URL (Neon pooled connection string):"
echo "   aws ssm put-parameter --region $REGION --name \"$PREFIX/database/url\" --type \"SecureString\" --value \"postgresql://app_user:<pass>@<endpoint>/yossid?sslmode=verify-full\" --overwrite"
echo ""
echo "2. SES From Email (validated email in Amazon SES ap-northeast-1):"
echo "   aws ssm put-parameter --region $REGION --name \"$PREFIX/ses/from-email\" --type \"SecureString\" --value \"no-reply@yourdomain.com\" --overwrite"
echo ""
echo "=========================================================="
