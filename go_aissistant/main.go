package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/fighthorse/aicode/go_aissistant/config"
	"github.com/fighthorse/aicode/go_aissistant/core/ai_model"
	"github.com/fighthorse/aicode/go_aissistant/core/knowledgebase"
	"github.com/fighthorse/aicode/go_aissistant/core/storage"
	"github.com/fighthorse/aicode/go_aissistant/core/websearch"
	"github.com/fighthorse/aicode/go_aissistant/gui"
	"time"

	"fyne.io/fyne/v2/app"
)

func main() {
	// 加载配置
	cc, err := config.LoadConfig("./config/app.json")
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		panic(err)
	}
	fmt.Println("配置加载成功")

	// 使用命令行参数选择启动模式
	mode := flag.String("mode", "gui", "选择启动模式: gui 或 cli")
	flag.Parse()

	fmt.Printf("启动模式: %s\n", *mode)
	switch *mode {
	case "gui":
		runGUI(cc)
	case "cli":
		runCLI(cc)
	default:
		fmt.Println("无效的模式，请选择 'gui' 或 'cli'")
	}
}

func runGUI(cc *config.AppConfig) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("发生错误:", r)
		}
	}()
	// 创建GUI
	fyneApp := app.New()
	mainWin := gui.NewMainWindow(fyneApp, cc)

	// 运行应用
	mainWin.Show()
	fyneApp.Run()
}

func runCLI(cc *config.AppConfig) {
	// 加载配置
	cc, err := config.LoadConfig("./config/app.json")
	if err != nil {
		panic(err)
	}

	// 导入文件到知识库
	parser := knowledgebase.NewFileParser()
	content, _ := parser.ParseFile("document.pdf")
	doc := knowledgebase.Document{
		ID:   "doc1",
		Text: content,
		Metadata: map[string]interface{}{
			"source": "document.pdf",
			"date":   time.Now().Format("2006-01-02"),
		},
	}
	kb, _ := knowledgebase.NewKnowledgeBase(cc)
	kb.AddDocuments([]knowledgebase.Document{doc})

	// 处理用户查询
	userQuery := "Go语言的主要特性是什么？"

	// 知识库检索
	kbResults, _ := kb.Query(userQuery, 3)

	// 网络搜索
	searchClient := websearch.NewGoogleSearchClient("", "")
	webResults, _ := searchClient.Search(context.Background(), userQuery, 5)

	// 构建提示
	prompt := websearch.BuildPrompt(userQuery, kbResults, webResults)

	// 生成回答
	aiClient := ai_model.NewOllamaClient("")
	response, _ := aiClient.Generate(prompt, "llama2")

	// 保存记录
	sto, _ := storage.NewSQLiteStorage("")
	sto.SaveChatRecord(userQuery, response, "llama2")
}
