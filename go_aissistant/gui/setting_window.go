package gui

import (
	"fmt"
	"github.com/fighthorse/aicode/go_aissistant/config"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type SettingsWindow struct {
	window fyne.Window
	config *config.AppConfig
	form   *widget.Form
}

func NewSettingsWindow(mw *MainWindow) *SettingsWindow {
	sw := &SettingsWindow{
		window: mw.app.NewWindow("设置"),
		config: mw.config,
	}

	sw.buildUI()
	sw.window.Resize(fyne.NewSize(400, 300))
	return sw
}

func (sw *SettingsWindow) buildUI() {
	// 配置项输入字段
	ollamaURL := widget.NewEntry()
	ollamaURL.SetText(sw.config.OllamaURL)

	googleAPIKey := widget.NewPasswordEntry()
	googleAPIKey.SetText(sw.config.GoogleAPIKey)

	chromaPath := widget.NewEntry()
	chromaPath.SetText(sw.config.ChromaPath)

	historyLimit := widget.NewEntry()
	historyLimit.SetText(fmt.Sprintf("%d", sw.config.HistoryLimit))

	// 表单
	sw.form = &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Ollama地址", Widget: ollamaURL},
			{Text: "Google API密钥", Widget: googleAPIKey},
			{Text: "知识库存储路径", Widget: chromaPath},
			{Text: "历史记录保留条数", Widget: historyLimit},
		},
		OnSubmit: func() {
			newOllamaURL := ollamaURL.Text
			newGoogleAPIKey := googleAPIKey.Text
			newChromaPath := chromaPath.Text
			newHistoryLimit, err := strconv.Atoi(historyLimit.Text)
			if err != nil {
				dialog.ShowError(fmt.Errorf("历史记录保留条数必须是整数: %v", err), sw.window)
				return
			}

			// 更新配置
			sw.config.OllamaURL = newOllamaURL
			sw.config.GoogleAPIKey = newGoogleAPIKey
			sw.config.ChromaPath = newChromaPath
			sw.config.HistoryLimit = newHistoryLimit

			// 保存配置
			if err := config.SaveConfig(sw.config); err != nil {
				dialog.ShowError(fmt.Errorf("保存配置时出错: %v", err), sw.window)
				return
			}

			// 关闭窗口
			sw.window.Close()
		},
	}

	// 添加保存按钮
	saveButton := widget.NewButton("保存并关闭", sw.form.OnSubmit)

	// 设置窗口内容
	sw.window.SetContent(container.NewVBox(
		sw.form,
		saveButton,
	))
}
