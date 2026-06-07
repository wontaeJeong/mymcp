package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"sync"
)

type JWK struct {
	Kid string `json:"kid,omitempty"`
	Alg string `json:"alg,omitempty"`
	Use string `json:"use,omitempty"`
	Kty string `json:"kty,omitempty"`
	N   string `json:"n,omitempty"`
	E   string `json:"e,omitempty"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWKResolver func(header map[string]any) (*rsa.PublicKey, error)

func importRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	if jwk.Kty != "" && jwk.Kty != "RSA" {
		return nil, errors.New("JWK is not an RSA key")
	}
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, err
	}
	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}
	if e == 0 {
		return nil, errors.New("invalid RSA public exponent")
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}, nil
}

func CreateStaticJwksResolver(jwks JWKS) JWKResolver {
	return func(header map[string]any) (*rsa.PublicKey, error) {
		kid, _ := header["kid"].(string)
		for _, candidate := range jwks.Keys {
			if kid == "" || candidate.Kid == kid {
				return importRSAPublicKey(candidate)
			}
		}
		return nil, errors.New("no matching JWK")
	}
}

func CreateRemoteJwksResolver(jwksURI string) JWKResolver {
	var (
		mu    sync.Mutex
		cache *JWKS
	)
	return func(header map[string]any) (*rsa.PublicKey, error) {
		mu.Lock()
		defer mu.Unlock()
		if cache == nil {
			res, err := http.Get(jwksURI)
			if err != nil {
				return nil, err
			}
			defer res.Body.Close()
			if res.StatusCode < 200 || res.StatusCode >= 300 {
				return nil, errors.New("JWKS fetch failed")
			}
			var parsed JWKS
			if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
				return nil, err
			}
			cache = &parsed
		}
		return CreateStaticJwksResolver(*cache)(header)
	}
}
