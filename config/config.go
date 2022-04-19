package config

type Config struct {
	LocalAddr  string
	ProxyAddr  string
	ServerAddr string
	ExposeAddr string
	Key        string
	ServerMode bool
	Timeout    int
}
