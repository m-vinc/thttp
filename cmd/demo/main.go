package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/docker/docker/client"
	"github.com/m-vinc/thttp/pkg/thttp"
)

func main() {
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err)
	}

	p, err := thttp.NewPool(&thttp.PoolOptions{
		Prefix:    "thttp_",
		Image:     "thttp:debian",
		Count:     5,
		PortStart: 7070,
		// Override:  true,
	}, docker)

	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		err := p.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 5; i++ {
		go func() {
			req, _ := http.NewRequest("GET", "http://ifconfig.me", nil)

			res, err := p.Do(req, true)
			if err != nil {
				log.Println(err)
				return
			}

			ip, err := io.ReadAll(res.Body)
			if err != nil {
				log.Println(err)
				return
			}

			fmt.Printf("request with ip : %s\n", string(ip))
			wg.Done()
		}()
	}

	for i := 0; i < 5; i++ {
		go func(txID int) {
			err := p.DoTx(func(httpClient *http.Client) error {
				for c := 0; c < 5; c++ {
					req, _ := http.NewRequest("GET", "http://ifconfig.me", nil)

					res, err := httpClient.Do(req)
					if err != nil {
						log.Println(err)
						return err
					}

					ip, err := io.ReadAll(res.Body)
					if err != nil {
						log.Println(err)
						return err
					}

					log.Printf("request [tx:%d;req:%d] within transaction with ip : %s\n", txID, c, string(ip))
				}
				return nil
			})

			if err != nil {
				log.Printf("error while executing transaction %d\n", txID)
			}
			wg.Done()
		}(i)
	}

	wg.Wait()

}
