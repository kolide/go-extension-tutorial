package spotlight

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/kolide/osquery-go/plugin/table"
	"github.com/pkg/errors"
)

func New() *table.Plugin {
	columns := []table.ColumnDefinition{
		table.TextColumn("query"),
		table.TextColumn("path"),
	}
	return table.NewPlugin("spotlight", columns, generateSpotlight)
}

func generateSpotlight(ctx context.Context, queryContext table.QueryContext) ([]map[string]string, error) {
	q, ok := queryContext.Constraints["query"]
	if !ok || len(q.Constraints) == 0 {
		return nil, errors.New("The spotlight table requires that you specify a constraint WHERE query =")
	}
	where := q.Constraints[0].Expression
	var query []string
	if strings.Contains(where, "-") {
		query = strings.Split(where, " ")
	} else {
		query = []string{where}
	}
	lines, err := mdfind(query...)
	if err != nil {
		return nil, errors.Wrap(err, "call mdfind")
	}
	var resp []map[string]string
	for _, line := range lines {
		m := make(map[string]string, 2)
		m["query"] = where
		m["path"] = line
		resp = append(resp, m)
	}
	return resp, nil
}

func mdfind(args ...string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	path := "/usr/bin/mdfind"

	out, err := exec.CommandContext(ctx, path, args...).Output()
	if err != nil {
		return nil, err
	}
	var lines []string
	lr := bufio.NewReader(bytes.NewReader(out))
	for {
		line, _, err := lr.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		lines = append(lines, string(line))
	}
	return lines, nil
}
