package vfs

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type File struct {
	name    string
	content string
	id      *ID
}

func (f *File) GetName() string {
	return f.name
}

func (f *File) GetID() *ID {
	return f.id
}

func (f *File) IsDir() bool {
	return false
}

func (f *File) GetContent() string {
	return f.content
}

type Item interface {
	GetName() string
	GetID() *ID
	IsDir() bool
}

type Dir struct {
	name  string
	id    *ID
	items []Item
}

func (d *Dir) GetName() string {
	return d.name
}

func (d *Dir) GetID() *ID {
	return d.id
}

func (d *Dir) GetItems() []Item {
	return d.items
}

func (d *Dir) IsDir() bool {
	return true
}

type VFS struct {
	rootDir   string
	itemCache map[string]*File
	dirCache  map[string]*Dir
}

func NewVFS(rootDir string) *VFS {
	rootDir = path.Clean(rootDir)

	v := &VFS{
		rootDir:   rootDir,
		itemCache: make(map[string]*File),
		dirCache:  make(map[string]*Dir),
	}

	if err := filepath.Walk(rootDir, v.walkDir); err != nil {
		log.Fatal(err)
	}

	return v
}

type ID struct {
	itemID  string
	pathIDs []string
}

func (id *ID) PathID() string {
	return strings.Join(id.pathIDs, ".")
}

func (id *ID) ItemID() string {
	return id.itemID
}

func (id *ID) String() string {
	if id.PathID() == "" {
		return id.itemID
	}

	return id.itemID + "." + id.PathID()
}

func (v *VFS) getPathID(path string) *ID {
	var ids []string

	path = strings.TrimPrefix(path, v.rootDir)
	dirs := strings.Split(path, string(os.PathSeparator))
	for _, d := range dirs {
		id := fmt.Sprintf("%x", md5.Sum([]byte(d)))
		if d == "" {
			id = "root"
		}

		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return nil
	}

	if len(ids) == 1 && ids[0] == "" {
		ids[0] = "root"
	}

	// Reverse the array
	for i, j := 0, len(ids)-1; i < j; i, j = i+1, j-1 {
		ids[i], ids[j] = ids[j], ids[i]
	}

	return &ID{
		itemID:  ids[0],
		pathIDs: ids[1:],
	}
}

func (v *VFS) walkDir(p string, info os.FileInfo, err error) error {
	id := v.getPathID(p)
	if id == nil {
		return nil
	}

	var item Item
	if info.IsDir() {
		item = &Dir{
			name: info.Name(),
			id:   id,
		}

		v.dirCache[id.String()] = item.(*Dir)
	} else {
		data, err := ioutil.ReadFile(p)
		if err != nil {
			return err
		}

		c := string(data)
		switch path.Ext(info.Name()) {
		case ".txt":
			c = "00 " + c
		case ".href":
			c = "01 " + c
		case ".torrent":
			torrent, err := newTorrent(data)
			if err != nil {
				log.Fatal(err)
			}

			c = "02 " + strings.ToUpper(torrent.HashInfoBytes().HexString())
		default:
			return fmt.Errorf("unkown format")
		}

		item = &File{
			name:    info.Name(),
			content: c,
			id:      id,
		}

		v.itemCache[item.GetID().String()] = item.(*File)
	}

	if id.PathID() == "" {
		return nil
	}

	if dir := v.dirCache[id.PathID()]; dir != nil {
		dir.items = append(dir.items, item)
	} else {
		log.Fatal("could not find parent:", id.PathID())
	}

	return nil
}

func (v *VFS) GetDir(path ...string) *Dir {
	path = append(path, "root")
	pathID := strings.Join(path, ".")

	return v.dirCache[pathID]
}

func (v *VFS) GetFile(path ...string) *File {
	path = append(path, "root")
	pathID := strings.Join(path, ".")

	return v.itemCache[pathID]
}
