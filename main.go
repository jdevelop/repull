package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

var (
	log     = logrus.NewEntry(logrus.New())
	verbose = flag.Bool("v", false, "verbose")
)

func main() {
	flag.Parse()
	if len(flag.Args()) == 0 {
		fmt.Println("Usage: " + os.Args[0] + " [flags] containerId1 containerName2 ... containerNameN")
		flag.PrintDefaults()
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
