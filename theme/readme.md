软件字体可以自己根据喜好，下载对应的字体ttf文件生成

`fyne bundle -package theme -o simkai.go simkai.ttf`

`fyne bundle -package theme -o simhei.go simhei.ttf`

根据ttf文件生成字体资源go文件后，需要替换theme.go中Font函数的返回值为自己生成的资源变量名