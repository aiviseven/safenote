package theme

//go:generate fyne bundle -package theme -o simkai.go simkai.ttf
//go:generate fyne bundle -package theme -o simhei.go simhei.ttf

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"image/color"
)

type ITheme struct{}

func (*ITheme) Font(s fyne.TextStyle) fyne.Resource {
	//return resourceSimkaiTtf	//楷体
	return resourceSimheiTtf //黑体
}
func (*ITheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(n, v)
}

func (*ITheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (*ITheme) Size(n fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(n)
}
