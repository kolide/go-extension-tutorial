package main

import (
	"flag"
	"log"
	"time"

	osquery "github.com/kolide/osquery-go"

	"github.com/kolide/go-extension-tutorial/pkg/gist"
	"github.com/kolide/go-extension-tutorial/pkg/journal"
	"github.com/kolide/go-extension-tutorial/pkg/systemd"
	"github.com/kolide/go-extension-tutorial/pkg/twitter"
)

func main() {
	var (
		flSocketPath = flag.String("socket", "", "")
		flTimeout    = flag.Int("timeout", 0, "")
		_            = flag.Int("interval", 0, "")
		_            = flag.Bool("verbose", false, "")
	)
	flag.Parse()

	// allow for osqueryd to create the socket path
	time.Sleep(2 * time.Second)

	// create an extension server
	server, err := osquery.NewExtensionManagerServer(
		"com.kolide.go_extension_tutorial",
		*flSocketPath,
		osquery.ServerTimeout(time.Duration(*flTimeout)*time.Second),
	)
	if err != nil {
		log.Fatalf("Error creating extension: %s\n", err)
	}

	// create and register the systemd table plugin.
	systemdPlugin, err := systemd.New()
	if err != nil {
		log.Fatalf("Error creating systemd plugin: %s\n", err)
	}
	server.RegisterPlugin(systemdPlugin.Table())

	// create and register the journal logger plugin.
	journalLogger := journal.New()
	server.RegisterPlugin(journalLogger)

	// create and register the twitter distributed plugin.
	// requires the configuration to be passed through env vars.
	twitterPlugin, err := twitter.New()
	if err != nil {
		log.Fatal(err)
	}
	go twitterPlugin.Run()
	defer twitterPlugin.Stop()

	server.RegisterPlugin(twitterPlugin.Distributed())

	// create and register gist config plugin.
	// requires configuration to be available through environment variables.
	server.RegisterPlugin(gist.New())

	// run the extension server
	log.Fatal(server.Run())
}
