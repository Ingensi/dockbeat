package beat

type TlsConfig struct {
	Enable   *bool   `config:"enable"`
	CaPath   *string `config:"ca_path"`
	CertPath *string `config:"cert_path"`
	KeyPath  *string `config:"key_path"`
}

type DockerConfig struct {
	Period *int64
	Socket *string
	Tls    TlsConfig
}

type ConfigSettings struct {
	Input DockerConfig
}
