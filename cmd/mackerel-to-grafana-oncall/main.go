package main

import (
	"log"
	"os"

	oncall "github.com/fujiwara/mackerel-to-grafana-oncall"
)

var Version = "current"

func main() {
	oncall.Version = Version
	if err := oncall.Run(); err != nil {
		log.Println("[error] ", err)
		os.Exit(1)
	}
}
