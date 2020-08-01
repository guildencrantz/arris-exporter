package main

import (
	"io/ioutil"
	"os"
	"testing"

	"golang.org/x/net/html"

	"github.com/stretchr/testify/assert"
)

func TestStatusScrape(t *testing.T) {
	f, err := os.Open("./test/status.html")
	if err != nil {
		t.Fatal(err)
	}

	body, err := html.Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	s := NewStatus()
	s.scrape(body)

	expected, err := ioutil.ReadFile("test/status.json")
	if err != nil {
		t.Log(err)
		t.Fatal("Unable to open expected status")
	}

	assert.JSONEq(t, string(expected), s.String())
}
