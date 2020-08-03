package main

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"

	"github.com/antchfx/htmlquery"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func init() {
	// TODO: Flag
	logrus.SetLevel(logrus.TraceLevel)
}

func main() {
	log := logrus.WithField("func", "main")
	defer log.Debug("Done")

	s := NewStatus()
	s.Watch(15 * time.Second)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":9100", nil)
}

func hz(node *html.Node, xp string) int64 {
	str := htmlquery.InnerText(htmlquery.FindOne(node, xp))
	hz, _ := strconv.ParseInt(str[:len(str)-3], 10, 64)
	return hz
}

func db(node *html.Node, xp string) float32 {
	str := htmlquery.InnerText(htmlquery.FindOne(node, xp))
	words := strings.Split(str, " ")
	db, _ := strconv.ParseFloat(words[0], 32)
	return float32(db)
}

func removedKeys(a, b map[string]interface{}) []string {
	x := sort.StringSlice(make([]string, 0, len(a)))
	for k, _ := range a {
		if _, ok := b[k]; !ok {
			x = append(x, k)
		}
	}

	sort.Sort(x)
	return x
}
