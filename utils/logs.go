package utils

import (
	"github.com/Sirupsen/logrus"
	"os"
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetOutput(os.Stdout)
}
