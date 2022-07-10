package tor

import (
	"context"
	"fmt"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func ContainerExist(ctx context.Context, docker *client.Client, name string) (*types.Container, error) {
	containerName := fmt.Sprintf("/%s", name)
	list, err := docker.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}

	for _, cont := range list {
		for _, name := range cont.Names {
			if name == containerName {
				return &cont, nil
			}
		}
	}

	return nil, nil
}

func NewTorContainer(ctx context.Context, docker *client.Client, options *TorOptions) (*TorContainer, error) {
	exist, err := ContainerExist(ctx, docker, options.Name)
	if err != nil {
		return nil, err
	}

	if options.Override {
		if exist != nil {
			err = docker.ContainerRemove(ctx, exist.ID, types.ContainerRemoveOptions{
				Force: true,
			})
			if err != nil {
				return nil, err
			}
		}
	} else if exist != nil {
		return &TorContainer{
			ID:    exist.ID,
			Name:  options.Name,
			Image: options.Image,

			Port:  options.Port,
			Debug: options.Debug,
		}, nil
	}

	res, err := docker.ContainerCreate(context.Background(), &container.Config{
		Image: options.Image,
		ExposedPorts: nat.PortSet{
			"9050/tcp": struct{}{},
		},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(fmt.Sprintf("%s/tcp", "9050")): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: options.Port}},
		},
	}, nil, nil, options.Name)
	if err != nil {
		log.Fatal(err)
	}

	err = docker.ContainerStart(ctx, res.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, err
	}

	return &TorContainer{
		ID:    res.ID,
		Name:  options.Name,
		Image: options.Image,

		Port: options.Port,
	}, nil
}
