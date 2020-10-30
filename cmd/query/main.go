package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/fionera/tttnsd/proto"
)

type reference struct {
	item        proto.Item
	path        []string
	itemContent string
}

func (r *reference) ChildPath() []string {
	return append([]string{r.item.GetID()}, r.path...)
}

func main() {
	flag.Parse()
	domain := flag.Arg(0)

	if domain == "" {
		log.Fatal("Please provide the Domain")
	}

	c, err := proto.NewClient(domain)
	if err != nil {
		log.Fatal(err)
	}

	app := tview.NewApplication()

	root := tview.NewTreeNode(domain).
		SetColor(tcell.ColorRed)
	tree := tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)

	addPath := func(target *tview.TreeNode, path ...string) {
		items, err := c.GetDir(path...)
		if err != nil {
			log.Println(path)
			target.SetColor(tcell.ColorDarkRed)
			return
		}

		for _, item := range items {
			r := reference{
				item: item,
				path: path,
			}

			node := tview.NewTreeNode(item.GetName()).
				SetReference(&r).
				SetSelectable(true)

			if !r.item.IsDir() {
				node.SetColor(tcell.ColorGreen)
			}

			target.AddChild(node)
		}
	}

	disp := func(target *tview.TreeNode, ref *reference) {
		if ref.itemContent == "" {
			data, err := c.GetFile(ref.item.GetID(), ref.path...)
			if err != nil {
				target.SetColor(tcell.ColorDarkRed)
				return
			}

			ref.itemContent = data
		}

		modal := tview.NewModal().
			SetText(fmt.Sprintf("%s\n\n%s", ref.item.GetName(), ref.itemContent[3:])).
			AddButtons([]string{"Close", "Save"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "Save" {
					cnt := ref.itemContent[3:]
					if filepath.Ext(ref.item.GetName()) == ".torrent" {
						magnet := metainfo.Magnet{
							InfoHash:    metainfo.NewHashFromHex(ref.itemContent),
							Trackers:    nil,
							DisplayName: ref.item.GetName(),
							Params:      nil,
						}

						cnt = magnet.String()
					}

					err := ioutil.WriteFile(ref.item.GetName(), []byte(cnt), 0775)
					if err != nil {
						log.Fatal(err)
					}
				}
				app.SetRoot(tree, false)
			})

		app.SetRoot(modal, false)
	}

	addPath(root)

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		r := node.GetReference()
		if r == nil {
			return
		}

		children := node.GetChildren()
		if len(children) == 0 {
			ref := r.(*reference)
			if !ref.item.IsDir() {
				disp(node, ref)
			} else {
				addPath(node, r.(*reference).ChildPath()...)
			}
		} else {
			node.SetExpanded(!node.IsExpanded())
		}
	})

	if err := app.SetRoot(tree, true).EnableMouse(true).Run(); err != nil {
		log.Fatal(err)
	}
}
