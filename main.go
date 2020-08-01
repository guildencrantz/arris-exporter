package main

import "github.com/sirupsen/logrus"

func init() {
	// TODO: Flag
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {
	log := logrus.WithField("func", "main")
	defer log.Debug("Done")

	scrape()
}
