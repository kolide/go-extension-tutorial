package gist

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/github"
	"github.com/kolide/osquery-go/plugin/config"
	"github.com/pkg/errors"
)

type Plugin struct {
	gistID string
	client *github.Client
}

func New() *config.Plugin {
	gistID := os.Getenv("OSQUERY_CONFIG_GIST")
	client := github.NewClient(nil)
	plugin := &Plugin{client: client, gistID: gistID}
	return config.NewPlugin("gist", plugin.GenerateConfigs)
}

func (p *Plugin) GenerateConfigs(ctx context.Context) (map[string]string, error) {
	gist, _, err := p.client.Gists.Get(ctx, p.gistID)
	if err != nil {
		return nil, errors.Wrap(err, "get config gist")
	}
	var config string
	if file, ok := gist.Files["osquery.conf"]; ok {
		config = file.GetContent()
	} else {
		return nil, fmt.Errorf("no osquery.conf file in gist %s", p.gistID)
	}
	return map[string]string{"gist": config}, nil
}
