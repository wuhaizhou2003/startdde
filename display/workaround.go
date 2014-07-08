package display

import "dbus/com/deepin/daemon/keybinding"

var __keepMediakeyManagerAlive interface{}

func (dpy *Display) workaroundBacklight() {
	mediaKeyManager, err := keybinding.NewMediaKey("com.deepin.daemon.KeyBinding", "/com/deepin/daemon/MediaKey")
	if err != nil {
		Logger.Error("Can't connect to /com/deepin/daemon/MediaKey", err)
		return
	}
	__keepMediakeyManagerAlive = mediaKeyManager

	workaround := func(m *Monitor) {
		dpyinfo := GetDisplayInfo()
		for _, name := range dpyinfo.ListNames() {
			op := dpyinfo.QueryOutputs(name)
			if backlight, support := supportedBacklight(xcon, op); support {
				dpy.setPropBrightness(name, backlight)
				dpy.saveBrightness(name, backlight)
			}
		}
	}

	mediaKeyManager.ConnectBrightnessUp(func(onPress bool) {
		if !onPress {
			for _, m := range dpy.Monitors {
				workaround(m)
			}
		}
	})
	mediaKeyManager.ConnectBrightnessDown(func(onPress bool) {
		if !onPress {
			for _, m := range dpy.Monitors {
				workaround(m)
			}
		}
	})

	mediaKeyManager.ConnectSwitchMonitors(func(onPress bool) {
		if !onPress {
			if int(dpy.DisplayMode) >= len(dpy.Monitors) {
				dpy.SwitchMode(DisplayModeMirrors)
			} else {
				dpy.SwitchMode(dpy.DisplayMode + 1)
			}
		}
	})
}