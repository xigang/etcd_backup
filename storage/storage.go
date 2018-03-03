package storage

type storage interface {
}

type StorageConfig struct {
	Mode string
	Path string
	Local
	Ceph
	S3
}
