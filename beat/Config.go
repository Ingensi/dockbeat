package beat

type DockerConfig struct {
	Period      *int64
	Socket      *string
	EnableTls   *bool
	TlsCaPath   *string
	TlsCertPath *string
	TlsKeyPath  *string
}

type ConfigSettings struct {
	Input DockerConfig
}
