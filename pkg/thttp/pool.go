package thttp

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/m-vinc/thttp/pkg/tor"
)

func (p *Pool) reload(container *tor.TorContainer) error {
	p.docker.ContainerKill(context.Background(), container.Connection.Hostname, "HUP")
	client, err := tor.NewTorClient(&tor.TorConfig{Host: container.Connection.Url})
	if err != nil {
		container.ChangeState(tor.ConnectionUnavailable, nil)
		return err
	}

	container.Connection.TorHttpClient = client

	return container.Healthcheck()
}

func (p *Pool) add(container *tor.TorContainer) error {
	p.mx.Lock()
	defer p.mx.Unlock()

	url := fmt.Sprintf("localhost:%s", container.Port)

	client, err := tor.NewTorClient(&tor.TorConfig{Host: url})
	if err != nil {
		if p.debug {
			log.Println("error while creating connection", err)
		}
		return err
	}

	conn := &tor.TorConnection{TorHttpClient: client, Url: url, Hostname: container.Name, State: tor.ConnectionCreated}
	container.Connection = conn

	err = container.Healthcheck()
	if err != nil {
		log.Println("cannot add connection to the pool: ", err)
		return err
	}

	log.Printf("connection added: %s -> %s", conn.Hostname, conn.IP)
	return nil
}

func (p *Pool) conn() (*tor.TorContainer, error) {
	p.mx.Lock()
	defer p.mx.Unlock()

	for {
		for _, c := range p.containers {
			if c.Connection.State == tor.ConnectionAvailable {
				return c, nil
			}
		}
		// log.Print("no connection available waiting 5 seconds")
		time.Sleep(time.Second * 1)
	}
}

func (p *Pool) DoTx(tx func(httpClient *http.Client) error) (err error) {
	container, err := p.conn()
	if err != nil {
		return err
	}

	container.ChangeState(tor.ConnectionLocked, nil)

	err = tx(container.Connection.TorHttpClient.Http)
	if err != nil {
		container.ChangeState(tor.ConnectionUnavailable, nil)
		return err
	}

	p.reload(container)

	container.ChangeState(tor.ConnectionAvailable, nil)

	return nil
}

func (p *Pool) Do(req *http.Request, reload bool) (resp *http.Response, err error) {
	container, err := p.conn()
	if err != nil {
		return nil, err
	}

	container.ChangeState(tor.ConnectionLocked, nil)
	res, err := container.Connection.TorHttpClient.Http.Do(req)
	if err != nil {
		container.ChangeState(tor.ConnectionUnavailable, nil)
		return nil, err
	}

	if reload {
		p.reload(container)
		return res, nil
	}

	container.ChangeState(tor.ConnectionAvailable, nil)
	return res, nil
}

func (p *Pool) monitor() {
	for {
		for _, c := range p.containers {
			if c.Connection.State != tor.ConnectionLocked {
				c.Healthcheck()
			}
		}
		time.Sleep(10 * time.Second)
	}
}

func (p *Pool) Close() error {
	errs := []error{}
	for _, container := range p.containers {
		_, err := tor.ContainerExist(context.Background(), p.docker, container.Name)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		err = p.docker.ContainerRemove(context.Background(), container.ID, types.ContainerRemoveOptions{
			Force: true,
		})
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return fmt.Errorf("errors occured while closing the pool - %+v", errs)
	}

	return nil
}

func NewPool(options *PoolOptions, docker *client.Client) (*Pool, error) {
	torContainers := []*tor.TorContainer{}
	for i := 0; i < options.Count; i++ {
		port := strconv.FormatInt(int64(options.PortStart+i), 10)
		if options.Debug {
			log.Printf("create container with port %s\n", port)
		}

		tcontainer, err := tor.NewTorContainer(context.Background(), docker, &tor.TorOptions{
			Name:     fmt.Sprintf("%s%d", options.Prefix, i),
			Image:    options.Image,
			Port:     port,
			Override: options.Override,
			Debug:    options.Debug,
		})
		if err != nil {
			log.Fatal(err)
		}
		torContainers = append(torContainers, tcontainer)
	}

	pool := &Pool{
		count:      options.Count,
		docker:     docker,
		containers: torContainers,
		debug:      options.Debug,
	}

	errs := []error{}

	for _, c := range pool.containers {
		err := pool.add(c)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == options.Count {
		return nil, fmt.Errorf("cannot initialize pool without one active connection")
	}

	go pool.monitor()

	return pool, nil
}
