package tor

import (
	"net"
	"net/http"
	"sync"
)

type TorHttpClient struct {
	Http *http.Client
}

type TorConfig struct {
	Host string
}

type TorConnection struct {
	*TorHttpClient
	mx sync.Mutex

	Hostname string
	Url      string
	IP       net.IP
	State    ConnectionState
}

type TorContainer struct {
	ID    string
	Name  string
	Image string

	Port string

	Connection *TorConnection

	Debug bool
}

type TorOptions struct {
	Name  string
	Image string

	Port string

	Override bool
	Debug    bool
}

type ConnectionState int

const (
	ConnectionCreated ConnectionState = iota
	ConnectionLocked
	ConnectionUnavailable
	ConnectionAvailable
)

func (s ConnectionState) String() string {
	return [...]string{"created", "locked", "unavailable", "available"}[s]
}
