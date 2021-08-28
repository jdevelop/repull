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
			"auth": "YWFhOmFhYQo="
		},
		"docker.io": {
			"auth": "eXl5Onl5eQo="
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
		{ref: "16371954.dkr.ecr.us-east-1.amazonaws.com/user/project/server", expected: "aaa"},
		{ref: "user/project/server", expected: "yyy"},
	} {
		named, err := reference.ParseAnyReference(tst.ref)
		require.NoError(t, err)
		t.Logf("Reference: +%v", named)
		authKey, err := parseDefaultAuth(named, []byte(auth))
		if err != nil {
			require.Fail(t, err.Error())
		}
		require.Equal(t, tst.expected, authKey.Username)
	}
}
