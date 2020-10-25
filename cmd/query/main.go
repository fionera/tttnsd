package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/miekg/dns"

	"github.com/fionera/tttnsd/client"
)

const baseAddress = "tttnsd.example.com"

func main() {
	//c, err := client.NewFromOS()
	//if err != nil {
	//	log.Fatal(err)
	//}

	c := client.NewClient("127.0.0.1:8053")

	list(c, "")
}

func list(c *client.Client, id string) {
	var listUrl string
	if id == "" {
		listUrl = "list." + baseAddress
	} else {
		listUrl = id + ".list." + baseAddress
	}
	resp := doQuery(c, listUrl)
	if resp == "" {
		log.Fatal("invalid response")
	}

	var pages int
	for _, v := range strings.Split(resp, ";") {
		if strings.HasPrefix(v, "PAGES ") {
			p, err := strconv.Atoi(strings.TrimPrefix(v, "PAGES "))
			if err != nil {
				log.Fatal(err)
			}

			pages = p
		}
	}

	for i := 0; i < pages; i++ {
		resp := doQuery(c, fmt.Sprintf("%d.%s", i, listUrl))
		items := strings.Split(resp, ";")
		for _, item := range items {
			type_ := item[:2]
			parts := strings.Split(item[2:], "|")
			name := parts[0]
			itemID := parts[1]

			if type_ == "FD" {
				next := itemID
				if id != "" {
					next += "." + id
				}

				defer list(c, next)
			} else if type_ == "IT" {
				next := itemID
				if id != "" {
					next += "." + id
				}

				resp := doQuery(c, fmt.Sprintf("%s", next+"."+baseAddress))
				log.Println(name, "-", resp)
			}
		}
	}
}

func doQuery(c *client.Client, address string) string {
	m := new(dns.Msg)
	m.SetQuestion(address+".", dns.TypeTXT)
	m.RecursionDesired = true
	r, _, err := c.Exchange(m)
	if err != nil {
		log.Fatal(err)
	}

	for _, a := range r.Answer {
		if txt, ok := a.(*dns.TXT); ok {
			return txt.Txt[0]
		}
	}

	return ""
}
