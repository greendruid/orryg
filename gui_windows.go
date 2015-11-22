package main

import (
	"image"
	"image/png"
	"os"

	"github.com/lxn/walk"
	"github.com/lxn/win"
)

const (
	wmShowUI = win.WM_USER + 1
)

type mainWindow struct {
	*walk.MainWindow
}

func newMainWindow() (*mainWindow, error) {
	mw, err := walk.NewMainWindow()
	if err != nil {
		return nil, err
	}

	w := &mainWindow{
		MainWindow: mw,
	}
	// NOTE(vincent): necessary so that our WndProc is called.
	err = walk.InitWrapperWindow(w)

	return w, err
}

func (m *mainWindow) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case wmShowUI:
		m.SetVisible(!m.Visible())
		return 0
	case win.WM_CLOSE:
		tray.notifyIcon.Dispose()
		// TODO(vincent): do we want to do something here ?
		fallthrough
	default:
		return m.FormBase.WndProc(hwnd, msg, wParam, lParam)
	}
}

// TODO(vincent): maybe we don't need this
type tabWidget struct {
	*walk.TabWidget
}

func newTabWidget(mw *mainWindow) (*tabWidget, error) {
	tw, err := walk.NewTabWidget(mw)
	if err != nil {
		return nil, err
	}

	w := &tabWidget{
		TabWidget: tw,
	}
	err = walk.InitWrapperWindow(w)

	return w, err
}

func (w *tabWidget) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	return w.TabWidget.WndProc(hwnd, msg, wParam, lParam)
}

type trayIcon struct {
	im         image.Image
	icon       *walk.Icon
	notifyIcon *walk.NotifyIcon

	stopAction                 *walk.Action
	stopActionTriggeredHandler int
	mouseDownHandler           int

	err error
}

func (i *trayIcon) loadImage() {
	if i.err != nil {
		return
	}

	var f *os.File
	f, i.err = os.Open("./cancel.png")
	if i.err != nil {
		return
	}
	defer f.Close()

	i.im, i.err = png.Decode(f)
	if i.err != nil {
		return
	}
}

func (i *trayIcon) makeIcon() {
	if i.err != nil {
		return
	}
	i.icon, i.err = walk.NewIconFromImage(i.im)
}

func (i *trayIcon) makeNotifyIcon() {
	if i.err != nil {
		return
	}
	i.notifyIcon, i.err = walk.NewNotifyIcon()
}

func (i *trayIcon) setIcon() {
	if i.err != nil {
		return
	}
	i.err = i.notifyIcon.SetIcon(i.icon)
}

func (i *trayIcon) setMenu() {
	if i.err != nil {
		return
	}
	menu := i.notifyIcon.ContextMenu()

	{
		i.stopAction = walk.NewAction()
		if i.err = i.stopAction.SetText("Quit"); i.err != nil {
			return
		}
		i.stopActionTriggeredHandler = i.stopAction.Triggered().Attach(func() {
			i.notifyIcon.Dispose()
			win.PostMessage(mw.Handle(), win.WM_CLOSE, 0, 0)
		})
		i.err = menu.Actions().Add(i.stopAction)
	}
}

func (i *trayIcon) setVisible(v bool) {
	if i.err != nil {
		return
	}
	i.err = i.notifyIcon.SetVisible(v)
}

func (i *trayIcon) attachMouseDownHandler() {
	if i.err != nil {
		return
	}
	i.mouseDownHandler = i.notifyIcon.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button == walk.LeftButton {
			win.PostMessage(mw.Handle(), wmShowUI, 0, 0)
		}
	})
}

func (i *trayIcon) init() error {
	i.loadImage()
	i.makeNotifyIcon()
	i.attachMouseDownHandler()
	i.makeIcon()
	i.setIcon()
	i.setMenu()
	i.setVisible(true)

	return i.err
}
