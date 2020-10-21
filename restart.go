package main

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
)

func restart(ctx context.Context, cli *client.Client, inspected *types.ContainerJSON, imageId string) error {
	log.Debugf("killing container %s", imageString(inspected))
	if err := cli.ContainerKill(ctx, inspected.ID, "sigkill"); err != nil {
		return errors.Wrapf(err, "can't kill container %s", imageString(inspected))
	}
	if _, err := cli.ContainerWait(ctx, inspected.ID); err != nil {
		return errors.Wrapf(err, "can't get the container status %s", imageString(inspected))
	}
	log.Debugf("removing container %s", imageString(inspected))
	if err := cli.ContainerRemove(ctx, inspected.ID, types.ContainerRemoveOptions{}); err != nil {
		return errors.Wrapf(err, "can't remove container %s", imageString(inspected))
	}
	log.Debugf("replacing image %s with %s", inspected.Config.Image, imageId)
	newC, err := cli.ContainerCreate(ctx, inspected.Config, inspected.HostConfig, &network.NetworkingConfig{
		EndpointsConfig: inspected.NetworkSettings.Networks,
	}, inspected.Name)
	if err != nil {
		return errors.Wrapf(err, "can't create container %s", imageString(inspected))
	}
	if err := cli.ContainerStart(ctx, newC.ID, types.ContainerStartOptions{}); err != nil {
		return errors.Wrapf(err, "can't statr new container %s", newC.ID)
	}
	log.Infof("started container %s", newC.ID)
	return nil
}
