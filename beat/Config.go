package beat

type DockerConfig struct {
	Period *int64
	Socket *string
}

type ConfigSettings struct {
	Input DockerConfig
}
