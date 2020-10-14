package runtime

import "github.com/sirupsen/logrus"

var (
	log    *logrus.Logger
	logger *logrus.Entry
)

func NewLogger(verbose bool) *logrus.Entry {
	if log == nil {
		log := logrus.New()
		if verbose {
			log.SetLevel(logrus.DebugLevel)
		}

		logger = logrus.NewEntry(log)
	}

	if logger == nil {
		logger = logrus.NewEntry(log)
	}

	return logger
}

func Logger() *logrus.Entry {
	if logger == nil {
		NewLogger(true)
	}

	return logger
}
