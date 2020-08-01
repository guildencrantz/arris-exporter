package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/net/html"

	"github.com/antchfx/htmlquery"
	"github.com/sirupsen/logrus"
)

type downstream struct {
	Id             int
	Locked         string
	Modulation     string
	Frequency      int64
	Power          float32
	SNR            float32
	Corrected      int64
	Uncorrectables int64
}

type status struct {
	*page
	AcquiredDownstreamChannel int64
	DownstreamChannelStatus   string
	Connectivity              string
	ConnectivityComment       string
	Downstream                []downstream
}

func (s *status) String() string {
	str, err := json.Marshal(s)
	if err != nil {
		return fmt.Sprintf("%#v", s)
	}
	return string(str)
}

func (s *status) Finalize() {
	log := logrus.WithField("method", "status.Finalize")
	defer log.Trace("done")

	if s.page == nil {
		s.page = NewPage()
	}
	s.page.Finalize()

	s.extracts.append(s.startup, s.downstreamChannels)
}

func NewStatus() *status {
	s := &status{}
	s.Finalize()
	return s
}

func (s *status) startup(node *html.Node) bool {
	log := logrus.WithField("method", "status.startup")
	defer log.Trace("Done")

	tbody := htmlquery.FindOne(node, `//th[text()="Startup Procedure"]/../..`)
	if tbody == nil {
		log.Trace("Table not found")
		return false
	}

	return s.acquiredDownstreamChannel(tbody) && s.connectivityStatus(tbody)
}

func (s *status) acquiredDownstreamChannel(node *html.Node) bool {
	log := logrus.WithField("method", "status.acquiredDownstreamChannel")
	defer log.Trace("Done")

	tr := htmlquery.FindOne(node, `//tr[3]`)
	if tr == nil {
		log.Trace("Downstream row not found")
		return false
	}
	s.AcquiredDownstreamChannel = hz(tr, `//td[2]`)
	s.DownstreamChannelStatus = htmlquery.InnerText(htmlquery.FindOne(tr, `//td[3]`))

	return true
}

func (s *status) connectivityStatus(node *html.Node) bool {
	log := logrus.WithField("method", "status.connectivityStatus")
	defer log.Trace("Done")

	tr := htmlquery.FindOne(node, `//tr[4]`)
	if tr == nil {
		log.Trace("Connectivity row not found")
		return false
	}
	s.Connectivity = htmlquery.InnerText(htmlquery.FindOne(tr, `//td[2]`))
	s.ConnectivityComment = htmlquery.InnerText(htmlquery.FindOne(tr, `//td[3]`))

	return true
}

func (s *status) downstreamChannels(node *html.Node) bool {
	log := logrus.WithField("method", "status.connectivityStatus")
	defer log.Trace("Done")

	tbody := htmlquery.FindOne(node, `//th/strong[text()="Downstream Bonded Channels"]/../../..`)

	trs := htmlquery.Find(tbody, `//tr`)

	s.Downstream = make([]downstream, len(trs)-2)
	for i := 2; i < len(trs); i++ {
		id, err := strconv.Atoi(htmlquery.InnerText(htmlquery.FindOne(trs[i], `/td[1]`)))
		if err != nil {
			log.WithError(err).WithField("index", i).Error("Unable to parse row id")
			continue
		}

		corrected, _ := strconv.ParseInt(htmlquery.InnerText(htmlquery.FindOne(trs[i], `/td[7]`)), 10, 64)
		uncorrectables, _ := strconv.ParseInt(htmlquery.InnerText(htmlquery.FindOne(trs[i], `/td[8]`)), 10, 64)

		s.Downstream[i-2] = downstream{
			Id:             id,
			Locked:         htmlquery.InnerText(htmlquery.FindOne(trs[i], `/td[2]`)),
			Modulation:     htmlquery.InnerText(htmlquery.FindOne(trs[i], `/td[3]`)),
			Frequency:      hz(trs[i], `/td[4]`),
			Power:          db(trs[i], `/td[5]`),
			SNR:            db(trs[i], `/td[6]`),
			Corrected:      corrected,
			Uncorrectables: uncorrectables,
		}
	}

	return true
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
