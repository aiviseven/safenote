package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	it "safenote/theme"
)

type fileInfo struct {
	Name, Path, Content string
}

var (
	FileName = map[string][]string{
		"":  {"a", "b", "c", "d"},
		"a": {"aa", "ab", "ac", "ad"},
		"c": {"ca", "cb", "cc"},
	}
	FileContent = map[string]fileInfo{
		"a":  fileInfo{"a", "aa", ""},
		"aa": fileInfo{"aa", "a/aa", "测试aa"},
		"ab": fileInfo{"ab", "a/ab", "测试ab"},
		"ac": fileInfo{"ac", "a/ac", "测试ac"},
		"ad": fileInfo{"ad", "a/ad", "测试ad"},
		"b":  fileInfo{"b", "b", ""},
		"c":  fileInfo{"c", "c", ""},
		"ca": fileInfo{"ca", "c/ca", "测试ca"},
		"cb": fileInfo{"cb", "c/cb", "测试cb"},
		"cc": fileInfo{"cc", "c/cc", "测试cc"},
		"d":  fileInfo{"d", "d", "测试d"},
	}
)

func main() {
	a := app.New()
	a.Settings().SetTheme(&it.ITheme{})
	w := a.NewWindow("SafeNote")
	w.SetMaster()

	lt := &widget.Tree{
		ChildUIDs: func(uid string) []string {
			return FileName[uid]
		},
		IsBranch: func(uid string) bool {
			children, ok := FileName[uid]

			return ok && len(children) > 0
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Collection Widgets")
		},
		UpdateNode: func(uid string, branch bool, obj fyne.CanvasObject) {
			f, ok := FileContent[uid]
			if !ok {
				fyne.LogError("Missing tutorial panel: "+uid, nil)
				return
			}
			obj.(*widget.Label).SetText(f.Name)
		},
		OnSelected: func(uid string) {
			if _, ok := FileContent[uid]; ok {
				a.Preferences().SetString("currFile", uid)

			}
		},
	}
	lb := container.NewBorder(nil, nil, nil, nil, lt)

	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentCutIcon(), func() { fmt.Println("Cut") }),
		widget.NewToolbarAction(theme.ContentCopyIcon(), func() { fmt.Println("Copy") }),
		widget.NewToolbarAction(theme.ContentPasteIcon(), func() { fmt.Println("Paste") }),
	)
	text := widget.NewMultiLineEntry()
	text.SetText("测试\naaa\nbbb\nccc\nddd\neee\nfff")
	text.Wrapping = fyne.TextTruncate
	rc := container.NewMax()
	rc.Objects = []fyne.CanvasObject{text}
	rc.Refresh()
	rb := container.NewBorder(toolbar, nil, nil, nil, rc)

	c := container.NewHSplit(lb, rb)
	c.Offset = 0.2
	w.SetContent(c)
	w.Resize(fyne.NewSize(640, 460))
	w.ShowAndRun()
}
