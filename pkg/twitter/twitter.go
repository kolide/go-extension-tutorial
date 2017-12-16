package twitter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/boltdb/bolt"
	"github.com/kolide/osquery-go/plugin/distributed"
	"github.com/pkg/errors"
)

type Plugin struct {
	api    *anaconda.TwitterApi
	stream *anaconda.Stream
	db     *bolt.DB
}

// Distributed creates an osquery distributed plugin which uses
// the twitter public stream for the read method and posts results as replies.
// The plugin uses boltdb as a local query queue.
func (p *Plugin) Distributed() *distributed.Plugin {
	return distributed.NewPlugin("twitter", p.getQueries, p.writeResults)
}

// getQueries fetches queries parsed from tweets using boltdb as a query queue.
func (p *Plugin) getQueries(ctx context.Context) (*distributed.GetQueriesResult, error) {
	queries := make(map[string]string)
	tx, err := p.db.Begin(true)
	if err != nil {
		return nil, err
	}
	bkt := tx.Bucket([]byte(bucketName))
	if bkt == nil {
		return nil, err
	}
	c := bkt.Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		queries[string(k)] = string(v)
		c.Delete()
	}
	return &distributed.GetQueriesResult{Queries: queries}, tx.Commit()
}

// writeResults writes the results of queries to a go playground snippet and posts a twitter reply.
func (p *Plugin) writeResults(ctx context.Context, results []distributed.Result) error {
	for _, result := range results {

		// encode query result as json
		buf := new(bytes.Buffer)
		enc := json.NewEncoder(buf)
		enc.SetIndent("", "  ")
		if err := enc.Encode(result); err != nil {
			return err
		}

		// post the result to the Go Playground
		share, err := postToPlayGround(buf.String())
		if err != nil {
			return err
		}

		id, _ := strconv.Atoi(result.QueryName)
		tweet, err := p.api.GetTweet(int64(id), nil)
		if err != nil {
			return err
		}

		// create a reply to the tweet which requested the query.
		v := url.Values{}
		v.Set("in_reply_to_status_id", result.QueryName)
		p.api.PostTweet(fmt.Sprintf("@%s https://play.golang.org/p/%s", tweet.User.ScreenName, share), v)
	}
	return nil
}

func New() (*Plugin, error) {
	ck := os.Getenv("TWITTER_CONSUMER_KEY")
	cs := os.Getenv("TWITTER_CONSUMER_SECRET")
	at := os.Getenv("TWITTER_ACCESS_TOKEN")
	as := os.Getenv("TWITTER_ACCESS_TOKEN_SECRET")
	anaconda.SetConsumerKey(ck)
	anaconda.SetConsumerSecret(cs)
	api := anaconda.NewTwitterApi(at, as)

	dbPath := filepath.Join("/tmp", "osquery_twitter_plugin.db")
	db, err := bolt.Open(dbPath, 0644, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, errors.Wrap(err, "opening boltdb for osquery twitter plugin")
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	})
	if err != nil {
		return nil, errors.Wrapf(err, "creating %s bucket", bucketName)
	}

	stream := api.PublicStreamFilter(url.Values{
		"track": []string{"#osqueryquery"},
	})

	plugin := &Plugin{api: api, db: db, stream: stream}
	return plugin, nil
}

const bucketName = "osqueryquerytweet"

func (p *Plugin) Run() {
	for v := range p.stream.C {
		t, ok := v.(anaconda.Tweet)
		if !ok {
			continue
		}

		if t.RetweetedStatus != nil {
			continue
		}

		query := queryFromTweet(t.FullText)
		if !strings.Contains(strings.ToUpper(query), "SELECT") {
			continue
		}

		tx, err := p.db.Begin(true)
		if err != nil {
			log.Println(err)
			continue
		}

		bkt := tx.Bucket([]byte(bucketName))
		if bkt == nil {
			continue
		}

		bkt.Put([]byte(t.IdStr), []byte(query))
		tx.Commit()

		p.api.Retweet(t.Id, false)
	}
}

func queryFromTweet(text string) string {
	query := strings.TrimSpace(strings.TrimRight(text, "#osqueryquery"))
	if strings.Contains(query, "#osqueryquery") {
		query = strings.TrimSpace(strings.TrimLeft(text, "#osqueryquery"))
	}
	return query
}

func (p *Plugin) Stop() {
	p.stream.Stop()
}

func postToPlayGround(text string) (string, error) {
	var tmpl = `
package main

import (
	"encoding/json"
	"fmt"
	"log"
)

func main() {
	var results = struct {
		Query_Name string
		Status     int
		Rows       []map[string]string
	}{}
	if err := json.Unmarshal(data, &results); err != nil {
		log.Fatal(err)
	}
	fmt.Println(results)
}


var data = []byte(%s)
`
	data := fmt.Sprintf(tmpl, "`"+text+"`")
	resp, err := http.Post("https://play.golang.org/share", "application/x-www-form-urlencoded; charset=UTF-8", strings.NewReader(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	url, err := ioutil.ReadAll(resp.Body)
	return string(url), nil
}
