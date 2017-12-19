package main

import (
	"log"

	osquery "github.com/kolide/osquery-go"

	"github.com/kolide/go-extension-tutorial/pkg/journal"
	"github.com/kolide/go-extension-tutorial/pkg/systemd"
)

func registerPlatformPlugins(server *osquery.ExtensionManagerServer) {
	// create and register the systemd table plugin.
	systemdPlugin, err := systemd.New()
	if err != nil {
		log.Fatalf("Error creating systemd plugin: %s\n", err)
	}
	server.RegisterPlugin(systemdPlugin.Table())

	// create and register the journal logger plugin.
	journalLogger := journal.New()
	server.RegisterPlugin(journalLogger)
}
