package main

import (
	"flag"
	"github.com/myl7/zyzzyva/pkg/client"
	"github.com/myl7/zyzzyva/pkg/conf"
	"github.com/myl7/zyzzyva/pkg/server"
	"log"
)

func main() {
	id := flag.Int("id", 0, "Client ID or Server ID")
	flag.Parse()
	conf.InitKeys(*id)

	log.Printf("ID %d started", *id)

	if *id >= conf.N {
		c := client.NewClient(*id)
		c.Run()
	} else {
		s := server.NewServer(*id)
		s.Run()
	}
}
