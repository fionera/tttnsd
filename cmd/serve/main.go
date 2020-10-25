package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

const BaseDomain = "tttnsd.example.com"

var Features = []string{"FOLDER", "HREF", "TXT"}
var listRegex = regexp.MustCompile("^(?P<page_number>\\d+\\.|)(?P<folder_id>|[\\.\\w+]+|)list\\.$")
var getRegex = regexp.MustCompile("^(?P<item_id>\\w+)((?P<folder_id>|[\\.\\w+]+)|).$")

var dirFlag = flag.String("dir", "test", "The directory to represent")
var itemCache = make(map[string]*item)
var dirCache = make(map[string]*dir)

type item struct {
	name    string
	content string
	type_   string
	id      string
}

type dir struct {
	name  string
	pages []string
	id    string
	items []*item
}

func (i *item) String() string {
	return fmt.Sprintf("%s %s|%s", i.type_, i.name, i.id)
}

func main() {
	flag.Parse()

	log.Println("Creating Dir Cache. Please wait...")
	if err := filepath.Walk(*dirFlag, walkDir); err != nil {
		log.Fatal(err)
	}

	log.Println("Creating Page Cache.")
	createPageCache()

	mux := dns.NewServeMux()
	mux.Handle(BaseDomain, newWrapper(handleGeneric))
	mux.Handle("list."+BaseDomain, newWrapper(handleList))

	addr := ":8053"
	log.Println("Starting DNS Server on " + addr)
	log.Fatal(dns.ListenAndServe(addr, "udp", mux))
}

func createPageCache() {
	for _, d := range dirCache {
		var page string
		newPage := false
		for _, item := range d.items {
			s := item.String()

			if len(page) > 240 {
				newPage = true
			}
			if len(page)+len(s)+1 > 240 {
				newPage = true
			}

			if newPage {
				d.pages = append(d.pages, page)
				page = ""
				newPage = false
			}

			if page != "" {
				page += ";"
			}

			page += s
		}

		if page != "" {
			d.pages = append(d.pages, page)
		}
	}
}

func walkDir(p string, info os.FileInfo, err error) error {
	dirs := strings.Split(p, string(os.PathSeparator))
	var ids []string
	for _, d := range dirs {
		if d == *dirFlag {
			if len(dirs) == 1 {
				ids = append(ids, "root")
			}
			continue
		}

		ids = append(ids, fmt.Sprintf("%x", md5.Sum([]byte(d))))
	}

	if len(ids) == 0 {
		return nil
	}

	// Reverse the array
	for i, j := 0, len(ids)-1; i < j; i, j = i+1, j-1 {
		ids[i], ids[j] = ids[j], ids[i]
	}

	filePath := strings.Join(ids, ".")
	dirPath := strings.Join(ids[1:], ".")
	if dirPath == "" {
		dirPath = "root"
	}

	var content []byte
	if !info.IsDir() {
		c, err := ioutil.ReadFile(p)
		if err != nil {
			return err
		}
		content = c
	}

	it := &item{
		name:    info.Name(),
		content: string(content),
		type_:   getType(info),
		id:      ids[0],
	}

	itemCache[filePath] = it

	if _, ok := dirCache[dirPath]; !ok {
		n, _ := filepath.Split(p)
		p := strings.Split(n, string(os.PathSeparator))
		d := strings.Split(dirPath, ".")
		var name string
		if info.IsDir() {
			name = dirPath
		} else {
			name = p[len(p)-2]
		}

		dirCache[dirPath] = &dir{
			name: name,
			id:   d[0],
		}
	}

	if it.id == "root" {
		return nil
	}

	dirCache[dirPath].items = append(dirCache[dirPath].items, it)

	return nil
}

func getType(info os.FileInfo) string {
	if info.IsDir() {
		return "FD"
	}
	return "IT"
}

type Handler func(address string) string

func newWrapper(handler Handler) dns.Handler {
	return dns.HandlerFunc(func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg).SetReply(r)

		defer func() {
			err := w.WriteMsg(m)
			if err != nil {
				log.Println(err)
			}
		}()

		if len(r.Question) > 1 || len(r.Question) == 0 {
			log.Println("invalid question length")
			return
		}

		q := r.Question[0]
		if q.Qtype != dns.TypeTXT {
			log.Println("invalid question type")
			return
		}

		a := handler(q.Name[:len(q.Name)-1])
		if a == "" {
			m.Response = false
			m.RecursionDesired = false
			m.Rcode = dns.RcodeNameError
			return
		}

		m.Answer = append(m.Answer, NewTXT(a, q))
	})
}

func NewTXT(txt string, q dns.Question) *dns.TXT {
	return &dns.TXT{
		Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0},
		Txt: []string{txt},
	}
}

func handleGeneric(address string) string {
	// handle initial request
	if address == BaseDomain {
		return fmt.Sprintf("SRV %s;FEAT %s;", address, strings.Join(Features, ","))
	}

	return handleGet(address)
}

func handleGet(address string) string {
	// remove BaseDomain to make matching easier
	ts := strings.TrimSuffix(address, BaseDomain)
	if !getRegex.MatchString(ts) {
		return ""
	}

	matches := getRegex.FindStringSubmatch(ts)

	itemID := matches[1]
	folderIDs := matches[2]

	if it, ok := itemCache[itemID+folderIDs]; ok {
		return it.content
	}

	return ""
}

func handleList(address string) string {
	// remove BaseDomain to make matching easier
	ts := strings.TrimSuffix(address, BaseDomain)
	if !listRegex.MatchString(ts) {
		return ""
	}

	matches := listRegex.FindStringSubmatch(ts)

	pageNumber := matches[1]
	folderIDs := matches[2]

	if folderIDs == "" {
		folderIDs = "root"
	} else {
		folderIDs = folderIDs[:len(folderIDs)-1]
	}

	dir := dirCache[folderIDs]
	if dir == nil {
		return ""
	}

	if pageNumber == "" {
		return fmt.Sprintf("PAGES %d;ITEMS %d", len(dir.pages), len(dir.items))
	} else {
		pageNumber = pageNumber[:len(pageNumber)-1]
	}

	if len(dir.items) == 0 {
		return ""
	}

	p, err := strconv.Atoi(pageNumber)
	if err != nil {
		return ""
	}

	if len(dir.pages) == 0 || len(dir.pages)-1 < p {
		return ""
	}

	return dir.pages[p]
}
