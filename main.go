package main

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/mitchellh/go-homedir"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	it "safenote/theme"
	"strings"
)

type fileNode struct {
	Name, Path    string
	IsDir, LoadOk bool
	Children      []string
}

var appHomeDir, appDataDir string
var treeData map[string]*fileNode
var currFile *fileNode
var password string

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

	dialogPwd(w)

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
			if curr, ok := treeData[uid]; ok {
				currFile = curr
				err := loadFileText(curr, text)
				if err != nil {
					fmt.Printf("load file[%s] error: %v\n", uid, err)
				}
			}
		},
		OnBranchOpened: func(uid string) {
			err := updateTreeNode(uid)
			if err != nil {
				fmt.Printf("update %s error: %v\n", uid, err)
			}
		},
	}
	ltb := widget.NewToolbar(
		//??????????????????
		widget.NewToolbarAction(theme.FileTextIcon(), func() { dialogAddFile(w, lc) }),
		//?????????????????????
		widget.NewToolbarAction(theme.FolderNewIcon(), func() { dialogAddDir(w, lc) }),
		//????????????????????????
		widget.NewToolbarAction(theme.DeleteIcon(), func() { dialogDelFile(w, lc) }),
		//???????????????????????????
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
			err := saveFile(text)
			if err != nil {
				fmt.Printf("save file error: %v\n", err)
				dialog.ShowError(err, w)
			}
		}),
	)
	lb := container.NewBorder(ltb, nil, nil, nil, lc)

	c := container.NewHSplit(lb, rb)
	c.Offset = 0.2
	w.SetContent(c)
	w.Resize(fyne.NewSize(640, 460))
	w.ShowAndRun()
}

// dialogPwd ??????????????????
func dialogPwd(w fyne.Window) {
	pwd := widget.NewPasswordEntry()
	pwd.Validator = validation.NewRegexp(`^[A-Za-z0-9_-]{6,}$`, "?????????6???????????????????????????????????????")
	items := []*widget.FormItem{
		widget.NewFormItem("??????", pwd),
	}
	dialog.ShowForm("???????????????", "??????", "??????", items, func(b bool) {
		if !b {
			return
		}
		password = pwd.Text
	}, w)
}

// initFileTreeData ?????????????????????
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

// updateTreeNode ???????????????
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
		treeData[cp] = &fileNode{Name: f.Name(), Path: cp, IsDir: f.IsDir()}
	}

	if nodePath == appDataDir {
		treeData[""] = &fileNode{Name: rf.Name(), Path: nodePath, IsDir: rf.IsDir(), Children: paths}
	} else {
		treeData[nodePath] = &fileNode{Name: rf.Name(), Path: nodePath, IsDir: rf.IsDir(), Children: paths}
	}
	return nil
}

// dialogAddFile ??????????????????
func dialogAddFile(w fyne.Window, tree *widget.Tree) {
	fileName := widget.NewEntry()
	//password.Validator = validation.NewRegexp(`^[A-Za-z0-9_-]{6,}$`, "?????????6???????????????????????????????????????")
	items := []*widget.FormItem{
		widget.NewFormItem("?????????", fileName),
	}
	dialog.ShowForm("???????????????", "??????", "??????", items, func(b bool) {
		if !b {
			return
		}
		//????????????
		var dir string
		if currFile != nil && currFile.Path != "" {
			if currFile.IsDir {
				dir = currFile.Path
			} else {
				dir = currFile.Path[0:strings.LastIndex(currFile.Path, string(filepath.Separator))]
			}
		} else {
			dir = appDataDir
		}
		p := dir + string(filepath.Separator) + fileName.Text
		f, err := os.Create(p)
		if err != nil {
			fmt.Printf("????????????[%s]??????: %v\n", p, err)
			return
		}
		defer func() {
			err := f.Close()
			if err != nil {
				fmt.Printf("????????????[%s]??????: %v\n", p, err)
			}
		}()

		err = updateTreeNode(dir)
		if err != nil {
			fmt.Printf("????????????[%s]??????: %v\n", dir, err)
			return
		}
		tree.Refresh()
	}, w)
}

// dialogAddDir ?????????????????????
func dialogAddDir(w fyne.Window, tree *widget.Tree) {
	dirName := widget.NewEntry()
	items := []*widget.FormItem{
		widget.NewFormItem("?????????", dirName),
	}
	dialog.ShowForm("??????????????????", "??????", "??????", items, func(b bool) {
		if !b {
			return
		}
		//????????????
		var dir string
		if currFile != nil && currFile.Path != "" {
			if currFile.IsDir {
				dir = currFile.Path
			} else {
				dir = currFile.Path[0:strings.LastIndex(currFile.Path, string(filepath.Separator))]
			}
		} else {
			dir = appDataDir
		}
		p := dir + string(filepath.Separator) + dirName.Text
		_, err := mkdir(p)
		if err != nil {
			fmt.Printf("???????????????[%s]??????: %v\n", p, err)
			return
		}

		err = updateTreeNode(dir)
		if err != nil {
			fmt.Printf("???????????????[%s]??????: %v\n", dir, err)
			return
		}
		tree.Refresh()
	}, w)
}

// dialogDelFile ??????????????????
func dialogDelFile(w fyne.Window, tree *widget.Tree) {
	cnf := dialog.NewConfirm("????????????", "????????????????????????????????????????????????", func(b bool) {
		if !b {
			return
		}
		if currFile != nil && currFile.Path != "" {
			var err error
			if currFile.IsDir {
				err = os.RemoveAll(currFile.Path)
			} else {
				err = os.Remove(currFile.Path)
			}
			if err != nil {
				dialog.ShowError(err, w)
			}

			node := currFile.Path[0:strings.LastIndex(currFile.Path, string(filepath.Separator))]
			err = updateTreeNode(node)
			if err != nil {
				fmt.Printf("????????????[%s]??????: %v\n", node, err)
				return
			}
			currFile = nil
			tree.Refresh()
		} else {
			err := errors.New("?????????????????????????????????")
			dialog.ShowError(err, w)
		}
	}, w)
	cnf.SetDismissText("??????")
	cnf.SetConfirmText("??????")
	cnf.Show()
}

// loadFileText ??????????????????
func loadFileText(node *fileNode, entry *widget.Entry) error {
	var content string
	defer func() {
		entry.SetText(content)
		entry.Refresh()
	}()

	if !node.IsDir {
		data, err := ioutil.ReadFile(node.Path)
		if err != nil {
			return err
		}
		if data != nil && len(data) > 0 {
			decryptData, err := DecryptSHA256(data, password)
			if err != nil {
				return err
			}
			content = string(decryptData)
		}
	}
	node.LoadOk = true
	return nil
}

// saveFile ????????????
func saveFile(entry *widget.Entry) error {
	if currFile == nil || currFile.IsDir || currFile.Path == "" {
		return errors.New("????????????????????????????????????")
	}

	if !currFile.LoadOk {
		return errors.New("?????????????????????????????????")
	}

	fmt.Println("start save file: ", currFile.Path)
	dataStr := entry.Text
	encryptData, err := EncryptSHA256([]byte(dataStr), password)

	f, err := os.Create(currFile.Path)
	if err != nil {
		return err
	}
	defer func() {
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
		fmt.Println("end save file: ", currFile.Path)
	}()
	_, err = f.Write(encryptData)
	return err
}

// isExists ????????????????????????
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

// mkdir ????????????????????????????????????
func mkdir(dirPath string) (string, error) {
	if isExist, _ := isExists(dirPath); !isExist {
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return "", errors.New("?????????????????????")
		}
	}
	return dirPath, nil
}
