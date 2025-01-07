package app

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"sync"

	"fyne.io/systray"
	"github.com/electricbubble/go-toast"
	"github.com/gorilla/websocket"
)

var (
	appOk = false
	wsUp  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	tray = &TrayItem{
		itemType: item_TYPE_SYSTRAY,
		menuItem: nil,
	}
	window = &UI{
		chrome: nil,
	}
	appReadyChan = make(chan struct{})
	appExitChan  = make(chan struct{})
	appName      string
	appReady     func()
	appTask      func(...any)
	webRoot      fs.FS
	trayReady    func()
	trayExit     func()
	bus          = make(map[string]func(*websocket.Conn, string, any))
	busLock      sync.RWMutex
)

type Ping struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

type Pong struct {
	Event string `json:"event"`
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Data  []any  `json:"data"`
}

func GetUI() *UI {
	return window
}

func OpenWindow(width, height int) {
	port := os.Getenv("APP_PORT")
	_ = window.Open(fmt.Sprintf("http://localhost:%s", port), "", width, height)
}

func GetTray() *TrayItem {
	return tray
}

func SetTrayIcon(appIcon []byte) {
	systray.SetIcon(appIcon)
}

func New(name string, onReady func()) {
	if appOk {
		panic("App can only be created once")
	}

	if onReady == nil {
		onReady = func() {
			InitTray(name)
			m := tray.AddItem(nil, "打开主界面")
			m.OnClick(func(item *TrayItem) {
				OpenWindow(1200, 800)
			})
			tray.AddSeparator()
			q := tray.AddItem(nil, "退出")
			q.OnClick(func(item *TrayItem) {
				Notify("已退出")
				Quit()
			})
		}
	}

	appOk = true
	appName = name
	trayReady = onReady
	trayExit = func() {
		_ = window.Close()
	}
}

func Run(args ...any) {
	go listenWs()
	<-appReadyChan
	if appReady != nil {
		appReady()
	}
	if appTask == nil {
		go func() {
			<-appExitChan
		}()
	} else {
		go appTask(args...)
	}
	systray.Run(trayReady, trayExit)
}

func Done() <-chan struct{} {
	return appExitChan
}

func Quit() {
	close(appExitChan)
	systray.Quit()
}

func Notify(message string) {
	_ = toast.Push(message, toast.WithTitle(appName), toast.WithAudio(toast.Default))
}

func OnReady(f func()) {
	appReady = f
}

func SetWebRoot(root fs.FS) {
	webRoot = root
}

func SetTask(f func(...any)) {
	appTask = f
}

func SetNoArgTask(f func()) {
	SetTask(func(...any) { f() })
}
