package e2e

type RedisOptions struct {
	Host        bool
	Port        bool
	Auth        bool
	TLS         bool
	CA          bool
	Version     string
	ClusterMode bool
	Retry       int // number of extra attempts on failure (0 = no retry)
}
