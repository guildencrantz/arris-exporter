package main

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

func scrape() {
	log := logrus.WithField("func", "scrape")
	defer log.Debug("Done")

	resp, err := http.Get("http://192.168.100.1")
	if err != nil {
		// FIXME
		log.WithError(err).Fatal("unable to get page")
	}
	defer resp.Body.Close()

	page, err := html.Parse(resp.Body)
	if err != nil {
		//FIXME
		log.WithError(err).Fatal("Unable to parse page")
	}

	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, page)
	log.Info(buf.String())

	time.Sleep(1000)
}
