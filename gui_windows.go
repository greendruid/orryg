package main

import (
	"syscall"
	"unsafe"

	"github.com/AllenDang/w32"
)

var trayIcon struct {
	nid w32.NOTIFYICONDATA
	// bitmap      w32.HBITMAP
	// bitmapWidth int
	// font        w32.HFONT
	hWnd w32.HWND
	// mdc         w32.HDC
}

// func rgb(r, g, b byte) w32.COLORREF {
// 	return w32.COLORREF(r) | (w32.COLORREF(g) << 8) | (w32.COLORREF(b) << 16)
// }
//
// func trayIconDraw() w32.HICON {
// 	oldBitmap := w32.HBITMAP(w32.SelectObject(trayIcon.mdc, w32.HGDIOBJ(trayIcon.bitmap)))
// 	oldFont := w32.HFONT(w32.SelectObject(trayIcon.mdc, w32.HGDIOBJ(trayIcon.font)))
//
// 	w32.TextOut(trayIcon.mdc, trayIcon.bitmapWidth/4, 0, syscall.StringToUTF16Ptr("O"), 1)
//
// 	// rect := w32.GetClientRect(trayIcon.hWnd)
// 	// w32.FillRect(trayIcon.mdc, rect, w32.CreateSolidBrush(rgb(0xff, 0x33, 0x33)))
//
// 	w32.SelectObject(trayIcon.mdc, w32.HGDIOBJ(oldBitmap))
// 	w32.SelectObject(trayIcon.mdc, w32.HGDIOBJ(oldFont))
//
// 	info := w32.ICONINFO{
// 		FIcon:    true,
// 		HBMMask:  trayIcon.bitmap,
// 		HBMColor: trayIcon.bitmap,
// 	}
//
// 	return w32.CreateIconIndirect(&info)
// }

const (
	wmTrayMessage = w32.WM_USER + 1
)

var (
	trayClassName = syscall.StringToUTF16Ptr("Tray")
)

func trayWndProc(hwnd w32.HWND, msg uint32, wparam, lparam uintptr) uintptr {
	switch msg {
	case wmTrayMessage:
		logger.Printf("wmTrayMessage")

	case w32.WM_SYSCOMMAND:

		logger.Printf("sys command !")
		trayIcon.nid.SetTooltip("TOOOLTIP")
		logger.Println(w32.ShellNotifyIcon(w32.NIM_ADD, &trayIcon.nid))

	default:
		return w32.DefWindowProc(hwnd, msg, wparam, lparam)
	}

	return 0
}

func trayIconInit() {
	appInstance := w32.GetModuleHandle("")

	classEx := w32.WNDCLASSEX{
		Style:      w32.CS_HREDRAW | w32.CS_VREDRAW,
		WndProc:    syscall.NewCallback(trayWndProc),
		Instance:   appInstance,
		Cursor:     w32.LoadCursor(0, w32.MakeIntResource(w32.IDC_ARROW)),
		Background: w32.HBRUSH(w32.COLOR_BACKGROUND),
		ClassName:  trayClassName,
	}
	classEx.Size = uint32(unsafe.Sizeof(classEx))

	logger.Println("oooaoaoao", w32.RegisterClassEx(&classEx))

	trayIcon.hWnd = w32.CreateWindowEx(
		0,
		trayClassName,
		syscall.StringToUTF16Ptr("Orryg"),
		w32.WS_OVERLAPPEDWINDOW|w32.WS_VISIBLE,
		w32.CW_USEDEFAULT, w32.CW_USEDEFAULT, // x, y
		400, 400, // width, height
		w32.HWND(0),                // parent
		w32.HMENU(0),               // menu
		appInstance,                // instance
		unsafe.Pointer(uintptr(0)), // param (??)
	)

	// trayIcon.bitmapWidth = w32.GetSystemMetrics(w32.SM_CXSMICON)
	// hdc := w32.GetDC(trayIcon.hWnd)
	// trayIcon.bitmap = w32.CreateCompatibleBitmap(hdc, trayIcon.bitmapWidth, trayIcon.bitmapWidth)
	// trayIcon.mdc = w32.CreateCompatibleDC(hdc)
	// w32.ReleaseDC(trayIcon.hWnd, hdc)
	// w32.SetBkColor(trayIcon.mdc, rgb(0xFF, 0xFF, 0xFF))
	// w32.SetTextColor(trayIcon.mdc, rgb(0x00, 0xFF, 0x00))
	// trayIcon.font = w32.CreateFont(
	// 	-w32.MulDiv(11, w32.GetDeviceCaps(trayIcon.mdc, w32.LOGPIXELSY), 72),
	// 	0, 0, 0, 0, 0, 0,
	// 	0, 0, 0, 0, 0, 0,
	// 	syscall.StringToUTF16Ptr("Arial"),
	// )

	trayIcon.nid.Size = uint32(unsafe.Sizeof(trayIcon.nid))
	trayIcon.nid.Hwnd = trayIcon.hWnd
	trayIcon.nid.ID = 5000
	trayIcon.nid.TimeoutOrVersion = 4
	trayIcon.nid.Flags = w32.NIF_ICON | w32.NIF_MESSAGE | w32.NIF_TIP

	// trayIcon.nid.Flags = w32.NIF_INFO
	// trayIcon.nid.InfoFlags = w32.NIIF_INFO
	// trayIcon.nid.SetTooltip("my tooltip")
	// trayIcon.nid.SetInfo("my info")
	// trayIcon.nid.SetInfoTitle("my title")
	// trayIcon.nid.Icon = trayIconDraw()
	// trayIcon.nid.Icon = w32.LoadImage(appInstance, syscall.StringToUTF16Ptr("./lalala.ico"), w32.IMAGE_ICON, 32, 32, w32.LR_LOADFROMFILE)

	trayIcon.nid.Icon = w32.LoadIcon(appInstance, w32.MakeIntResource(w32.IDI_APPLICATION))
	trayIcon.nid.CallbackMessage = wmTrayMessage

	logger.Printf("nid: %#v", trayIcon.nid)
}
