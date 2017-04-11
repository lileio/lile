package lile

import (
	"os"

	"github.com/lileio/lile/pubsub"
	"github.com/sirupsen/logrus"
)

// PubSubProviderFromEnv initializes a pubsub provider from environment variables if present
func PubSubProviderFromEnv(opts options) pubsub.Provider {
	gpid := os.Getenv("GOOGLE_PUBSUB_PROJECT_ID")
	if gpid != "" {
		gc, err := pubsub.NewGoogleCloud(gpid, opts.name)
		if err != nil {
			logrus.Fatalf("Can't create google cloud pubsub: %s", err)
			return nil
		}

		logrus.Infof("Google PubSub enabled for pubsub: %s", gpid)
		return gc
	}

	return nil
}
