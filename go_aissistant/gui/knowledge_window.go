package gui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/fighthorse/aicode/go_aissistant/core/knowledgebase"
	"path/filepath"
	"time"
)

type KnowledgeWindow struct {
	window     fyne.Window
	mainWindow *MainWindow
	selectedId int
	list       *widget.List
	documents  []knowledgebase.Document
}

func NewKnowledgeWindow(mw *MainWindow) *KnowledgeWindow {
	kw := &KnowledgeWindow{
		mainWindow: mw,
		window:     mw.app.NewWindow("知识库管理"),
	}

	kw.buildUI()
	kw.refreshDocuments()
	kw.window.Resize(fyne.NewSize(600, 400))
	return kw
}

func (kw *KnowledgeWindow) buildUI() {
	// 文档列表
	kw.list = widget.NewList(
		func() int { return len(kw.documents) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.DocumentIcon()),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			container := obj.(*fyne.Container)
			label := container.Objects[1].(*widget.Label)
			label.SetText(kw.documents[id].Metadata["source"].(string))
		},
	)
	kw.list.OnSelected = func(id widget.ListItemID) {
		kw.selectedId = id
	}
	// 工具栏
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.FileIcon(), kw.onAddFile),
		widget.NewToolbarAction(theme.DeleteIcon(), kw.onDeleteDocument),
		widget.NewToolbarAction(theme.ViewRefreshIcon(), kw.refreshDocuments),
	)

	// 布局
	content := container.NewBorder(
		toolbar,
		nil, nil, nil,
		kw.list,
	)

	kw.window.SetContent(content)
}

func (kw *KnowledgeWindow) onAddFile() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			dialog.ShowError(err, kw.window)
			return
		}
		kw.importFile(reader.URI().Path())
	}, kw.window)
}

func (kw *KnowledgeWindow) importFile(path string) {
	parser := knowledgebase.NewFileParser()
	content, err := parser.ParseFile(path)
	if err != nil {
		dialog.ShowError(err, kw.window)
		return
	}

	doc := knowledgebase.Document{
		ID:   generateDocID(),
		Text: content,
		Metadata: map[string]interface{}{
			"source": filepath.Base(path),
			"date":   time.Now().Format("2006-01-02"),
		},
	}

	if err := kw.mainWindow.knowledgeBase.AddDocuments([]knowledgebase.Document{doc}); err != nil {
		dialog.ShowError(err, kw.window)
		return
	}

	kw.refreshDocuments()
}

func (kw *KnowledgeWindow) onDeleteDocument() {
	selected := kw.selectedId
	if selected < 0 {
		dialog.ShowInformation("提示", "请选择一个文档进行删除", kw.window)
		return
	}

	doc := kw.documents[selected]

	if err := kw.mainWindow.knowledgeBase.DeleteDocument(doc.ID); err != nil {
		dialog.ShowError(err, kw.window)
		return
	}

	kw.refreshDocuments()
}

func (kw *KnowledgeWindow) refreshDocuments() {
	// 获取所有文档
	var err error
	kw.documents, err = kw.mainWindow.knowledgeBase.ListDocuments()
	if err != nil {
		dialog.ShowError(err, kw.window)
		return
	}

	kw.list.Refresh()
}

func generateDocID() string {
	// 简单的文档ID生成逻辑，可以根据需要改进
	return fmt.Sprintf("doc-%d", time.Now().UnixNano())
}
