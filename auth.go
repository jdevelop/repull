package main

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// findDefaultAuth reads Docker JSON config from user's home directory and
// attempts to extract an authentication information from it.
func findDefaultAuth(ref reference.Reference) (*types.AuthConfig, error) {
	u, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	confPath := filepath.Join(u, ".docker", "config.json")
	conf, err := os.Open(confPath)
	if err != nil {
		return nil, errors.Wrapf(err, "can't read JSON from %s", confPath)
	}
	defer conf.Close()
	data, err := ioutil.ReadAll(conf)
	if err != nil {
		return nil, err
	}
	return parseDefaultAuth(ref, data)
}

// parseDefaultAuth parses a Base-64 string from docker config file and returns
// either AuthConfig or error. If no authentication was found - AuthConfig is nil.
func parseDefaultAuth(ref reference.Reference, src []byte) (*types.AuthConfig, error) {
	parsed := gjson.ParseBytes(src)
	var auth gjson.Result
	switch v := ref.(type) {
	case reference.Named:
		auth = parsed.Get("auths").Get(cleanup(reference.Domain(v))).Get("auth")
	default:
		auth = parsed.Get("auths").Get(cleanup(v.String())).Get("auth")
	}
	if auth.Exists() {
		res, err := base64.StdEncoding.DecodeString(auth.String())
		if err != nil {
			return nil, errors.Wrapf(err, "can't decode base64 '%s'", auth.String())
		}
		args := strings.SplitN(string(res), ":", 2)
		if len(args) != 2 {
			return nil, errors.Errorf("expected 2 elements, actual %d", len(args))
		}
		return &types.AuthConfig{
			Username: strings.TrimSpace(args[0]),
			Password: strings.TrimSpace(args[1]),
		}, nil
	}
	return nil, nil
}

func cleanup(domainString string) string {
	return strings.ReplaceAll(domainString, ".", `\.`)
}
