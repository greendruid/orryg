package main

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"net/http"
	_ "net/http/pprof"

	"github.com/lxn/walk"
	"github.com/vrischmann/userdir"
)

func getLogFile() io.Writer {
	dir := filepath.Join(userdir.GetDataHome(), "orryg")

	{
		fi, err := os.Stat(dir)
		if err != nil && os.IsNotExist(err) {
			os.MkdirAll(dir, 0700)
		} else if err != nil && !os.IsNotExist(err) {
			log.Printf("unable to create log directory %s. err=%v", dir, err)
			return ioutil.Discard
		} else {
			if !fi.IsDir() {
				log.Printf("unable to create log directory %s because it's already a file", dir)
				return ioutil.Discard
			}
		}
	}

	file := filepath.Join(dir, "main.log")
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Printf("unable to create log file %s. err=%v", file, err)
		return ioutil.Discard
	}

	return f
}

var (
	flConfigure bool
	flVerbose   bool
)

func init() {
	flag.BoolVar(&flConfigure, "c", false, "Run the configuration prompt")
	flag.BoolVar(&flConfigure, "configure", false, "Run the configuration prompt")
	flag.BoolVar(&flVerbose, "v", false, "Be verbose (print to stdout too)")
}

func main() {
	flag.Parse()

	go http.ListenAndServe(":6060", nil)

	// if flConfigure {
	// 	cp := configurePrompt{conf: newWindowsConfiguration()}
	// 	cp.run()
	// 	return
	// }

	if flVerbose {
		logger = log.New(io.MultiWriter(getLogFile(), os.Stdout), "orryg: ", log.LstdFlags)
	} else {
		logger = log.New(getLogFile(), "orryg: ", log.LstdFlags)
	}

	{
		conf := newWindowsConfiguration()
		s, err := conf.DumpConfig()
		if err != nil {
			logger.Printf("there was a problem while dumping the configuration. err=%v", err)
			return
		}

		logger.Printf("configuration dump")
		for _, line := range s {
			logger.Printf("%s", line)
		}
	}

	e := newEngine(newWindowsConfiguration())
	go e.run()

	var (
		mw                 *mainWindow
		tw                 *tabWidget
		directoriesTabPage *walk.TabPage
		directoriesListBox *walk.ListBox
		copiersTabPage     *walk.TabPage
		copiersListBox     *walk.ListBox

		err error
	)

	{
		mw, err = newMainWindow()
		if err != nil {
			logger.Printf("unable to create new main window. err=%v", err)
			return
		}
		mw.SetSize(walk.Size{Width: 640, Height: 480})
		mw.SetLayout(walk.NewHBoxLayout())
	}

	{
		tw, err = newTabWidget(mw)
		if err != nil {
			logger.Printf("unable to create new tab widget. err=%v", err)
			return
		}
		tw.SetVisible(true)

		pages := tw.Pages()

		{
			directoriesTabPage, err = walk.NewTabPage()
			if err != nil {
				logger.Printf("unable to create directories tab page. err=%v", err)
				return
			}
			directoriesTabPage.SetTitle("Directories")
			directoriesTabPage.SetLayout(walk.NewHBoxLayout())
			pages.Add(directoriesTabPage)
		}

		{
			directoriesListBox, err = walk.NewListBox(directoriesTabPage)
			if err != nil {
				logger.Printf("unable to create directories list box. err=%v", err)
			}
			directoriesListBox.SetModel([]string{"foo", "bar"})
		}

		{
			copiersTabPage, err = walk.NewTabPage()
			if err != nil {
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

	{
		tray := trayIcon{
			mwHwnd: mw.Handle(),
		}

		if err := tray.init(); err != nil {
			logger.Printf("unable to initialize tray icon. err=%v", err)
			return
		}
	}

	mw.SetVisible(true)
	mw.Run()

	// 	msg := new(win.MSG)
	// loop:
	// 	for {
	// 		if n := win.GetMessage(msg, 0, 0, 0); n == 0 || n == -1 {
	// 			break
	// 		}
	//
	// 		switch msg.Message {
	// 		case wmShowUI:
	//
	// 		case win.WM_CLOSE:
	// 			break loop
	// 		default:
	// 			win.TranslateMessage(msg)
	// 			win.DispatchMessage(msg)
	// 		}
	// 	}

	// err := e.stop()
	// if err != nil {
	// 	logger.Printf("unable to stop engine correctly. err=%v", err)
	// }
}
