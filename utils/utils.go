package utils

import (
	goflag "flag"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/pflag"
)

func WordSepNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	if strings.Contains(name, "_") {
		return pflag.NormalizedName(strings.Replace(name, "_", "-", -1))
	}
	return pflag.NormalizedName(name)
}

func InitFlags() {
	pflag.CommandLine.SetNormalizeFunc(WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	pflag.Parse()

	pflag.VisitAll(func(flag *pflag.Flag) {
		logrus.Infof("FLAG: --%s=%q", flag.Name, flag.Value)
	})
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func CreateDirIfNotExist(path string) error {
	var exist bool
	var err error
	exist, err = exists(path)
	if err != nil {
		return err
	}

	if !exist {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}

	return err
}
