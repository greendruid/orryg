package main

import (
	"image"
	"image/png"
	"os"

	"github.com/lxn/walk"
)

type mouseDownEvent struct {
	x      int
	y      int
	button walk.MouseButton
}

type trayIcon struct {
	im         image.Image
	icon       *walk.Icon
	notifyIcon *walk.NotifyIcon

	mouseDownHandler int
	mouseDownCh      chan mouseDownEvent

	err error
}

func newTrayIcon() *trayIcon {
	return &trayIcon{
		mouseDownCh: make(chan mouseDownEvent),
	}
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
		i.mouseDownCh <- mouseDownEvent{x, y, button}
	})
}

func (i *trayIcon) init() error {
	i.loadImage()
	i.makeNotifyIcon()
	i.attachMouseDownHandler()
	i.makeIcon()
	i.setIcon()
	i.setVisible(true)

	return i.err
}
