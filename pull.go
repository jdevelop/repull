package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/pkg/errors"
)

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
