// package fromenv provides utilities to create lile options from environment
// variables. fromenv will error with fatal if it cannot resolve or errors
package fromenv

import (
	"os"

	"github.com/lileio/lile/pubsub"
	"github.com/lileio/lile/pubsub/google"
	"github.com/sirupsen/logrus"
)

func PubSubProvider(name string) pubsub.Provider {
	gpid := os.Getenv("GOOGLE_PUBSUB_PROJECT_ID")
	if gpid != "" {
		gc, err := google.NewGoogleCloud(gpid, name)
		if err != nil {
			logrus.Fatalf("fronenv: Google Cloud pubsub err: %s", err)
			return nil
		}

		logrus.Infof("Using Google Cloud pubsub: %s", gpid)
		return gc
	}

	logrus.Warn("Using noop pubsub provider")
	return pubsub.NoopProvider{}
}
