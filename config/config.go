// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

type Config struct {
	Dockbeat DockbeatConfig
}

type TlsConfig struct {
	Enable   *bool   `config:"enable"`
	CaPath   *string `config:"ca_path"`
	CertPath *string `config:"cert_path"`
	KeyPath  *string `config:"key_path"`
}

type StatsConfig struct {
	Container *bool `config:"container"`
	Net       *bool `config:"net"`
	Memory    *bool `config:"memory"`
	Blkio     *bool `config:"blkio"`
	Cpu       *bool `config:"cpu"`
}

type DockbeatConfig struct {
	Period *int64      `config:"period"`
	Socket *string     `config:"socket"`
	Tls    TlsConfig   `config:"tls"`
	Stats  StatsConfig `config:"stats"`
}
