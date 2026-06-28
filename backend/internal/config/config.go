package config

import (
	"context"
	"crypto/rsa"
	"fmt"
	"os"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/ys-1052/yossid/backend/internal/security"
)

type Config struct {
	Env              string          // dev, prod
	Port             string          // e.g. "8080"
	Issuer           string          // OIDC issuer URL
	DatabaseURL      string          // Database connection string
	JWTPrivateKeyPEM string          // RSA private key PEM for token signing
	JWTPrivateKey    *rsa.PrivateKey // Parsed RSA private key
	CookieSigningKey []byte          // Key for secure cookies
	TokenPepper      string          // Pepper for token hashing
	OTPPepper        string          // Pepper for OTP hashing
	SesFromEmail     string          // SES sender email
	RunMode          string          // lambda, http
}

func LoadConfig(ctx context.Context) (*Config, error) {
	inLambda := os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != ""
	env := "local"
	if inLambda {
		env = "lambda"
	}
	port := getEnv("PORT", "8080")
	runMode := getEnv("RUN_MODE", "http")
	sesFromEmail := getEnv("SES_FROM_EMAIL", "no-reply@example.com")

	cfg := &Config{
		Env:          env,
		Port:         port,
		RunMode:      runMode,
		SesFromEmail: sesFromEmail,
	}

	if env == "local" {
		// Load from local environment variables
		cfg.Issuer = getEnv("ISSUER", "http://localhost:8080")
		cfg.DatabaseURL = getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5433/yossid?sslmode=disable")
		cfg.TokenPepper = getEnv("TOKEN_PEPPER", "local_token_pepper_value_32_chars_long")
		cfg.OTPPepper = getEnv("OTP_PEPPER", "local_otp_pepper_value_32_chars_long")

		cookieKeyStr := getEnv("COOKIE_SIGNING_KEY", "local_cookie_signing_key_32_bytes_long")
		cfg.CookieSigningKey = []byte(cookieKeyStr)

		// A dummy RSA private key for local development
		// Real environment should load from file or environment
		cfg.JWTPrivateKeyPEM = getEnv("JWT_PRIVATE_KEY_PEM", "")
		if cfg.JWTPrivateKeyPEM == "" {
			var err error
			cfg.JWTPrivateKeyPEM, err = security.GenerateRSAPrivateKeyPEM()
			if err != nil {
				return nil, fmt.Errorf("failed to generate local dev RSA private key: %w", err)
			}
		}

		// Parse RSA private key for dev path
		if cfg.JWTPrivateKeyPEM != "" {
			parsedKey, err := security.ParseRSAPrivateKeyPEM(cfg.JWTPrivateKeyPEM)
			if err != nil {
				return nil, fmt.Errorf("failed to parse local dev private key: %w", err)
			}
			cfg.JWTPrivateKey = parsedKey
		}

		return cfg, nil
	}

	// Load from SSM Parameter Store in production
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	ssmClient := ssm.NewFromConfig(awsCfg)

	// Read parameter names from environment variables
	dbParamName := os.Getenv("DATABASE_URL_PARAMETER")
	jwtParamName := os.Getenv("JWT_PRIVATE_KEY_PARAMETER")
	cookieParamName := os.Getenv("COOKIE_SIGNING_KEY_PARAMETER")
	tokenPepperParamName := os.Getenv("TOKEN_PEPPER_PARAMETER")
	otpPepperParamName := os.Getenv("OTP_PEPPER_PARAMETER")
	issuerParamName := os.Getenv("ISSUER_PARAMETER")
	sesFromEmailParamName := os.Getenv("SES_FROM_EMAIL_PARAMETER")

	if dbParamName == "" || jwtParamName == "" || cookieParamName == "" {
		return nil, fmt.Errorf("SSM Parameter name environment variables are not set")
	}

	cfg.DatabaseURL, err = getSSMParameter(ctx, ssmClient, dbParamName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch database URL from SSM: %w", err)
	}

	cfg.JWTPrivateKeyPEM, err = getSSMParameter(ctx, ssmClient, jwtParamName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWT private key from SSM: %w", err)
	}

	cookieKeyStr, err := getSSMParameter(ctx, ssmClient, cookieParamName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cookie signing key from SSM: %w", err)
	}
	cfg.CookieSigningKey = []byte(cookieKeyStr)

	cfg.TokenPepper, err = getSSMParameter(ctx, ssmClient, tokenPepperParamName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch token pepper from SSM: %w", err)
	}

	cfg.OTPPepper, err = getSSMParameter(ctx, ssmClient, otpPepperParamName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OTP pepper from SSM: %w", err)
	}

	if issuerParamName != "" {
		cfg.Issuer, err = getSSMParameter(ctx, ssmClient, issuerParamName)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch issuer from SSM: %w", err)
		}
	} else {
		cfg.Issuer = os.Getenv("ISSUER")
	}

	if cfg.Issuer == "" {
		return nil, fmt.Errorf("ISSUER is not set")
	}

	if sesFromEmailParamName != "" {
		cfg.SesFromEmail, err = getSSMParameter(ctx, ssmClient, sesFromEmailParamName)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch SES from email from SSM: %w", err)
		}
	} else {
		cfg.SesFromEmail = os.Getenv("SES_FROM_EMAIL")
	}
	if cfg.SesFromEmail == "" {
		cfg.SesFromEmail = "no-reply@example.com"
	}

	// Parse RSA private key for prod path
	if cfg.JWTPrivateKeyPEM != "" {
		cfg.JWTPrivateKey, err = security.ParseRSAPrivateKeyPEM(cfg.JWTPrivateKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to parse production private key: %w", err)
		}
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func getSSMParameter(ctx context.Context, client *ssm.Client, name string) (string, error) {
	out, err := client.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           &name,
		WithDecryption: boolPtr(true),
	})
	if err != nil {
		return "", err
	}
	if out.Parameter == nil || out.Parameter.Value == nil {
		return "", fmt.Errorf("parameter %s has nil value", name)
	}
	return *out.Parameter.Value, nil
}

func boolPtr(b bool) *bool {
	return &b
}
