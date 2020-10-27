package proto

import (
	"fmt"
	"strconv"
	"strings"
)

type ServerInfo struct {
	BaseURL  string
	Features []string
}

// Decode decodes the ServerInfo response
func (s *ServerInfo) Decode(resp string) *ServerInfo {
	if s == nil {
		s = &ServerInfo{}
	}

	parts := strings.Split(resp, ";")
	for _, part := range parts {
		kv := strings.Split(part, " ")
		if len(kv) != 2 {
			continue
		}

		switch kv[0] {
		case "SRV":
			s.BaseURL = kv[1]
		case "FEAT":
			s.Features = strings.Split(kv[1], ",")
		}
	}

	return s
}

// Encode encodes a ServerInfo struct into a response
func (s *ServerInfo) Encode() string {
	r := []string{
		fmt.Sprintf("SRV %s", s.BaseURL),
		fmt.Sprintf("FEAT %s", strings.Join(s.Features, ",")),
	}

	return strings.Join(r, ";")
}

type FolderInfo struct {
	Pages int
	Items int
}

// Decode decodes the FolderInfo response
func (f *FolderInfo) Decode(resp string) *FolderInfo {
	if f == nil {
		f = &FolderInfo{}
	}

	parts := strings.Split(resp, ";")
	for _, part := range parts {
		kv := strings.Split(part, " ")
		if len(kv) != 2 {
			continue
		}

		switch kv[0] {
		case "PAGES":
			f.Pages = mustInt(kv[1])
		case "ITEMS":
			f.Items = mustInt(kv[1])
		}
	}

	return f
}

// Encode encodes a FolderInfo struct into a response
func (f *FolderInfo) Encode() string {
	r := []string{
		fmt.Sprintf("PAGES %d", f.Pages),
		fmt.Sprintf("ITEMS %d", f.Items),
	}

	return strings.Join(r, ";")
}

func mustInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}

	return i
}

type FolderPage struct {
	Items []Item
}

type Item interface {
	fmt.Stringer
}

type Dir struct {
	Name string
	ID   string
}

func (d *Dir) String() string {
	return fmt.Sprintf("FD %s|%s", d.Name, d.ID)
}

type File struct {
	Name string
	ID   string
}

func (d *File) String() string {
	return fmt.Sprintf("IT %s|%s", d.Name, d.ID)
}

func DecodeItem(s string) Item {
	var item Item

	switch s[:2] {
	case "FD":
		kv := strings.Split(s[2:], "|")
		if len(kv) != 2 {
			return nil
		}

		item = &Dir{
			Name: kv[0],
			ID:   kv[1],
		}

	case "IT":
		kv := strings.Split(s[2:], "|")
		if len(kv) != 2 {
			return nil
		}

		item = &File{
			Name: kv[0],
			ID:   kv[1],
		}
	}

	return item
}

// Decode decodes the FolderPage response
func (f *FolderPage) Decode(resp string) *FolderPage {
	if f == nil {
		f = &FolderPage{}
	}

	parts := strings.Split(resp, ";")
	for _, part := range parts {
		item := DecodeItem(part)
		if item == nil {
			continue
		}

		f.Items = append(f.Items, item)
	}

	return f
}

// Encode encodes a FolderPage struct into a response
func (f *FolderPage) Encode() string {
	var r []string
	for _, item := range f.Items {
		r = append(r, item.String())
	}

	return strings.Join(r, ";")
}
