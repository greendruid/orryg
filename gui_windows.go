package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/lxn/walk"
	"github.com/lxn/win"
	"golang.org/x/sys/windows/registry"
)

var (
	mw *mainWindow
	tw *tabWidget

	dirTabPage *walk.TabPage
	dirTable   *walk.TableView
	dirModel   *directoriesModel

	copierTabPage *walk.TabPage
	copierTable   *walk.TableView
	copierModel   *copiersModel

	tray trayIcon
)

const (
	wmShowUI = win.WM_USER + iota + 1
	wmEnableAutoRun
	wmDisableAutoRun
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

// WndProc is where all the "business logic" occurs for the main window.
func (m *mainWindow) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case wmShowUI:
		if m.Visible() {
			m.SetVisible(false)
		} else {
			m.SetVisible(true)
			win.SetForegroundWindow(m.Handle())
			win.ShowWindow(m.Handle(), win.SW_RESTORE)
		}

		return 0

	case wmEnableAutoRun:
		err := enableAutoRun()
		// TODO(vincent): show a message box maybe ?
		if err != nil {
			logger.Printf("unable to change the autorun. err=%v", err)
		}

		return 0

	case wmDisableAutoRun:
		err := disableAutoRun()
		// TODO(vincent): show a message box maybe ?
		if err != nil {
			logger.Printf("unable to change the autorun. err=%v", err)
		}

		return 0

	case win.WM_SIZE:
		// 1 means minimized
		// https://msdn.microsoft.com/en-us/library/windows/desktop/ms632646(v=vs.85).aspx
		if wParam == 1 {
			win.PostMessage(m.Handle(), wmShowUI, 0, 0)
		}

	case win.WM_CLOSE:
		tray.notifyIcon.Dispose()
		// TODO(vincent): do we want to do something here ?
	}

	return m.FormBase.WndProc(hwnd, msg, wParam, lParam)
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

// TODO(vincent): refactor that shit
type trayIcon struct {
	im         image.Image
	icon       *walk.Icon
	notifyIcon *walk.NotifyIcon

	stopAction          *walk.Action
	enableAutoRunAction *walk.Action

	err error
}

func (i *trayIcon) loadImage() {
	if i.err != nil {
		return
	}

	i.im, i.err = png.Decode(bytes.NewReader(cancelPNG[:]))
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
		i.enableAutoRunAction = walk.NewAction()

		if i.err = i.enableAutoRunAction.SetText("Autorun with Windows"); i.err != nil {
			return
		}

		if i.err = i.enableAutoRunAction.SetCheckable(true); i.err != nil {
			return
		}

		var autoRunEnabled bool
		autoRunEnabled, i.err = isAutoRunEnabled()
		if i.err != nil {
			return
		}

		if i.err = i.enableAutoRunAction.SetChecked(autoRunEnabled); i.err != nil {
			return
		}

		i.enableAutoRunAction.Triggered().Attach(func() {
			switch i.enableAutoRunAction.Checked() {
			case true:
				win.PostMessage(mw.Handle(), wmDisableAutoRun, 0, 0)
				i.enableAutoRunAction.SetChecked(false)
			case false:
				win.PostMessage(mw.Handle(), wmEnableAutoRun, 0, 0)
				i.enableAutoRunAction.SetChecked(true)
			}
		})

		if i.err = menu.Actions().Add(i.enableAutoRunAction); i.err != nil {
			return
		}
	}

	menu.Actions().Add(walk.NewSeparatorAction())

	{
		i.stopAction = walk.NewAction()
		if i.err = i.stopAction.SetText("Quit"); i.err != nil {
			return
		}

		i.stopAction.Triggered().Attach(func() {
			i.notifyIcon.Dispose()
			win.PostMessage(mw.Handle(), win.WM_CLOSE, 0, 0)
		})

		if i.err = menu.Actions().Add(i.stopAction); i.err != nil {
			return
		}
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
	i.notifyIcon.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
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
			return fmt.Errorf("unable to create new main window. err=%v", err)
		}
		mw.SetSize(walk.Size{Width: 800, Height: 480})
		mw.SetLayout(walk.NewHBoxLayout())
	}

	{
		if tw, err = newTabWidget(mw); err != nil {
			return fmt.Errorf("unable to create new tab widget. err=%v", err)
		}
		tw.SetVisible(true)

		// Make the tabs
		//

		pages := tw.Pages()

		// Directories table
		{
			if dirTabPage, err = walk.NewTabPage(); err != nil {
				return fmt.Errorf("unable to create directories tab page. err=%v", err)
			}
			dirTabPage.SetTitle("Directories")
			dirTabPage.SetLayout(walk.NewHBoxLayout())
			pages.Add(dirTabPage)

			dirTable, err = walk.NewTableView(dirTabPage)
			if err != nil {
				return fmt.Errorf("unable to create directories list box. err=%v", err)
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
				return fmt.Errorf("unable to create directories model. err=%v", err)
			}
			go dirModel.monitorDirChanged(en.DirectoryChangedCh)

			dirTable.SetModel(dirModel)
		}

		// Copiers table
		{
			if copierTabPage, err = walk.NewTabPage(); err != nil {
				return fmt.Errorf("unable to create copiers tab page. err=%v", err)
			}
			copierTabPage.SetTitle("Copiers")
			copierTabPage.SetLayout(walk.NewHBoxLayout())
			pages.Add(copierTabPage)

			copierTable, err = walk.NewTableView(copierTabPage)
			if err != nil {
				return fmt.Errorf("unable to create directories list box. err=%v", err)
			}
			copierTable.SetLastColumnStretched(true)

			nameColumn := walk.NewTableViewColumn()
			nameColumn.SetName("Name")
			nameColumn.SetTitle("Name")
			nameColumn.SetWidth(140)
			copierTable.Columns().Add(nameColumn)

			userColumn := walk.NewTableViewColumn()
			userColumn.SetName("User")
			userColumn.SetTitle("User")
			userColumn.SetWidth(60)
			copierTable.Columns().Add(userColumn)

			hostColumn := walk.NewTableViewColumn()
			hostColumn.SetName("Host")
			hostColumn.SetTitle("Host")
			hostColumn.SetWidth(140)
			copierTable.Columns().Add(hostColumn)

			portColumn := walk.NewTableViewColumn()
			portColumn.SetName("Port")
			portColumn.SetTitle("Port")
			portColumn.SetWidth(50)
			copierTable.Columns().Add(portColumn)

			privateKeyFileColumn := walk.NewTableViewColumn()
			privateKeyFileColumn.SetName("PrivateKeyFile")
			privateKeyFileColumn.SetTitle("Private key file")
			privateKeyFileColumn.SetWidth(140)
			copierTable.Columns().Add(privateKeyFileColumn)

			backupsDirColumn := walk.NewTableViewColumn()
			backupsDirColumn.SetName("BackupsDir")
			backupsDirColumn.SetTitle("Backups directory")
			copierTable.Columns().Add(backupsDirColumn)

			if copierModel, err = newCopiersModel(); err != nil {
				return fmt.Errorf("unable to create directories model. err=%v", err)
			}

			copierTable.SetModel(copierModel)
		}

		tw.SetCurrentIndex(0)
		tw.CurrentIndexChanged().Attach(func() {
			// TODO(vincent): do we need to do something here ?
			logger.Printf("current index: %v", tw.CurrentIndex())
		})
	}

	return tray.init()
}

func isAutoRunEnabled() (bool, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE)
	if err != nil {
		return false, err
	}
	defer key.Close()

	s, _, err := key.GetStringValue("Orryg")
	if err == registry.ErrNotExist {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return s != "", nil
}

func enableAutoRun() error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer key.Close()

	exe, err := executablePath()
	if err != nil {
		return err
	}

	return key.SetStringValue("Orryg", `"`+exe+`" --minimized`)
}

func disableAutoRun() error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer key.Close()

	return key.DeleteValue("Orryg")
}

// https://github.com/golang/sys/blob/master/windows/svc/example/install.go#L18-L42
func executablePath() (string, error) {
	prog := os.Args[0]
	p, err := filepath.Abs(prog)
	if err != nil {
		return "", err
	}
	fi, err := os.Stat(p)
	if err == nil {
		if !fi.Mode().IsDir() {
			return p, nil
		}
		err = fmt.Errorf("%s is directory", p)
	}
	if filepath.Ext(p) == "" {
		p += ".exe"
		fi, err := os.Stat(p)
		if err == nil {
			if !fi.Mode().IsDir() {
				return p, nil
			}
			err = fmt.Errorf("%s is directory", p)
		}
	}
	return "", err
}
