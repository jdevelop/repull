package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"os"
	"strings"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	log     = logrus.NewEntry(logrus.New())
	verbose = flag.Bool("v", false, "verbose")
)

func main() {
	flag.Parse()
	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	if *verbose {
		log.Logger.SetLevel(logrus.DebugLevel)
	}

	include := make(map[string]struct{})
	for _, incl := range os.Args[1:] {
		include[incl] = struct{}{}
	}
	cli, err := client.NewEnvClient()
	if err != nil {
		log.WithError(err).Fatal("can't create docker client")
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		log.WithError(err).Fatalf("can't list containers")
	}
	ctx := context.Background()
	for _, container := range containers {
		pullRestart := func() {
			log.Debugf("inspecting container %s", container.ID)
			inspected, err := cli.ContainerInspect(ctx, container.ID)
			if err != nil {
				log.WithError(err).Errorf("can't inspect container %s", container.ID)
				return
			}
			log.Debugf("pulling a new image for container %s", imageString(&inspected))
			if imageId, err := pull(ctx, cli, &container); err != nil {
				log.WithError(err).Errorf("can't pull/restart container %s", imageString(&inspected))
			} else {
				log.Infof("received image id: %s", imageId)
				if err := restart(ctx, cli, &inspected, imageId); err != nil {
					log.WithError(err).Errorf("can't restart container %s", imageString(&inspected))
				}
			}
		}
		if _, ok := include[container.ID[:12]]; ok {
			pullRestart()
		} else if _, ok := include[container.ID]; ok || findByName(include, container) {
			pullRestart()
		} else {
			continue
		}
	}
}

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

func pull(ctx context.Context, cli *client.Client, c *types.Container) (string, error) {
	log.Debugf("pulling image %s", c.Image)
	distributionRef, err := reference.ParseNormalizedNamed(c.Image)
	switch {
	case err != nil:
		return "", errors.Wrapf(err, "can't parse image '%s'", c.Image)
	case reference.IsNameOnly(distributionRef):
		distributionRef = reference.TagNameOnly(distributionRef)
		if tagged, ok := distributionRef.(reference.Tagged); ok {
			log.Infof("using default tag: %s\n", tagged.Tag())
		}
	}
	auth, err := findDefaultAuth(distributionRef)
	if err != nil {
		return "", errors.Wrapf(err, "can't fetch auth for %s", distributionRef.String())
	}
	var regAuth string
	if auth != nil {
		encodedJSON, err := json.Marshal(auth)
		if err != nil {
			return "", err
		}
		regAuth = base64.URLEncoding.EncodeToString(encodedJSON)
	}
	var imageId string
	if r, err := cli.ImagePull(context.Background(), distributionRef.String(), types.ImagePullOptions{
		RegistryAuth: regAuth,
	}); err != nil {
		return "", errors.Wrapf(err, "can't pull image %s", distributionRef.String())
	} else {
		var jm jsonmessage.JSONMessage
		dec := json.NewDecoder(r)
		defer r.Close()
		for {
			if err := dec.Decode(&jm); err != nil {
				break
			}
			if strings.HasPrefix(jm.Status, "Digest: sha256:") {
				imageId = jm.Status[8:]
			}
		}
	}
	return imageId, nil
}

func findByName(include map[string]struct{}, c types.Container) bool {
	for _, name := range c.Names {
		if _, ok := include[name[1:]]; ok {
			return true
		}
	}
	return false
}

func imageString(c *types.ContainerJSON) string {
	return c.ID + " : [" + c.Name + "]"
}
