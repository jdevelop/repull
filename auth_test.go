package main

import (
	"testing"

	"github.com/docker/distribution/reference"
	"github.com/stretchr/testify/require"
)

func TestAuthParse(t *testing.T) {
	var auth = `
{
	"auths": {
		"16371954.dkr.ecr.us-east-1.amazonaws.com": {
			"auth": "aaaaa"
		},
		"docker.io": {
			"auth": "yyyyyyy"
		}
	},
	"HttpHeaders": {
		"User-Agent": "Docker-Client/19.03.12-ce (linux)"
	}
}
`
	for _, tst := range []struct {
		ref      string
		expected string
	}{
		{ref: "16371954.dkr.ecr.us-east-1.amazonaws.com/user/project/server", expected: "aaaaa"},
		{ref: "user/project/server", expected: "yyyyyyy"},
	} {
		named, err := reference.ParseAnyReference(tst.ref)
		require.NoError(t, err)
		t.Logf("Reference: +%v", named)
		authKey := parseDefaultAuth(named, []byte(auth))
		require.Equal(t, tst.expected, authKey)
	}
}
