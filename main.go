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
	Name, Path string
	IsDir      bool
	Children   []string
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
		//添加文件按钮
		widget.NewToolbarAction(theme.FileTextIcon(), func() { dialogAddFile(w, lc) }),
		widget.NewToolbarAction(theme.FolderNewIcon(), func() { dialogAddDir(w, lc) }),
		widget.NewToolbarAction(theme.DeleteIcon(), func() { dialogDelFile(w, lc) }),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
			err := saveFile(a, text)
			if err != nil {
				fmt.Printf("save file error: %v\n", err)
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

// dialogPwd 输入密码弹窗
func dialogPwd(w fyne.Window) {
	pwd := widget.NewPasswordEntry()
	pwd.Validator = validation.NewRegexp(`^[A-Za-z0-9_-]{6,}$`, "由至少6位的数字、字母、下划线组成")
	items := []*widget.FormItem{
		widget.NewFormItem("密码", pwd),
	}
	dialog.ShowForm("请输入密码", "确定", "取消", items, func(b bool) {
		if !b {
			return
		}
		password = pwd.Text
	}, w)
}

// initFileTreeData 初始化目录信息
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

// updateTreeNode 更新树节点
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

// dialogAddFile 添加文件弹窗
func dialogAddFile(w fyne.Window, tree *widget.Tree) {
	fileName := widget.NewEntry()
	//password.Validator = validation.NewRegexp(`^[A-Za-z0-9_-]{6,}$`, "由至少6位的数字、字母、下划线组成")
	items := []*widget.FormItem{
		widget.NewFormItem("文件名", fileName),
	}
	dialog.ShowForm("输入文件名", "确定", "取消", items, func(b bool) {
		if !b {
			return
		}
		//添加文件
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
			fmt.Printf("创建文件[%s]失败: %v\n", p, err)
			return
		}
		defer f.Close()

		err = updateTreeNode(dir)
		if err != nil {
			fmt.Printf("更新节点[%s]失败: %v\n", dir, err)
			return
		}
		tree.Refresh()
	}, w)
}

// dialogAddDir 添加文件夹弹窗
func dialogAddDir(w fyne.Window, tree *widget.Tree) {
	dirName := widget.NewEntry()
	items := []*widget.FormItem{
		widget.NewFormItem("文件名", dirName),
	}
	dialog.ShowForm("输入文件夹名", "确定", "取消", items, func(b bool) {
		if !b {
			return
		}
		//添加文件
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
			fmt.Printf("创建文件夹[%s]失败: %v\n", p, err)
			return
		}

		err = updateTreeNode(dir)
		if err != nil {
			fmt.Printf("更新文件夹[%s]失败: %v\n", dir, err)
			return
		}
		tree.Refresh()
	}, w)
}

// dialogDelFile 删除文件弹窗
func dialogDelFile(w fyne.Window, tree *widget.Tree) {
	cnf := dialog.NewConfirm("删除文件", "是否确定删除选中的文件或文件夹？", func(b bool) {
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
				fmt.Printf("更新节点[%s]失败: %v\n", node, err)
				return
			}
			tree.Refresh()
		} else {
			err := errors.New("请选择一个文件或文件夹！")
			dialog.ShowError(err, w)
		}
	}, w)
	cnf.SetDismissText("取消")
	cnf.SetConfirmText("确定")
	cnf.Show()
}

// loadFileText 加载文件内容
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

	return nil
}

// saveFile 保存文件
func saveFile(a fyne.App, entry *widget.Entry) error {
	if currFile == nil || currFile.IsDir || currFile.Path == "" {
		return nil
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
