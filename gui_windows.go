package main

import (
	"errors"
	"image"
	"image/png"
	"os"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
)

var trayIcon struct {
	nid win.NOTIFYICONDATA
	// bitmap      w32.HBITMAP
	// bitmapWidth int
	// font        w32.HFONT
	hWnd win.HWND
	// mdc         w32.HDC
}

const (
	wmTrayMessage = win.WM_USER + 1
)

var (
	trayClassName = syscall.StringToUTF16Ptr("Tray")
)

func hBitmapFromImage(im image.Image) (win.HBITMAP, error) {
	var bi win.BITMAPV5HEADER
	bi.BiSize = uint32(unsafe.Sizeof(bi))
	bi.BiWidth = int32(im.Bounds().Dx())
	bi.BiHeight = -int32(im.Bounds().Dy())
	bi.BiPlanes = 1
	bi.BiBitCount = 32
	bi.BiCompression = win.BI_BITFIELDS
	// The following mask specification specifies a supported 32 BPP
	// alpha format for Windows XP.
	bi.BV4RedMask = 0x00FF0000
	bi.BV4GreenMask = 0x0000FF00
	bi.BV4BlueMask = 0x000000FF
	bi.BV4AlphaMask = 0xFF000000

	hdc := win.GetDC(0)
	defer win.ReleaseDC(0, hdc)

	var lpBits unsafe.Pointer

	// Create the DIB section with an alpha channel.
	hBitmap := win.CreateDIBSection(hdc, &bi.BITMAPINFOHEADER, win.DIB_RGB_COLORS, &lpBits, 0, 0)
	switch hBitmap {
	case 0, win.ERROR_INVALID_PARAMETER:
		return 0, errors.New("CreateDIBSection failed")
	}

	// Fill the image
	bitmapArray := (*[1 << 30]byte)(unsafe.Pointer(lpBits))
	i := 0
	for y := im.Bounds().Min.Y; y != im.Bounds().Max.Y; y++ {
		for x := im.Bounds().Min.X; x != im.Bounds().Max.X; x++ {
			r, g, b, a := im.At(x, y).RGBA()
			bitmapArray[i+3] = byte(a >> 8)
			bitmapArray[i+2] = byte(r >> 8)
			bitmapArray[i+1] = byte(g >> 8)
			bitmapArray[i+0] = byte(b >> 8)
			i += 4
		}
	}

	return hBitmap, nil
}

// create an Alpha Icon or Cursor from an Image
// http://support.microsoft.com/kb/318876
func createAlphaCursorOrIconFromImage(im image.Image, fIcon bool) (win.HICON, error) {
	hBitmap, err := hBitmapFromImage(im)
	if err != nil {
		return 0, err
	}
	defer win.DeleteObject(win.HGDIOBJ(hBitmap))

	// Create an empty mask bitmap.
	hMonoBitmap := win.CreateBitmap(int32(im.Bounds().Dx()), int32(im.Bounds().Dy()), 1, 1, nil)
	if hMonoBitmap == 0 {
		return 0, errors.New("CreateBitmap failed")
	}
	defer win.DeleteObject(win.HGDIOBJ(hMonoBitmap))

	var ii win.ICONINFO
	if fIcon {
		ii.FIcon = win.TRUE
	}
	ii.XHotspot = 0
	ii.YHotspot = 0
	ii.HbmMask = hMonoBitmap
	ii.HbmColor = hBitmap

	// Create the alpha cursor with the alpha DIB section.
	hIconOrCursor := win.CreateIconIndirect(&ii)

	return hIconOrCursor, nil
}

func trayWndProc(hwnd win.HWND, msg uint32, wparam, lparam uintptr) uintptr {
	switch msg {
	case wmTrayMessage:
		logger.Printf("wmTrayMessage")

	default:
		return win.DefWindowProc(hwnd, msg, wparam, lparam)
	}

	return 0
}

func trayIconInit() error {
	appInstance := win.GetModuleHandle(nil)

	classEx := win.WNDCLASSEX{
		Style:         win.CS_HREDRAW | win.CS_VREDRAW,
		LpfnWndProc:   syscall.NewCallback(trayWndProc),
		HInstance:     appInstance,
		HCursor:       win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_ARROW)),
		HbrBackground: win.HBRUSH(win.COLOR_BACKGROUND),
		LpszClassName: trayClassName,
	}
	classEx.CbSize = uint32(unsafe.Sizeof(classEx))

	logger.Println("oooaoaoao", win.RegisterClassEx(&classEx))

	trayIcon.hWnd = win.CreateWindowEx(
		0,
		trayClassName,
		syscall.StringToUTF16Ptr("Orryg"),
		win.WS_OVERLAPPEDWINDOW,
		win.CW_USEDEFAULT, win.CW_USEDEFAULT, // x, y
		400, 400, // width, height
		win.HWND(0),                // parent
		win.HMENU(0),               // menu
		appInstance,                // instance
		unsafe.Pointer(uintptr(0)), // param (??)
	)

	trayIcon.nid.CbSize = uint32(unsafe.Sizeof(trayIcon.nid))
	trayIcon.nid.HWnd = trayIcon.hWnd
	trayIcon.nid.UID = 5000
	trayIcon.nid.UVersion = 4
	trayIcon.nid.UFlags = win.NIF_ICON | win.NIF_MESSAGE | win.NIF_TIP
	trayIcon.nid.UCallbackMessage = wmTrayMessage

	// trayIcon.nid.HIcon = win.HICON(win.LoadImage(appInstance, syscall.StringToUTF16Ptr("./temp.ico"), win.IMAGE_ICON, 32, 32, win.LR_LOADFROMFILE))
	{
		// img := image.NewRGBA(image.Rect(0, 0, 32, 32))
		// for y := img.Bounds().Min.Y; y != img.Bounds().Max.Y; y++ {
		// 	for x := img.Bounds().Min.X; x != img.Bounds().Max.X; x++ {
		// 		img.Set(x, y, color.RGBA{0xFF, 0, 0, 0xFF})
		// 	}
		// }
		//

		var im image.Image
		{
			f, err := os.Open("./cancel.png")
			if err != nil {
				return err
			}

			img, err := png.Decode(f)
			if err != nil {
				return err
			}
			im = img
		}

		icon, err := createAlphaCursorOrIconFromImage(im, true)
		if err != nil {
			return err
		}
		trayIcon.nid.HIcon = icon
	}

	data, _ := syscall.UTF16FromString("TOOLTIP")
	copy(trayIcon.nid.SzTip[:], data)

	logger.Println(win.Shell_NotifyIcon(win.NIM_ADD, &trayIcon.nid))

	logger.Printf("nid: %#v", trayIcon.nid)

	return nil
}
