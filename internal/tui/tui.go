package tui

import (
	"fmt"
	"sync"

	"github.com/mirkobrombin/goup/internal/logger"
	"github.com/rivo/tview"
)

var (
	app         *tview.Application
	textViews   = make(map[string]*tview.TextView)
	enabled     bool
	initOnce    sync.Once
	updateQueue = make(chan func(), 100)
)

// InitTUI initializes the TUI application.
func InitTUI() {
	initOnce.Do(func() {
		app = tview.NewApplication()
		enabled = true
		go processUpdateQueue()
	})
}

// IsEnabled returns whether the TUI is enabled or not.
func IsEnabled() bool {
	return enabled
}

// SetupView sets up a text view for the given identifier.
func SetupView(identifier string) {
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	textView.SetBorder(true).SetTitle(identifier)
	textViews[identifier] = textView
}

// Run starts the TUI application.
func Run() {
	grid := tview.NewGrid().SetRows(0).SetColumns(0)
	row := 0

	for _, textView := range textViews {
		grid.AddItem(textView, row, 0, 1, 1, 0, 0, false)
		row++
	}

	app.SetRoot(grid, true)

	if err := app.Run(); err != nil {
		panic(err)
	}
}

// UpdateLog updates the TUI log for a given identifier.
func UpdateLog(identifier string, fields logger.Fields) {
	textView, ok := textViews[identifier]
	if !ok {
		return
	}

	domain, _ := fields["domain"].(string)
	method, _ := fields["method"].(string)
	urlStr, _ := fields["url"].(string)
	statusCode, _ := fields["status_code"].(int)
	durationSec, _ := fields["duration_sec"].(float64)

	logLine := fmt.Sprintf("[%s] %s %s %d (%.4fs)\n",
		domain, method, urlStr, statusCode, durationSec)

	updateQueue <- func() {
		fmt.Fprint(textView, logLine)
	}
}

func processUpdateQueue() {
	for update := range updateQueue {
		app.QueueUpdateDraw(update)
	}
}
