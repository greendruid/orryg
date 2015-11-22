package main

import (
	"image"
	"image/png"
	"os"

	"github.com/lxn/walk"
	"github.com/lxn/win"
)

var (
	mw *mainWindow
	tw *tabWidget

	dirTabPage *walk.TabPage
	dirTable   *walk.TableView
	dirModel   *directoriesModel

	copiersTabPage *walk.TabPage
	copiersListBox *walk.ListBox

	tray trayIcon
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

func buildUI() (err error) {
	{
		mw, err = newMainWindow()
		if err != nil {
			logger.Printf("unable to create new main window. err=%v", err)
			return
		}
		mw.SetSize(walk.Size{Width: 800, Height: 480})
		mw.SetLayout(walk.NewHBoxLayout())
	}

	{
		if tw, err = newTabWidget(mw); err != nil {
			logger.Printf("unable to create new tab widget. err=%v", err)
			return
		}
		tw.SetVisible(true)

		// Make the tabs
		//

		pages := tw.Pages()

		// Directories table
		{
			if dirTabPage, err = walk.NewTabPage(); err != nil {
				logger.Printf("unable to create directories tab page. err=%v", err)
				return
			}
			dirTabPage.SetTitle("Directories")
			dirTabPage.SetLayout(walk.NewHBoxLayout())
			pages.Add(dirTabPage)

			dirTable, err = walk.NewTableView(dirTabPage)
			if err != nil {
				logger.Printf("unable to create directories list box. err=%v", err)
			}
			dirTable.SetLastColumnStretched(true)

			archiveNameColumn := walk.NewTableViewColumn()
			archiveNameColumn.SetName("ArchiveName")
			archiveNameColumn.SetTitle("Archive name")
			archiveNameColumn.SetWidth(140)
			dirTable.Columns().Add(archiveNameColumn)

			originalPathColumn := walk.NewTableViewColumn()
			originalPathColumn.SetName("OriginalPath")
			originalPathColumn.SetTitle("Path")
			originalPathColumn.SetWidth(140)
			dirTable.Columns().Add(originalPathColumn)

			frequencyColumn := walk.NewTableViewColumn()
			frequencyColumn.SetName("Frequency")
			frequencyColumn.SetTitle("Frequency")
			frequencyColumn.SetWidth(100)
			dirTable.Columns().Add(frequencyColumn)

			maxBackupsColumn := walk.NewTableViewColumn()
			maxBackupsColumn.SetName("MaxBackups")
			maxBackupsColumn.SetTitle("Max nb. of backups")
			maxBackupsColumn.SetWidth(120)
			dirTable.Columns().Add(maxBackupsColumn)

			maxBackupAgeColumn := walk.NewTableViewColumn()
			maxBackupAgeColumn.SetName("MaxBackupAge")
			maxBackupAgeColumn.SetTitle("Max age of backups")
			maxBackupAgeColumn.SetWidth(120)
			dirTable.Columns().Add(maxBackupAgeColumn)

			lastUpdatedColumn := walk.NewTableViewColumn()
			lastUpdatedColumn.SetName("LastUpdated")
			lastUpdatedColumn.SetTitle("Last updated")
			dirTable.Columns().Add(lastUpdatedColumn)

			if dirModel, err = newDirectoriesModel(); err != nil {
				logger.Printf("unable to create directories model. err=%v", err)
				return
			}

			dirTable.SetModel(dirModel)
		}

		// Copiers table
		{
			if copiersTabPage, err = walk.NewTabPage(); err != nil {
				logger.Printf("unable to create copiers tab page. err=%v", err)
				return
			}
			copiersTabPage.SetTitle("Copiers")
			copiersTabPage.SetLayout(walk.NewHBoxLayout())
			pages.Add(copiersTabPage)
		}

		{
			copiersListBox, err = walk.NewListBox(copiersTabPage)
			if err != nil {
				logger.Printf("unable to create copiers list box. err=%v", err)
			}
			copiersListBox.SetModel([]string{"foo copier", "bar copier"})
		}

		tw.SetCurrentIndex(0)
		tw.CurrentIndexChanged().Attach(func() {
			logger.Printf("current index: %v", tw.CurrentIndex())
		})
	}

	if err = tray.init(); err != nil {
		logger.Printf("unable to initialize tray icon. err=%v", err)
		return
	}

	mw.SetVisible(true)

	return
}
