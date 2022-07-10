# THTTP

Thttp is a cloud-native (that sound cooler than "using docker") HTTP connection pooler which use a docker client to create tor containers then give you a Go HTTP client you can use to execute http call anonymously. Let's see how we can use it :

First you need to build the container using the Dockerfile included in this repo :

```bash
$ docker build -t thttp:debian .
```

Then let's make your first request using thttp :

```go
package main

import (
    "github.com/docker/docker/client"
	"github.com/m-vinc/thttp/pkg/thttp"
)

// In this case, docker is going to use the local unix socket but you can configure this client as you want.
docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
if err != nil {
    log.Fatal(err)
}

p, err := thttp.NewPool(&thttp.PoolOptions{
    // containers names prefix, in this case containers is going to be called : thttp_0, thttp_1, thttp_2, thttp_3, thttp_4
    Prefix:    "thttp_",
    
    // the docker image to run for these containers, we use the image we built above
    Image:     "thttp:debian",
    
    // number of tor client to run
    Count:     2,

    // thttp use the PortStart value to know on which port the tor sock5 will listen
    // is this case, there will be 5 tor clients using port : 7070, 7071, 7072, 7073, 7074.
    PortStart: 7070,
    // if Override is set to true, thttp is going to clear all previous containers and re-create them.
    // Override:  true,

    // the debug flag print a lots of informations about connections
    // Debug: true
}, docker)

if err != nil {
    log.Fatal(err)
}
```

For now the pool is going to check each connections on first start, this can be long, let's call `ifconfig.me` with the pool :

```Go
req, _ := http.NewRequest("GET", "http://ifconfig.me", nil)
res, err := p.Do(req, true) // This boolean is to indicate thttp to change the IP after this call
if err != nil {
    log.Println(err)
    return
}
log.Println(res, err)
```

If you want to multiple call using a same IP you can use the pool's `doTx` function :

```Go
err := p.DoTx(func(httpClient *http.Client) error {
    // Launch 5 call using the same connections
    for c := 0; c < 5; c++ {
        req, _ := http.NewRequest("GET", "http://ifconfig.me", nil)

        res, err := httpClient.Do(req)
        if err != nil {
            return err
        }

        ip, err := io.ReadAll(res.Body)
        if err != nil {
            return err
        }

        log.Printf("request [req:%d] within transaction with ip : %s\n", c, string(ip))
        return nil
    }
    return nil
})

if err != nil {
    log.Println("error while executing transaction")
}
```

You can find this example in `cmd/readme` and an other example making parralels requests in `cmd/demo` !

## Improvements

* Using the control socket of tor to renew the circuit instead of using docker to trigger a SIGHUP.
* Don't wait the startup initialization and start the requests directly when a single tor instance is up and ready to accept connections.

## Why

Why not ?



