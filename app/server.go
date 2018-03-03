package app

import (
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"etcd_backup/storage"
	"etcd_backup/utils"
)

const (
	defaultStorage = "local"
	cephStorage    = "ceph"
	s3Storage      = "s3"
)

type Options struct {
	etcdConfig
	cronConfig
	storage.StorageConfig
}

type etcdConfig struct {
	serverAddress string
	cacert        string
	cert          string
	key           string
}

type cronConfig struct {
	spec string
}

func NewOptions() *Options {
	etcdCfg := etcdConfig{
		serverAddress: "http://localhost:2379",
		cacert:        "/etc/etcd/ssl/ca.pem",
		cert:          "/etc/etcd/ssl/etcd.pem",
		key:           "/etc/etcd/ssl/etcd-key.pem",
	}

	cronCfg := cronConfig{
		spec: "0 0 23 * * *",
	}

	storageCfg := storage.StorageConfig{
		Mode: defaultStorage,
		Path: "/var/lib/etcd_backup",
	}

	return &Options{
		etcdConfig:    etcdCfg,
		cronConfig:    cronCfg,
		StorageConfig: storageCfg,
	}
}

func (o *Options) Validate(args []string) error {
	if len(args) != 0 {
		return errors.New("no arguments are supported")
	}

	return nil
}

func (o *Options) Run() error {
	errChan := make(chan error)
	defer close(errChan)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	go func(mux http.Handler) {
		if err := http.ListenAndServe(":9100", mux); err != nil {
			errChan <- err
		}
	}(mux)

	go func() {
		c := cron.New()
		spec := o.cronConfig.spec
		c.AddFunc(spec, func() {
			start := time.Now()
			logrus.Infof("start back up etcd data.\n")

			var err error
			switch o.StorageConfig.Mode {
			case defaultStorage:
				if err = utils.CreateDirIfNotExist(o.StorageConfig.Path); err != nil {
					errChan <- err
				}
			case cephStorage:
				//TODO
			case s3Storage:
				//TODO
			}
			command := fmt.Sprintf("ETCDCTL_API=3 etcdctl snapshot --endpoints=%s --cacert=%s --cert=%s --key=%s save %s/etcd_$(date \"+%%Y%%m%%d%%H%%M%%S\").db",
				o.etcdConfig.serverAddress, o.etcdConfig.cacert, o.etcdConfig.cert, o.etcdConfig.key, o.StorageConfig.Path)

			out, err := exec.Command("sh", "-c", command).Output()
			if err != nil {
				errChan <- err
			}

			logrus.Infof("etcd back up command: %v\n", command)
			logrus.Infof("Backup etcd data %s to be completed in %v seconds.\n", string(out), time.Now().Sub(start))
		})

		c.Start()
	}()

	err := <-errChan
	return err
}

func AddFlags(options *Options, fs *pflag.FlagSet) {
	fs.StringVar(&options.etcdConfig.serverAddress, "server-address", options.etcdConfig.serverAddress, "specify the etcd cluster address.")
	fs.StringVar(&options.etcdConfig.cacert, "cacert", options.etcdConfig.cacert, "verify certificates of TLS-enabled secure servers using this CA bundle")
	fs.StringVar(&options.etcdConfig.cert, "cert", options.etcdConfig.cert, "identify secure client using this TLS certificate file")
	fs.StringVar(&options.etcdConfig.key, "key", options.etcdConfig.key, "identify secure client using this TLS key file")
	fs.StringVar(&options.cronConfig.spec, "spec", options.cronConfig.spec, "a configuration that specifies shell commands to run periodically on a given schedule")
	fs.StringVar(&options.StorageConfig.Mode, "mode", "local", "storage mode, eg:local| ceph | s3")
	fs.StringVar(&options.StorageConfig.Path, "path", options.StorageConfig.Path, "storage path")
}

func NewEtcdBackUpCommand() *cobra.Command {
	opts := NewOptions()
	cmd := &cobra.Command{
		Use:  "etcd-backup",
		Long: "this is a etcd backup tool",
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			if err = opts.Validate(args); err != nil {
				panic(err)
			}

			if err = opts.Run(); err != nil {
				logrus.Fatalf("Failed to runing etcd backup server: %v", err)
			}
		},
	}

	flags := cmd.Flags()
	AddFlags(opts, flags)

	return cmd
}
