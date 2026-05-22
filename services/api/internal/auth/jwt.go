package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	priv      *rsa.PrivateKey
	pub       *rsa.PublicKey
	accessTTL time.Duration
}

// LoadJWTManager đọc khoá RSA từ PEM. Nếu path rỗng (dev mode) → sinh ephemeral 2048-bit.
func LoadJWTManager(privPath, pubPath string, accessTTL time.Duration) (*JWTManager, error) {
	if accessTTL == 0 {
		accessTTL = 15 * time.Minute
	}
	if privPath == "" {
		priv, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, fmt.Errorf("ephemeral key gen: %w", err)
		}
		return &JWTManager{priv: priv, pub: &priv.PublicKey, accessTTL: accessTTL}, nil
	}
	privBytes, err := os.ReadFile(privPath)
	if err != nil {
		return nil, fmt.Errorf("read priv: %w", err)
	}
	pubBytes, err := os.ReadFile(pubPath)
	if err != nil {
		return nil, fmt.Errorf("read pub: %w", err)
	}
	pBlock, _ := pem.Decode(privBytes)
	if pBlock == nil {
		return nil, errors.New("invalid priv PEM")
	}
	privKey, err := x509.ParsePKCS8PrivateKey(pBlock.Bytes)
	if err != nil {
		k2, err2 := x509.ParsePKCS1PrivateKey(pBlock.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("parse priv: %w / %v", err, err2)
		}
		privKey = k2
	}
	rsaPriv, ok := privKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("priv key not RSA")
	}
	pubBlock, _ := pem.Decode(pubBytes)
	if pubBlock == nil {
		return nil, errors.New("invalid pub PEM")
	}
	pubAny, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse pub: %w", err)
	}
	rsaPub, ok := pubAny.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("pub key not RSA")
	}
	return &JWTManager{priv: rsaPriv, pub: rsaPub, accessTTL: accessTTL}, nil
}

type Claims struct {
	UserID  string `json:"sub"`
	Version int    `json:"ver"`
	jwt.RegisteredClaims
}

func (m *JWTManager) Issue(userID string, ver int) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:  userID,
		Version: ver,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
			Issuer:    "buocvang",
			Subject:   userID,
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return tok.SignedString(m.priv)
}

func (m *JWTManager) Parse(tokenStr string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected alg: %v", t.Header["alg"])
		}
		return m.pub, nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return c, nil
}

func (m *JWTManager) AccessTTLSeconds() int { return int(m.accessTTL.Seconds()) }

// NewRefreshToken sinh 256-bit ngẫu nhiên, trả token raw + hash để lưu DB.
func NewRefreshToken() (raw, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return
	}
	raw = base64.RawURLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(raw))
	hash = hex.EncodeToString(sum[:])
	return
}

func HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
