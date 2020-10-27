package main

import (
	"flag"
	"log"

	"github.com/miekg/dns"

	"github.com/fionera/tttnsd/proto"
)

var dir = flag.String("dir", "", "The directory to represent")
var addr = flag.String("addr", ":8053", "The Port to bind to")
var domain = flag.String("domain", "", "The Domain on which to host")

func main() {
	flag.Parse()

	if *dir == "" {
		log.Fatal("Please provide a directory")
	}

	if *domain == "" {
		log.Fatal("Please provide the Domain")
	}

	log.Println("Starting DNS Server on " + *addr)
	log.Fatal(dns.ListenAndServe(*addr, "udp", proto.NewServer(*domain, *dir)))
}
