package gui

import (
	"context"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/fighthorse/aicode/go_aissistant/config"
	"github.com/fighthorse/aicode/go_aissistant/core/ai_model"
	"github.com/fighthorse/aicode/go_aissistant/core/knowledgebase"
	"github.com/fighthorse/aicode/go_aissistant/core/storage"
	"github.com/fighthorse/aicode/go_aissistant/core/websearch"
	"log"
	"strings"
	"time"
)

type MainWindow struct {
	app    fyne.App
	window fyne.Window
	config *config.AppConfig

	// 核心组件
	aiClient      *ai_model.OllamaClient
	knowledgeBase knowledgebase.KnowledgeBaseI
	storage       *storage.SQLiteStorage
	searchClient  websearch.WebSearchI

	// UI组件
	inputEntry  *widget.Entry
	outputText  *widget.Label
	statusLabel *widget.Label
	modelSelect *widget.Select
	historyList *widget.List
	progressBar *widget.ProgressBarInfinite
}

func NewMainWindow(app fyne.App, config *config.AppConfig, kknowledgeBase knowledgebase.KnowledgeBaseI) *MainWindow {
	mw := &MainWindow{
		app:           app,
		config:        config,
		window:        app.NewWindow("GoAIssistant"),
		inputEntry:    widget.NewEntry(),
		outputText:    widget.NewLabel(""),
		statusLabel:   widget.NewLabel("就绪"),
		progressBar:   widget.NewProgressBarInfinite(),
		knowledgeBase: kknowledgeBase,
	}

	mw.outputText.TextStyle = fyne.TextStyle{
		Monospace: true,
	}

	// 初始化核心组件
	mw.initializeComponents()
	mw.buildUI()
	//mw.setupShortcuts()
	mw.window.Resize(fyne.NewSize(800, 600))

	// 加载初始数据
	go mw.loadInitialData()

	return mw
}

func (mw *MainWindow) initializeComponents() {
	var err error

	// 初始化AI客户端
	mw.aiClient = ai_model.NewOllamaClient(mw.config.OllamaURL)

	// 初始化存储
	mw.storage, err = storage.NewSQLiteStorage(mw.config.SQLitePath)
	if err != nil {
		dialog.ShowError(err, mw.window)
	}

	// 初始化搜索客户端
	if mw.config.GoogleAPIKey != "" {
		mw.searchClient = websearch.NewGoogleSearchClient(
			mw.config.GoogleAPIKey,
			mw.config.GoogleCX,
		)
	} else if mw.config.BingAPIKey != "" {
		mw.searchClient = websearch.NewBingSearchClient(
			mw.config.BingAPIKey,
		)
	} else {
		mw.searchClient = websearch.NewBaseClient()
	}

}

func (mw *MainWindow) buildUI() {
	// 构建工具栏
	toolbar := mw.buildToolbar()

	// 构建模型选择器
	mw.modelSelect = widget.NewSelect([]string{}, func(s string) {
		mw.config.DefaultModel = s
	})
	go mw.refreshModelList()

	// 构建历史记录列表
	mw.historyList = widget.NewList(
		func() int { return 0 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {},
	)
	mw.refreshHistory()

	// 主布局
	leftPanel := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("历史记录"),
			container.NewHBox(
				widget.NewButtonWithIcon("刷新", theme.ViewRefreshIcon(), mw.refreshHistory),
				widget.NewButtonWithIcon("清除", theme.DeleteIcon(), mw.clearHistory),
			),
		),
		nil, nil, nil,
		mw.historyList,
	)

	// 使用 widget.NewLabelWrap 创建自动换行的标签
	// 修改输出区域的创建方式
	mw.outputText = widget.NewLabel("")
	// 修改后
	mw.outputText.Wrapping = fyne.TextWrapWord // 启用自动换行

	rightPanel := container.NewBorder(
		container.NewVBox(
			toolbar,
			container.NewHBox(
				widget.NewLabel("选择模型:"),
				mw.modelSelect,
				layout.NewSpacer(),
				mw.statusLabel,
			),
			mw.inputEntry,
			container.NewHBox(
				widget.NewButtonWithIcon("发送", theme.MailSendIcon(), mw.onSend),
				mw.progressBar, // 确保 progressBar 在这里
			),
		),
		nil, nil, nil,
		container.NewVScroll(mw.outputText), // 使用垂直滚动容器
		//outputScroll, // 使用垂直滚动容器
	)

	split := container.NewHSplit(leftPanel, rightPanel)
	split.Offset = 0.2

	mw.window.SetContent(split)

	mw.progressBar.Resize(fyne.NewSize(200, 20)) // 设置一个合理的大小
	mw.progressBar.Hide()                        // 确保 progressBar 默认是隐藏的
}

func (mw *MainWindow) buildToolbar() *widget.Toolbar {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.FolderOpenIcon(), mw.onImportFile),
		widget.NewToolbarAction(theme.StorageIcon(), mw.showKnowledgeManager),
		widget.NewToolbarAction(theme.HistoryIcon(), mw.showFullHistory),
		widget.NewToolbarAction(theme.SettingsIcon(), mw.showSettings),
	)
}

func (mw *MainWindow) onSend() {
	question := strings.TrimSpace(mw.inputEntry.Text)
	if question == "" {
		return
	}

	mw.progressBar.Show()
	mw.statusLabel.SetText("处理中...")
	mw.progressBar.Refresh()
	mw.statusLabel.Refresh()

	go func() {
		// 执行查询流程
		response, err := mw.processQuery(question, mw.modelSelect.Selected)
		if err != nil {
			dialog.ShowError(err, mw.window)
			return
		}

		// 更新UI
		mw.app.SendNotification(fyne.NewNotification("收到回复", "点击查看"))
		mw.outputText.SetText(response)
		mw.inputEntry.SetText("")
		mw.inputEntry.Refresh()

		// 刷新历史记录
		mw.refreshHistory()

		// loading end
		mw.progressBar.Hide()
		mw.statusLabel.SetText("就绪")
		mw.progressBar.Refresh()
		mw.statusLabel.Refresh()
	}()
}

func (mw *MainWindow) processQuery(question string, model string) (string, error) {
	// 步骤1：查询知识库
	kbResults, _ := mw.knowledgeBase.Query(question, 3)

	var kbStr []string
	for _, doc := range kbResults {
		kbStr = append(kbStr, doc.Text)
	}
	fmt.Println("knowledgeBase Query")
	// 步骤2：执行网络搜索
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	webResults, _ := mw.searchClient.Search(ctx, question, 5)

	fmt.Println("searchClient Search")
	// 步骤3：构建提示
	prompt := question
	if len(kbStr) > 0 && len(webResults) > 0 {
		prompt = buildPrompt(question, kbStr, webResults)
	}

	// 步骤4：调用AI生成
	if model == "" {
		model = mw.config.DefaultModel
	}
	fmt.Println("aiClient Generate", prompt, " ", model)
	response, err := mw.aiClient.Generate(prompt, model)
	if err != nil {
		return err.Error(), nil
	}

	// 步骤5：保存记录
	_ = mw.storage.SaveChatRecord(question, response, mw.config.DefaultModel)
	log.Println("模型：", model, " 构建提示：", prompt, " 返回：", len(response))
	return response, nil
}

func (mw *MainWindow) refreshModelList() {
	models, err := mw.aiClient.ListLocalModels()
	if err != nil {
		dialog.ShowError(err, mw.window)
		return
	}

	mw.modelSelect.Options = models
	if len(models) > 0 {
		mw.modelSelect.SetSelected(mw.config.DefaultModel)
	}
}

func (mw *MainWindow) refreshHistory() {
	history, err := mw.storage.GetRecentHistory(20)
	if err != nil {
		return
	}

	mw.historyList.Length = func() int { return len(history) }
	mw.historyList.UpdateItem = func(id widget.ListItemID, obj fyne.CanvasObject) {
		label := obj.(*widget.Label)
		entry := history[id]
		label.SetText(entry["query"])
	}
	mw.historyList.OnSelected = func(id widget.ListItemID) {
		entry := history[id]
		mw.showHistoryDetail(entry)
	}
	mw.historyList.Refresh()
}

func (mw *MainWindow) loadInitialData() {
	if err := mw.storage.CleanOldRecords(mw.config.RetentionDays); err != nil {
		dialog.ShowError(err, mw.window)
	}
}

func (mw *MainWindow) showHistoryDetail(entry map[string]string) {
	dialog.ShowCustom("历史详情", "关闭", container.NewVBox(
		widget.NewLabel(fmt.Sprintf("问题：%s", entry["query"])),
		widget.NewLabel(fmt.Sprintf("回答：%s", entry["response"])),
		widget.NewLabel(fmt.Sprintf("模型：%s", entry["model"])),
	), mw.window)
	mw.refreshHistory()
}

func (mw *MainWindow) onImportFile() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
	}, mw.window)
}

func (mw *MainWindow) clearHistory() {
	dialog.ShowConfirm("确认", "确定要清除历史记录吗？", func(b bool) {
		if b {
			if err := mw.storage.ClearHistory(); err != nil {
				dialog.ShowError(err, mw.window)
			}
		}
	}, mw.window)
	mw.refreshHistory()
	mw.refreshModelList()
}

func (mw *MainWindow) showKnowledgeManager() {
	kw := NewKnowledgeWindow(mw)
	kw.window.Show()
}

func (mw *MainWindow) showFullHistory() {
	hw := NewHistoryWindow(mw)
	hw.window.Show()
}

func (mw *MainWindow) showSettings() {
	sw := NewSettingsWindow(mw)
	sw.window.Show()
	defer func() {
		mw.config.DefaultModel = sw.config.DefaultModel
		mw.refreshModelList()
	}()
}

func (mw *MainWindow) Show() {
	mw.window.Show()
	defer func() {
		mw.refreshHistory()
		mw.refreshModelList()
	}()
}

func buildPrompt(question string, kbResults []string, webResults []websearch.SearchResult) string {
	var builder strings.Builder

	// 知识库内容
	builder.WriteString("知识库参考内容：\n")
	for i, doc := range kbResults {
		builder.WriteString(fmt.Sprintf("[知识%d] %s\n", i+1, doc))
	}

	// 网络搜索结果
	builder.WriteString("\n网络搜索结果：\n")
	for i, result := range webResults {
		builder.WriteString(fmt.Sprintf("[网络%d] %s\n%s\n", i+1, result.Title, result.Snippet))
	}

	// 最终问题
	builder.WriteString(fmt.Sprintf("\n请根据以上信息回答：%s", question))

	return builder.String()
}
