package security

import (
	"context"
	"crypto/rsa"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/token/jwt"
)

// NewOAuth2Provider initializes Ory Fosite Provider with the custom postgres store and RS256 keys.
func NewOAuth2Provider(tokenPepper string, jwtPrivateKey *rsa.PrivateKey, store fosite.Storage) fosite.OAuth2Provider {
	// Configure token lifespans as specified in 07_token_key_design.md
	fositeConfig := &fosite.Config{
		AccessTokenLifespan:   15 * time.Minute,
		RefreshTokenLifespan:  30 * 24 * time.Hour,
		AuthorizeCodeLifespan: 5 * time.Minute,
		IDTokenLifespan:       15 * time.Minute,
		GlobalSecret:          []byte(tokenPepper),
	}

	// Cast JWTPrivateKey if present, or generate a dummy one (safeguard)
	var privateKey *rsa.PrivateKey
	if jwtPrivateKey != nil {
		privateKey = jwtPrivateKey
	} else {
		// Fallback for security
		var err error
		privateKey, err = GenerateRSAPrivateKey()
		if err != nil {
			panic("failed to generate fallback RSA key: " + err.Error())
		}
	}

	keyGetter := func(ctx context.Context) (interface{}, error) {
		return privateKey, nil
	}

	strategy := &compose.CommonStrategy{
		CoreStrategy:               compose.NewOAuth2HMACStrategy(fositeConfig),
		OpenIDConnectTokenStrategy: compose.NewOpenIDConnectStrategy(keyGetter, fositeConfig),
		Signer:                     &jwt.DefaultSigner{GetPrivateKey: keyGetter},
	}

	// Compose Ory Fosite provider with explicit authorize code, refresh token, PKCE and OIDC core handlers
	provider := compose.Compose(
		fositeConfig,
		store,
		strategy,
		compose.OAuth2AuthorizeExplicitFactory,
		compose.OAuth2RefreshTokenGrantFactory,
		compose.OpenIDConnectExplicitFactory,
		compose.OAuth2PKCEFactory,
	)

	return provider
}
