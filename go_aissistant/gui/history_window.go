package gui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"os"
	"strconv"
	"time"
)

// 自定义日期选择器
type DatePicker struct {
	year  *widget.Entry
	month *widget.Entry
	day   *widget.Entry
}

// 创建日期选择器
func NewDatePicker(selectedDate time.Time) *DatePicker {
	year := widget.NewEntry()
	year.SetPlaceHolder("年份")
	year.Validator = func(text string) error {
		if _, err := strconv.Atoi(text); err != nil {
			return fmt.Errorf("无效的年份格式")
		}
		return nil
	}

	month := widget.NewEntry()
	month.SetPlaceHolder("月份")
	month.Validator = func(text string) error {
		monthInt, err := strconv.Atoi(text)
		if err != nil {
			return fmt.Errorf("无效的月份格式")
		}
		if monthInt < 1 || monthInt > 12 {
			return fmt.Errorf("月份应在1到12之间")
		}
		return nil
	}

	day := widget.NewEntry()
	day.SetPlaceHolder("日期")
	day.Validator = func(text string) error {
		dayInt, err := strconv.Atoi(text)
		if err != nil {
			return fmt.Errorf("无效的日期格式")
		}
		if dayInt < 1 || dayInt > 31 {
			return fmt.Errorf("日期应在1到31之间")
		}
		return nil
	}

	year.SetText(strconv.Itoa(selectedDate.Year()))
	month.SetText(fmt.Sprintf("%02d", selectedDate.Month()))
	day.SetText(fmt.Sprintf("%02d", selectedDate.Day()))

	return &DatePicker{
		year:  year,
		month: month,
		day:   day,
	}
}

// 获取选中的日期
func (dp *DatePicker) SelectedDate() (time.Time, error) {
	year, err := strconv.Atoi(dp.year.Text)
	if err != nil {
		return time.Time{}, fmt.Errorf("无效的年份格式: %v", err)
	}

	month, err := strconv.Atoi(dp.month.Text)
	if err != nil {
		return time.Time{}, fmt.Errorf("无效的月份格式: %v", err)
	}

	day, err := strconv.Atoi(dp.day.Text)
	if err != nil {
		return time.Time{}, fmt.Errorf("无效的日期格式: %v", err)
	}

	selectedDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
	if selectedDate.Year() != year || selectedDate.Month() != time.Month(month) || selectedDate.Day() != day {
		return time.Time{}, fmt.Errorf("无效的日期")
	}

	return selectedDate, nil
}

type HistoryWindow struct {
	mainWindow *MainWindow
	window     fyne.Window
	list       *widget.List
	entries    []map[string]interface{}
	startDate  *DatePicker
	endDate    *DatePicker
}

func NewHistoryWindow(mw *MainWindow) *HistoryWindow {
	hw := &HistoryWindow{
		mainWindow: mw,
		window:     mw.app.NewWindow("完整对话历史"),
		entries:    []map[string]interface{}{},
	}

	hw.buildUI()
	hw.refreshData(time.Time{}, time.Now())
	hw.window.Resize(fyne.NewSize(800, 600))
	return hw
}

func (hw *HistoryWindow) buildUI() {
	// 时间范围选择器
	hw.startDate = NewDatePicker(time.Now().AddDate(0, 0, -1))
	hw.endDate = NewDatePicker(time.Now())

	filterBtn := widget.NewButton("筛选", func() {
		start, err := hw.startDate.SelectedDate()
		if err != nil {
			dialog.ShowError(err, hw.window)
			return
		}

		end, err := hw.endDate.SelectedDate()
		if err != nil {
			dialog.ShowError(err, hw.window)
			return
		}

		hw.refreshData(start, end)
	})

	// 历史记录列表
	hw.list = widget.NewList(
		func() int { return len(hw.entries) },
		func() fyne.CanvasObject {
			return container.NewVBox(
				widget.NewLabel(""),
				widget.NewLabel(""),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			entry := hw.entries[id]
			container := obj.(*fyne.Container)
			container.Objects[0].(*widget.Label).SetText("问题: " + entry["query"].(string))
			container.Objects[1].(*widget.Label).SetText("回答: " + entry["response"].(string))
			container.Objects[2].(*widget.Label).SetText("时间: " + entry["timestamp"].(string))
		},
	)

	// 右键菜单
	hw.list.OnSelected = func(id widget.ListItemID) {
		menu := fyne.NewMenu("操作",
			fyne.NewMenuItem("删除", func() {
				hw.deleteEntry(id)
			}),
			fyne.NewMenuItem("导出", func() {
				hw.exportEntry(id)
			}),
		)
		widget.ShowPopUpMenuAtPosition(menu, hw.window.Canvas(), fyne.CurrentApp().Driver().AbsolutePositionForObject(hw.list))
	}

	// 布局
	toolbar := container.NewHBox(
		widget.NewLabel("开始时间:"),
		hw.startDate.year,
		hw.startDate.month,
		hw.startDate.day,
		widget.NewLabel("结束时间:"),
		hw.endDate.year,
		hw.endDate.month,
		hw.endDate.day,
		filterBtn,
	)

	hw.window.SetContent(container.NewBorder(
		toolbar,
		nil, nil, nil,
		hw.list,
	))
}

func (hw *HistoryWindow) refreshData(start, end time.Time) {
	entries, err := hw.mainWindow.storage.GetHistoryByTimeRange(start, end)
	if err != nil {
		dialog.ShowError(err, hw.window)
		return
	}

	hw.entries = entries
	hw.list.Refresh()
}

func (hw *HistoryWindow) deleteEntry(id widget.ListItemID) {
	if id < 0 || int(id) >= len(hw.entries) {
		dialog.ShowInformation("提示", "请选择一个有效的条目进行删除", hw.window)
		return
	}

	entryID := hw.entries[id]["id"].(int)
	if err := hw.mainWindow.storage.DeleteEntry(entryID); err != nil {
		dialog.ShowError(err, hw.window)
		return
	}

	hw.refreshData(time.Time{}, time.Now()) // 重新加载数据
}

func (hw *HistoryWindow) exportEntry(id widget.ListItemID) {
	if id < 0 || int(id) >= len(hw.entries) {
		dialog.ShowInformation("提示", "请选择一个有效的条目进行导出", hw.window)
		return
	}

	entry := hw.entries[id]
	query := entry["query"].(string)
	response := entry["response"].(string)
	timestamp := entry["timestamp"].(string)

	// 创建一个临时文件，并将问题、回答和时间写入文件中
	tempFile, err := os.CreateTemp("", "exported_entry_*.txt")
	if err != nil {
		dialog.ShowError(err, hw.window)
		return
	}
	defer tempFile.Close()

	_, err = tempFile.WriteString("问题: " + query + "\n")
	if err != nil {
		dialog.ShowError(err, hw.window)
		return
	}

	_, err = tempFile.WriteString("回答: " + response + "\n")
	if err != nil {
		dialog.ShowError(err, hw.window)
		return
	}

	_, err = tempFile.WriteString("时间: " + timestamp + "\n")
	if err != nil {
		dialog.ShowError(err, hw.window)
		return
	}

	// 获取文件路径
	filePath := tempFile.Name()

	// 提示用户导出成功
	dialog.ShowInformation("导出成功", fmt.Sprintf("文件已导出到: %s", filePath), hw.window)
}
