package proto

import (
	"log"

	"github.com/miekg/dns"

	"github.com/fionera/tttnsd/vfs"
)

type Server struct {
	baseAddress string
	rootFolder  string
	vfs         *vfs.VFS
}

func NewServer(baseAddress string, rootFolder string) *Server {
	return &Server{
		baseAddress: baseAddress,
		vfs:         vfs.NewVFS(rootFolder),
	}
}

func (s Server) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)

	defer func() {
		if err := w.WriteMsg(m); err != nil {
			log.Println(err)
		}
	}()

	if r.Question[0].Qtype != dns.TypeTXT {
		return
	}

	name := r.Question[0].Name
	switch GetAddressType(s.baseAddress+".", name) {
	case Unknown:
		return
	case ServerInfoAddress:
		s.handleServerInfo(m)
	case ListAddress:
		page, folderPath := DecodeListAddress(s.baseAddress+".", name)
		s.handleList(m, page, folderPath)
	case FolderInfoAddress:
		folderPath := DecodeFolderInfoAddress(s.baseAddress+".", name)
		s.handleFolderInfo(m, folderPath)
	case ItemAddress:
		itemID, folderPath := DecodeItemAddress(s.baseAddress+".", name)
		s.handleItem(m, itemID, folderPath)
	}

	if len(m.Answer) == 0 && len(m.Extra) == 0 {
		m.Rcode = dns.RcodeNameError
	}

	for _, rr := range m.Answer {
		hdr := rr.Header()
		hdr.Rrtype = dns.TypeTXT
		hdr.Class = dns.ClassINET
		hdr.Name = name
	}

	for _, rr := range m.Extra {
		hdr := rr.Header()
		hdr.Rrtype = dns.TypeTXT
		hdr.Class = dns.ClassINET
		hdr.Name = name
	}
}

func (s *Server) handleServerInfo(m *dns.Msg) {
	si := &ServerInfo{
		BaseURL:  s.baseAddress,
		Features: []string{"FOLDER", "HREF", "TXT"},
	}

	m.Answer = append(m.Answer, &dns.TXT{
		Txt: []string{si.Encode()},
	})

	return
}

func (s *Server) handleList(m *dns.Msg, page int, path []string) {
	dir := s.vfs.GetDir(path...)
	if dir == nil {
		return
	}

	pages := getPages(dir.GetItems())
	p := &FolderPage{
		Items: pages[page],
	}

	m.Answer = append(m.Answer, &dns.TXT{
		Txt: []string{p.Encode()},
	})

	return
}

func (s *Server) handleFolderInfo(m *dns.Msg, path []string) {
	dir := s.vfs.GetDir(path...)
	if dir == nil {
		return
	}

	pages := getPages(dir.GetItems())

	si := &FolderInfo{
		Pages: len(pages),
		Items: len(dir.GetItems()),
	}

	m.Answer = append(m.Answer, &dns.TXT{
		Txt: []string{si.Encode()},
	})

	for i := 0; i < si.Pages; i++ {
		p := &FolderPage{
			Items: pages[i],
		}

		m.Extra = append(m.Extra, &dns.TXT{
			Txt: []string{p.Encode()},
		})
	}
}

func (s *Server) handleItem(m *dns.Msg, id string, path []string) {
	file := s.vfs.GetFile(append([]string{id}, path...)...)

	if file == nil {
		return
	}

	m.Answer = append(m.Answer, &dns.TXT{
		Txt: []string{file.GetContent()},
	})
}

func getPages(items []vfs.Item) [][]Item {
	var pages [][]Item
	var page []Item

	var l int
	newPage := false
	for _, it := range items {
		var ci Item
		if it.IsDir() {
			ci = &Dir{
				Name: it.GetName(),
				ID:   it.GetID().ItemID(),
			}
		} else {
			ci = &File{
				Name: it.GetName(),
				ID:   it.GetID().ItemID(),
			}
		}

		il := len(ci.String())
		if il > 200 {
			log.Fatal("file name too long:", ci.String())
		}

		l += il
		if l > 240 {
			newPage = true
		}

		if newPage {
			pages = append(pages, page)
			page = nil
			newPage = false
		}

		page = append(page, ci)
	}

	if len(page) != 0 {
		pages = append(pages, page)
	}

	return pages
}
