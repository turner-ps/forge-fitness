// Package auth
package auth

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	firebase "firebase.google.com/go/v4"
	firebaseauth "firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

const (
	authEnabledEnv             = "AUTH_ENABLED"
	firebaseProjectIDEnv       = "FIREBASE_PROJECT_ID"
	firebaseCredentialsFileEnv = "FIREBASE_CREDENTIALS_FILE"
)

type FirebaseConfig struct {
	ProjectID       string
	CredentialsFile string
}

type VerifiedToken struct {
	UID           string
	Email         string
	Name          string
	EmailVerified bool
}

type TokenVerifier interface {
	VerifyIDToken(ctx context.Context, token string) (*VerifiedToken, error)
}

type FirebaseVerifier struct {
	client *firebaseauth.Client
}

func EnabledFromEnv() (bool, error) {
	value := strings.TrimSpace(os.Getenv(authEnabledEnv))
	if value == "" {
		return false, nil
	}

	enabled, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean", authEnabledEnv)
	}

	return enabled, nil
}

func FirebaseConfigFromEnv() (FirebaseConfig, error) {
	config := FirebaseConfig{
		ProjectID:       os.Getenv(firebaseProjectIDEnv),
		CredentialsFile: os.Getenv(firebaseCredentialsFileEnv),
	}

	if config.ProjectID == "" {
		return FirebaseConfig{}, fmt.Errorf("%s is required", firebaseProjectIDEnv)
	}

	if config.CredentialsFile == "" {
		return FirebaseConfig{}, fmt.Errorf("%s is required", firebaseCredentialsFileEnv)
	}

	return config, nil
}

func NewFirebaseVerifierFromEnv(ctx context.Context) (*FirebaseVerifier, error) {
	config, err := FirebaseConfigFromEnv()
	if err != nil {
		return nil, err
	}

	return NewFirebaseVerifier(ctx, config)
}

func NewFirebaseVerifier(ctx context.Context, config FirebaseConfig) (*FirebaseVerifier, error) {
	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: config.ProjectID,
	}, option.WithAuthCredentialsFile(option.ServiceAccount, config.CredentialsFile))
	if err != nil {
		return nil, fmt.Errorf("firebase:new app %w", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebase:auth client %w", err)
	}

	return &FirebaseVerifier{client: client}, nil
}

func (v *FirebaseVerifier) VerifyIDToken(ctx context.Context, token string) (*VerifiedToken, error) {
	verifiedToken, err := v.client.VerifyIDToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("firebase:verify id token %w", err)
	}

	return verifiedTokenFromFirebase(verifiedToken), nil
}

func verifiedTokenFromFirebase(token *firebaseauth.Token) *VerifiedToken {
	return &VerifiedToken{
		UID:           token.UID,
		Email:         stringClaim(token.Claims, "email"),
		Name:          stringClaim(token.Claims, "name"),
		EmailVerified: boolClaim(token.Claims, "email_verified"),
	}
}

func stringClaim(claims map[string]any, key string) string {
	value, ok := claims[key].(string)
	if !ok {
		return ""
	}

	return value
}

func boolClaim(claims map[string]any, key string) bool {
	value, ok := claims[key].(bool)
	if !ok {
		return false
	}

	return value
}
