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
	ProviderFirebase           = "firebase"
	authEnabledEnv             = "AUTH_ENABLED"
	firebaseProjectIDEnv       = "FIREBASE_PROJECT_ID"
	firebaseCredentialsFileEnv = "FIREBASE_CREDENTIALS_FILE"
	firebaseWebAPIKeyEnv       = "FIREBASE_WEB_API_KEY"
	firebaseWebAuthDomainEnv   = "FIREBASE_WEB_AUTH_DOMAIN"
	firebaseWebAppIDEnv        = "FIREBASE_WEB_APP_ID"
)

type FirebaseConfig struct {
	ProjectID       string
	CredentialsFile string
}

type FirebaseWebConfig struct {
	APIKey     string `json:"apiKey"`
	AuthDomain string `json:"authDomain"`
	ProjectID  string `json:"projectId"`
	AppID      string `json:"appId"`
}

type Identity struct {
	Provider      string
	Subject       string
	Email         string
	Name          string
	EmailVerified bool
}

type TokenVerifier interface {
	VerifyIDToken(ctx context.Context, token string) (*Identity, error)
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

func FirebaseWebConfigFromEnv() (FirebaseWebConfig, error) {
	config := FirebaseWebConfig{
		APIKey:     strings.TrimSpace(os.Getenv(firebaseWebAPIKeyEnv)),
		AuthDomain: strings.TrimSpace(os.Getenv(firebaseWebAuthDomainEnv)),
		ProjectID:  strings.TrimSpace(os.Getenv(firebaseProjectIDEnv)),
		AppID:      strings.TrimSpace(os.Getenv(firebaseWebAppIDEnv)),
	}

	required := []struct {
		name  string
		value string
	}{
		{firebaseWebAPIKeyEnv, config.APIKey},
		{firebaseWebAuthDomainEnv, config.AuthDomain},
		{firebaseProjectIDEnv, config.ProjectID},
		{firebaseWebAppIDEnv, config.AppID},
	}
	for _, field := range required {
		if field.value == "" {
			return FirebaseWebConfig{}, fmt.Errorf("%s is required", field.name)
		}
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

func (v *FirebaseVerifier) VerifyIDToken(ctx context.Context, token string) (*Identity, error) {
	verifiedToken, err := v.client.VerifyIDToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("firebase:verify id token %w", err)
	}

	return identityFromFirebaseToken(verifiedToken), nil
}

func identityFromFirebaseToken(token *firebaseauth.Token) *Identity {
	return &Identity{
		Provider:      ProviderFirebase,
		Subject:       token.UID,
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
