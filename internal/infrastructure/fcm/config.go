// internal/infrastructure/fcm/config.go
package fcm

import (
	"context"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

type FCMConfig struct {
	CredentialsFile string `json:"credentials_file"`
}

func InitializeFCM(config *FCMConfig) (*FCMService, error) {
	// Load credentials from file
	opt := option.WithCredentialsFile(config.CredentialsFile)

	// Initialize Firebase app
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, err
	}

	// Create FCM service
	return NewFCMService(app)
}
