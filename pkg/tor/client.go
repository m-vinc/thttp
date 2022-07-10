package tor

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"time"

	"golang.org/x/net/proxy"
)

func NewTorClient(c *TorConfig) (*TorHttpClient, error) {
	j, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	dialSocksProxy, err := proxy.SOCKS5("tcp", c.Host, nil, proxy.Direct)
	if err != nil {
		fmt.Println("Error connecting to proxy:", err)
	}

	tr := &http.Transport{Dial: dialSocksProxy.Dial}

	client := &http.Client{
		Transport: tr,
		Jar:       j,
		Timeout:   60 * time.Second,
	}

	return &TorHttpClient{
		Http: client,
	}, nil
}
