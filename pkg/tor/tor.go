package tor

import (
	"fmt"
	"io"
	"log"
	"net"
)

func (c *TorContainer) ChangeState(state ConnectionState, ip net.IP) {
	c.Connection.mx.Lock()

	if ip != nil {
		oldIP := c.Connection.IP
		if c.Connection.IP != nil && !c.Connection.IP.Equal(ip) {
			if c.Debug {
				log.Printf("%s ip changed: %s -> %s", c.Connection.Hostname, oldIP, ip)
			}
		}
		c.Connection.IP = ip
	}

	old := c.Connection.State
	if old != state {
		c.Connection.State = state
		if c.Debug {
			log.Printf("%s state changed: %s -> %s", c.Connection.Hostname, old, c.Connection.State)
		}
	}

	c.Connection.mx.Unlock()
}

func (c *TorContainer) Healthcheck() error {
	res, err := c.Connection.Http.Get("https://ifconfig.me")
	if err != nil {
		c.ChangeState(ConnectionUnavailable, nil)
		return fmt.Errorf("healthcheck failed: %s", err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	ip := net.ParseIP(string(body))
	if ip == nil {
		return fmt.Errorf("can't parse IP from ifconfig")
	}

	c.ChangeState(ConnectionAvailable, ip)
	return nil
}
