package main

import (
	"encoding/json"
	"fmt"
	"sort"

	"golang.org/x/net/html"

	"github.com/antchfx/htmlquery"
	"github.com/sirupsen/logrus"
)

type extract func(node *html.Node) bool
type extracts []extract

func (xs *extracts) append(x ...extract) {
	log := logrus.WithField("method", "extracts.append")
	defer log.Trace("Done")

	n := append(*xs, x...)
	*xs = n
}

func (xs *extracts) remove(i int) {
	log := logrus.WithField("index", i).WithField("method", "extracts.remove")
	defer func() { log.WithField("len", len(*xs)).Trace("Done") }()

	var n extracts

	l := len(*xs)

	if i == l-1 {
		n = (*xs)[:l-1]
	} else if i == 0 {
		n = (*xs)[1:]
	} else {
		n = append((*xs)[:i], (*xs)[i+1:]...)
	}

	*xs = n
	return
}

func (xs *extracts) removes(is []int) {
	log := logrus.WithField("method", "extracts.removes")
	defer log.WithField("indexes", is).Trace("Done")

	sort.Sort(sort.Reverse(sort.IntSlice(is)))
	for _, i := range is {
		xs.remove(i)
	}
}

type page struct {
	Model    string
	Host     string
	Page     string
	extracts *extracts
}

func (p *page) Finalize() {
	if p.Host == "" {
		p.Host = "http://192.168.100.1"
	}

	if p.extracts == nil {
		p.extracts = &extracts{p.model}
	}
}

func NewPage() *page {
	p := &page{}
	p.Finalize()
	return p
}

func (p *page) String() string {
	ret, err := json.Marshal(p)
	if err != nil {
		return fmt.Sprintf("%#v", p)
	}
	return string(ret)
}

func (p *page) scrape(node *html.Node) bool {
	log := logrus.WithField("method", "page.scrape")
	defer log.WithField("page", p).Trace("Done")

	ret := false
	for _, x := range *p.extracts {
		ret = x(node) && ret
	}

	return ret
}

func (p *page) model(node *html.Node) bool {
	span := htmlquery.FindOne(node, `//span[@id="thisModelNumberIs"]`)
	if span == nil {
		return false
	}
	p.Model = htmlquery.InnerText(span)
	return true
}
