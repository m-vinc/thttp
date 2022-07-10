package thttp

import (
	"sync"

	"github.com/docker/docker/client"
	"github.com/m-vinc/thttp/pkg/tor"
)

type Pool struct {
	docker *client.Client

	count int
	mx    sync.Mutex

	containers []*tor.TorContainer

	debug bool
}

type PoolOptions struct {
	Prefix    string
	Image     string
	Count     int
	PortStart int
	Override  bool

	Debug bool
}

type PoolEvent int

func (pe PoolEvent) String() string {
	return [...]string{"healthcheck_failed", "healthcheck_alive"}[pe]
}

const (
	HealthcheckFailed PoolEvent = iota
	HealthcheckAlive
)

type Event struct {
	Type PoolEvent
}
