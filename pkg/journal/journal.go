// +build linux

package journal

import (
	"context"
	"fmt"

	"github.com/coreos/go-systemd/journal"
	_ "github.com/coreos/go-systemd/journal"
	"github.com/kolide/osquery-go/plugin/logger"
)

func New() *logger.Plugin {
	return logger.NewPlugin("journal", Log)
}

func Log(_ context.Context, logType logger.LogType, logText string) error {
	return journal.Send(
		fmt.Sprintf("{ %q : %q, %q : %s}", "osquery_log_type", logType, "osquery_log_json", logText),
		journal.PriInfo,
		map[string]string{"OSQUERY_LOG_TYPE": logType.String()},
	)
}
