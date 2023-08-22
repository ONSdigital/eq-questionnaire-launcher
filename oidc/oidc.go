package oidc

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/ONSdigital/eq-questionnaire-launcher/settings"
	"github.com/patrickmn/go-cache"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
	"google.golang.org/api/option"
)

func GenerateIdToken() (oauth2.TokenSource, error) {
	oidcBackend := settings.Get("OIDC_TOKEN_BACKEND")
	if oidcBackend == "gcp" {
		audience := settings.Get("SDS_OAUTH2_CLIENT_ID")
		return getGCPIdToken(audience)
	}
	return nil, nil
}

func cachedWithTTL(fn func(audience string) (oauth2.TokenSource, error)) func(audience string) (oauth2.TokenSource, error) {
	validitySeconds, _ := strconv.Atoi(settings.Get("OIDC_TOKEN_VALIDITY_IN_SECONDS"))
	leewaySeconds, _ := strconv.Atoi(settings.Get("OIDC_TOKEN_LEEWAY_IN_SECONDS"))

	ttl := validitySeconds - leewaySeconds
	// Create cache with default expiration of TTL seconds and cleanup interval of 1 minute
	ttlCache := cache.New(time.Duration(ttl)*time.Second, time.Minute)

	cachedFunc := func(audience string) (oauth2.TokenSource, error) {
		cachedSource, found := ttlCache.Get(audience)
		if found {
			log.Print("Found cached GCP ID token source")
			return cachedSource.(oauth2.TokenSource), nil
		}
		tokenSource, err := fn(audience)
		if err != nil {
			return nil, err
		}
		ttlCache.Set(audience, tokenSource, cache.DefaultExpiration)
		return tokenSource, nil
	}
	return cachedFunc
}

// uses the Google Cloud metadata server environment to create an identity token that can be added to a HTTP request
func getIdTokenFromMetadataServer(audience string) (oauth2.TokenSource, error) {
	ctx := context.Background()
	// Construct the GoogleCredentials object which obtains the default configuration from your working environment.
	credentials, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate default credentials: %w", err)
	}

	ts, err := idtoken.NewTokenSource(ctx, audience, option.WithCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("failed to create NewTokenSource: %w", err)
	}

	// Generate the ID token.
	_, err = ts.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to receive token: %w", err)
	}
	log.Print("Succesfully generated GCP ID token")

	return ts, nil
}

var getGCPIdToken = cachedWithTTL(getIdTokenFromMetadataServer)
