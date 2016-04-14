// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

type Config struct {
	Dockerbeat DockerbeatConfig
}

type TlsConfig struct {
	Enable   *bool   `config:"enable"`
	CaPath   *string `config:"ca_path"`
	CertPath *string `config:"cert_path"`
	KeyPath  *string `config:"key_path"`
}

type DockerbeatConfig struct {
	Period *int64
	Socket *string
	Tls    TlsConfig
}
