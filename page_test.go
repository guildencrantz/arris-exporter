package main

import (
	"os"
	"testing"

	"golang.org/x/net/html"

	"github.com/stretchr/testify/assert"
)

func TestPageScrape(t *testing.T) {
	f, err := os.Open("./test/status.html")
	if err != nil {
		t.Fatal(err)
	}

	htm, err := html.Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	p := NewPage()
	p.scrape(htm)

	assert.Equal(t, "SB8200", p.Model, "Model Version")
}
