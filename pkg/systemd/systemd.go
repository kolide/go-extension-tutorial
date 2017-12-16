package systemd

import (
	"context"
	"strconv"

	"github.com/coreos/go-systemd/dbus"

	"github.com/kolide/osquery-go/plugin/table"
)

type Plugin struct {
	conn *dbus.Conn
}

func New() (*Plugin, error) {
	conn, err := dbus.New()
	if err != nil {
		return nil, err
	}
	return &Plugin{conn: conn}, nil
}

func (p *Plugin) Table() *table.Plugin {
	columns := []table.ColumnDefinition{
		table.TextColumn("name"),
		table.IntegerColumn("pid"),
		table.TextColumn("load_state"),
		table.TextColumn("active_state"),
		table.TextColumn("exec_start"),
		table.TextColumn("unit_path"),
		table.TextColumn("stdout_path"),
		table.TextColumn("stderr_path"),
	}
	return table.NewPlugin("systemd", columns, p.generateSystemdUnitStatus)
}

func (p *Plugin) generateSystemdUnitStatus(ctx context.Context, queryContext table.QueryContext) ([]map[string]string, error) {
	units, err := p.conn.ListUnits()
	if err != nil {
		return nil, err
	}
	var results []map[string]string
	for _, unit := range units {
		var execStart string
		if p, err := p.conn.GetServiceProperty(unit.Name, "ExecStart"); err == nil {
			execStart = p.Value.String()
		}

		var pid int
		if p, err := p.conn.GetServiceProperty(unit.Name, "MainPID"); err == nil {
			pid = int(p.Value.Value().(uint32))
		}

		var stdoutPath string
		if p, err := p.conn.GetServiceProperty(unit.Name, "StandardOutput"); err == nil {
			stdoutPath = p.Value.String()
		}

		var stderrPath string
		if p, err := p.conn.GetServiceProperty(unit.Name, "StandardError"); err == nil {
			stderrPath = p.Value.String()
		}

		results = append(results, map[string]string{
			"name":         unit.Name,
			"load_state":   unit.LoadState,
			"active_state": unit.ActiveState,
			"exec_start":   execStart,
			"pid":          strconv.Itoa(pid),
			"stdout_path":  stdoutPath,
			"stderr_path":  stderrPath,
		})
	}
	return results, nil
}
