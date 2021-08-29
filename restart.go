package main

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
)

func containerWait(ctx context.Context, cli *client.Client, inspected *types.ContainerJSON) error {
	resultC, errC := cli.ContainerWait(ctx, inspected.ID, container.WaitConditionNotRunning)
	select {
	case <-resultC:
		return nil
	case err := <-errC:
		return errors.Wrapf(err, "can't get the container status %s", imageString(inspected))
	case <-ctx.Done():
		return fmt.Errorf("container wait timed out %s", imageString(inspected))
	}
}

func restart(ctx context.Context, cli *client.Client, inspected *types.ContainerJSON, imageId string) error {
	log.Debugf("stopping container %s", imageString(inspected))
	if err := cli.ContainerStop(ctx, inspected.ID, nil); err != nil {
		return errors.Wrapf(err, "can't kill container %s", imageString(inspected))
	}
	if err := containerWait(ctx, cli, inspected); err != nil {
		return err
	}
	log.Debugf("removing container %s", imageString(inspected))
	if err := cli.ContainerRemove(ctx, inspected.ID, types.ContainerRemoveOptions{}); err != nil {
		return errors.Wrapf(err, "can't remove container %s", imageString(inspected))
	}
	log.Debugf("replacing image %s with %s", inspected.Config.Image, imageId)
	newC, err := cli.ContainerCreate(ctx, inspected.Config, inspected.HostConfig, &network.NetworkingConfig{
		EndpointsConfig: inspected.NetworkSettings.Networks,
	}, nil, inspected.Name)
	if err != nil {
		return errors.Wrapf(err, "can't create container %s", imageString(inspected))
	}
	if err := cli.ContainerStart(ctx, newC.ID, types.ContainerStartOptions{}); err != nil {
		return errors.Wrapf(err, "can't statr new container %s", newC.ID)
	}
	log.Infof("started container %s", newC.ID)
	return nil
}
