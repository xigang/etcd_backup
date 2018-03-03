package main

import (
	goflag "flag"
	"os"

	"github.com/spf13/pflag"

	"etcd_backup/app"
	"etcd_backup/utils"
)

func main() {
	command := app.NewEtcdBackUpCommand()

	pflag.CommandLine.SetNormalizeFunc(utils.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
