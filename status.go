package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/net/html"

	"github.com/antchfx/htmlquery"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/util/sets"
)

type downstream struct {
	statusPage     *status
	Id             int
	Locked         string
	Modulation     string
	Frequency      int64
	Power          float32
	SNR            float32
	Corrected      int64
	Uncorrectables int64
}

func promPower(s *status, id int) prometheus.GaugeFunc {
	return prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace:   "arris",
			Subsystem:   "downstream",
			Name:        "power",
			Help:        "Power, in Hz, of downstream channel.",
			ConstLabels: prometheus.Labels{"channel": strconv.Itoa(id)},
		},
		func() float64 {
			return float64((*s.Downstream)[id].Power)
		},
	)
}

func promCorrected(s *status, id int) prometheus.GaugeFunc {
	return prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace:   "arris",
			Subsystem:   "downstream",
			Name:        "corrected",
			Help:        "",
			ConstLabels: prometheus.Labels{"channel": strconv.Itoa(id)},
		},
		func() float64 {
			return float64((*s.Downstream)[id].Corrected)
		},
	)
}

func promUncorrectables(s *status, id int) prometheus.GaugeFunc {
	return prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace:   "arris",
			Subsystem:   "downstream",
			Name:        "uncorrectables",
			Help:        "",
			ConstLabels: prometheus.Labels{"channel": strconv.Itoa(id)},
		},
		func() float64 {
			return float64((*s.Downstream)[id].Uncorrectables)
		},
	)
}

func promRegister(s *status, id int) {
	log := logrus.WithField("method", "downstream.Register").WithField("id", id)
	defer log.Trace("done")

	prometheus.Register(promPower(s, id))
	prometheus.Register(promCorrected(s, id))
	prometheus.Register(promUncorrectables(s, id))
}

func promUnregister(s *status, id int) {
	log := logrus.WithField("function", "promUnregister").WithField("id", id)
	defer log.Trace("done")

	prometheus.Unregister(promPower(s, id))
	prometheus.Unregister(promCorrected(s, id))
	prometheus.Unregister(promUncorrectables(s, id))
}

type upstream struct {
	Channel   int
	Locked    string
	Type      string
	Frequency int64
	Width     int64
	Power     float32
}

type status struct {
	*page
	AcquiredDownstreamChannel int64
	DownstreamChannelStatus   string
	Connectivity              string
	ConnectivityComment       string
	Downstream                *map[int]downstream
	downstreamChannels        *sets.Int
	Upstream                  *map[int]upstream
	upstreamChannels          *sets.Int
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
	if s.page.Page == "" {
		s.page.Page = "cmconnectionstatus.html"
	}
	s.page.Finalize()

	s.Downstream = &map[int]downstream{}
	s.downstreamChannels = &sets.Int{}
	s.Upstream = &map[int]upstream{}
	s.upstreamChannels = &sets.Int{}

	s.extracts.append(
		s.startup,
		s.downstream,
		s.upstream,
	)
}

func NewStatus() *status {
	s := &status{}
	s.Finalize()
	return s
}

func (s *status) Watch(d time.Duration) {
	log := logrus.WithField("method", "status.Watch")
	defer log.Trace("done")

	url := fmt.Sprintf("%s/%s", s.Host, s.Page)
	log = log.WithField("url", url)
	go func() {
		for {
			log.Trace("Get...")
			resp, err := http.Get(url)
			if err != nil {
				log.WithError(err).Error("Unable to retrieve page")
				continue
			}
			defer resp.Body.Close()
			log.Trace("Got")

			htm, err := html.Parse(resp.Body)
			if err != nil {
				log.WithError(err).Error("Unable to parse page HTML")
			}

			s.scrape(htm)

			time.Sleep(d)
		}
	}()
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

func (s *status) downstream(node *html.Node) bool {
	log := logrus.WithField("method", "status.downstreamChannels")
	defer log.Trace("Done")

	tbody := htmlquery.FindOne(node, `//th/strong[text()="Downstream Bonded Channels"]/../../..`)

	trs := htmlquery.Find(tbody, `//tr`)

	down := map[int]downstream{}
	channels := sets.Int{}
	for i := 2; i < len(trs); i++ {
		id, err := strconv.Atoi(htmlquery.InnerText(htmlquery.FindOne(trs[i], `/td[1]`)))
		if err != nil {
			log.WithError(err).WithField("index", i).Error("Unable to parse row id")
			continue
		}

		corrected, _ := strconv.ParseInt(htmlquery.InnerText(htmlquery.FindOne(trs[i], `/td[7]`)), 10, 64)
		uncorrectables, _ := strconv.ParseInt(htmlquery.InnerText(htmlquery.FindOne(trs[i], `/td[8]`)), 10, 64)

		down[id] = downstream{
			Id:             id,
			Locked:         htmlquery.InnerText(htmlquery.FindOne(trs[i], `/td[2]`)),
			Modulation:     htmlquery.InnerText(htmlquery.FindOne(trs[i], `/td[3]`)),
			Frequency:      hz(trs[i], `/td[4]`),
			Power:          db(trs[i], `/td[5]`),
			SNR:            db(trs[i], `/td[6]`),
			Corrected:      corrected,
			Uncorrectables: uncorrectables,
		}
		channels.Insert(id)
	}

	if s.downstreamChannels != nil {
		gone := s.downstreamChannels.Difference(channels)
		for id, _ := range gone {
			promUnregister(s, id)
		}

		added := channels.Difference(*s.downstreamChannels)
		for id, _ := range added {
			promRegister(s, id)
		}
	} else {
		for id, _ := range channels {
			promRegister(s, id)
		}
	}

	*s.Downstream = down
	*s.downstreamChannels = channels

	return true
}

func (s *status) upstream(node *html.Node) bool {
	log := logrus.WithField("method", "status.upstreamBonds")
	defer log.Trace("Done")

	tbody := htmlquery.FindOne(node, `//th/strong[text()="Upstream Bonded Channels"]/../../..`)

	trs := htmlquery.Find(tbody, `//tr`)

	up := map[int]upstream{}
	for i := 2; i < len(trs); i++ {
		channel, _ := strconv.Atoi(htmlquery.InnerText(htmlquery.FindOne(trs[i], `/td[1]`)))
		id, _ := strconv.Atoi(htmlquery.InnerText(htmlquery.FindOne(trs[i], `/td[2]`)))
		up[id] = upstream{
			Channel:   channel,
			Locked:    htmlquery.InnerText(htmlquery.FindOne(trs[i], `/td[3]`)),
			Type:      htmlquery.InnerText(htmlquery.FindOne(trs[i], `/td[4]`)),
			Frequency: hz(trs[i], `/td[5]`),
			Width:     hz(trs[i], `/td[6]`),
			Power:     db(trs[i], `/td[7]`),
		}
	}

	*s.Upstream = up

	return true
}
