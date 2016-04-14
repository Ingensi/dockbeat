// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

type Config struct {
	Dockerbeat DockerbeatConfig
}

type DockerbeatConfig struct {
	Period string `yaml:"period"`
}
