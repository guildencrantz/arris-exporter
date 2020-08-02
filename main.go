package main

import (
	"strconv"
	"strings"

	"golang.org/x/net/html"

	"github.com/antchfx/htmlquery"
	"github.com/sirupsen/logrus"
)

func init() {
	// TODO: Flag
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {
	log := logrus.WithField("func", "main")
	defer log.Debug("Done")

	scrape()
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
