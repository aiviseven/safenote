package main

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/mitchellh/go-homedir"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	it "safenote/theme"
)

type fileNode struct {
	Name, Path string
	IsDir      bool
	Children   []string
}

const proKeyCurrFile = "currFile"

var treeData map[string]*fileNode
var appHomeDir, appDataDir string

func init() {
	dir, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("User home dir: %s\n", dir)

	appHomeDir = dir + string(filepath.Separator) + ".safenote"
	appDataDir = appHomeDir + string(filepath.Separator) + "data"
	fmt.Printf("App home dir: %s\n", appHomeDir)
	_, err = mkdir(appDataDir)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	initFileTreeData()

	a := app.NewWithID("SafeNote_v0.01")
	a.Settings().SetTheme(&it.ITheme{})
	w := a.NewWindow("SafeNote")
	w.SetMaster()

	text := widget.NewMultiLineEntry()
	text.Wrapping = fyne.TextTruncate
	rc := container.NewMax()
	rc.Objects = []fyne.CanvasObject{text}
	rc.Refresh()
	rb := container.NewBorder(nil, nil, nil, nil, rc)

	lc := &widget.Tree{
		ChildUIDs: func(uid string) []string {
			return treeData[uid].Children
		},
		IsBranch: func(uid string) bool {
			node, ok := treeData[uid]
			return ok && node.IsDir
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Collection Widgets")
		},
		UpdateNode: func(uid string, branch bool, obj fyne.CanvasObject) {
			f, ok := treeData[uid]
			if !ok {
				fyne.LogError("Missing tutorial panel: "+uid, nil)
				return
			}
			obj.(*widget.Label).SetText(f.Name)
		},
		OnSelected: func(uid string) {
			if _, ok := treeData[uid]; ok {
				a.Preferences().SetString(proKeyCurrFile, uid)
				loadFileText(uid, text)
			}
		},
		OnBranchOpened: func(uid string) {
			err := updateTreeNode(uid)
			if err != nil {
				fmt.Printf("update %s error: %v\n", uid, err)
			}
		},
	}
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentAddIcon(), func() { fmt.Println("add") }),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
			err := saveFile(a, text)
			if err != nil {
				fmt.Printf("save file error: %v\n", err)
			}
		}),
	)
	lb := container.NewBorder(toolbar, nil, nil, nil, lc)

	c := container.NewHSplit(lb, rb)
	c.Offset = 0.2
	w.SetContent(c)
	w.Resize(fyne.NewSize(640, 460))
	w.ShowAndRun()
}

func initFileTreeData() {
	_, err := mkdir(appDataDir)
	if err != nil {
		log.Fatal(err)
	}
	treeData = make(map[string]*fileNode)
	err = updateTreeNode(appDataDir)
	if err != nil {
		fmt.Printf("update %s error: %v\n", appDataDir, err)
	}
}

func updateTreeNode(nodePath string) error {
	rf, err := os.Stat(nodePath)
	if err != nil {
		return err
	}
	files, err := ioutil.ReadDir(nodePath)
	if err != nil {
		return err
	}
	var paths []string
	for _, f := range files {
		cp := nodePath + string(filepath.Separator) + f.Name()
		paths = append(paths, cp)
		treeData[cp] = &fileNode{f.Name(), cp, f.IsDir(), nil}
	}

	if nodePath == appDataDir {
		treeData[""] = &fileNode{rf.Name(), nodePath, rf.IsDir(), paths}
	} else {
		treeData[nodePath] = &fileNode{rf.Name(), nodePath, rf.IsDir(), paths}
	}
	return nil
}

// loadFileText 加载文件内容
func loadFileText(filePath string, entry *widget.Entry) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	entry.SetText(string(data))
	entry.Refresh()
	return nil
}

// saveFile 保存文件
func saveFile(a fyne.App, entry *widget.Entry) error {
	filePath := a.Preferences().String(proKeyCurrFile)
	dataStr := entry.Text

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() {
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	_, err = f.WriteString(dataStr)
	return err
}

// isExists 检查路径是否存在
func isExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// mkdir 创建目录，存在则直接返回
func mkdir(dirPath string) (string, error) {
	if isExist, _ := isExists(dirPath); !isExist {
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return "", errors.New("创建输出路径出错！")
		}
	}
	return dirPath, nil
}
