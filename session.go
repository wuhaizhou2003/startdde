package main

import (
	"dbus/org/freedesktop/login1"
	"dlib/dbus"
	"fmt"
	"os"
	"time"
)

type SessionManager struct {
	CurrentUid string
	cookies    map[string]chan time.Time
	Stage      int32
}

const (
	_LOCK_EXEC        = "/usr/bin/dde-lock"
	_SHUTDOWN_CMD     = "/usr/bin/dde-shutdown"
	_REBOOT_ARG       = "--reboot"
	_LOGOUT_ARG       = "--logout"
	_SHUTDOWN_ARG     = "--shutdown"
	_POWER_CHOOSE_ARG = "--choice"
)

const (
	SessionStageInitBegin int32 = iota
	SessionStageInitEnd
	SessionStageCoreBegin
	SessionStageCoreEnd
	SessionStageAppsBegin
	SessionStageAppsEnd
)

var (
	objLogin *login1.Manager
)

func (m *SessionManager) CanLogout() bool {
	return true
}

func (m *SessionManager) Logout() {
	execCommand(_SHUTDOWN_CMD, _LOGOUT_ARG)
}

func (m *SessionManager) RequestLogout() {
	os.Exit(0)
}

func (m *SessionManager) ForceLogout() {
	os.Exit(0)
}

func (shudown *SessionManager) CanShutdown() bool {
	str, _ := objLogin.CanPowerOff()
	if str == "yes" {
		return true
	}

	return false
}

func (m *SessionManager) Shutdown() {
	execCommand(_SHUTDOWN_CMD, _SHUTDOWN_ARG)
}

func (m *SessionManager) RequestShutdown() {
	objLogin.PowerOff(true)
}

func (m *SessionManager) ForceShutdown() {
	objLogin.PowerOff(false)
}

func (shudown *SessionManager) CanReboot() bool {
	str, _ := objLogin.CanReboot()
	if str == "yes" {
		return true
	}

	return false
}

func (m *SessionManager) Reboot() {
	execCommand(_SHUTDOWN_CMD, _REBOOT_ARG)
}

func (m *SessionManager) RequestReboot() {
	objLogin.Reboot(true)
}

func (m *SessionManager) ForceReboot() {
	objLogin.Reboot(false)
}

func (m *SessionManager) CanSuspend() bool {
	str, _ := objLogin.CanSuspend()
	if str == "yes" {
		return true
	}
	return false
}

func (m *SessionManager) RequestSuspend() {
	objLogin.Suspend(false)
}

func (m *SessionManager) CanHibernate() bool {
	str, _ := objLogin.CanHibernate()
	if str == "yes" {
		return true
	}
	return false
}

func (m *SessionManager) RequestHibernate() {
	objLogin.Hibernate(false)
}

func (m *SessionManager) RequestLock() {
	execCommand(_LOCK_EXEC, "")
}

func (m *SessionManager) PowerOffChoose() {
	execCommand(_SHUTDOWN_CMD, _POWER_CHOOSE_ARG)
}

func initSession() {
	var err error

	objLogin, err = login1.NewManager("org.freedesktop.login1",
		"/org/freedesktop/login1")
	if err != nil {
		panic(fmt.Errorf("New Login1 Failed: %s", err))
	}
}

func newSessionManager() *SessionManager {
	m := &SessionManager{}
	m.cookies = make(map[string]chan time.Time)
	m.setPropName("CurrentUid")

	return m
}

func startSession() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("StartSession recover:", err)
			return
		}
	}()

	initSession()
	manager := newSessionManager()
	err := dbus.InstallOnSession(manager)
	if err != nil {
		logger.Error("Install Session DBus Failed:", err)
		return
	}

	manager.setPropStage(SessionStageInitBegin)

	initBackground()
	manager.launch("/usr/bin/gtk-window-decorator", false)
	manager.launch("/usr/bin/compiz", false)
	initBackgroundAfterDependsLoaded()
	manager.setPropStage(SessionStageInitEnd)

	manager.setPropStage(SessionStageCoreBegin)
	startStartManager()

	manager.launch("/usr/lib/deepin-daemon/dde-session-daemon", true)

	showGuide := manager.ShowGuideOnce()

	manager.launch("/usr/bin/dde-desktop", true)
	manager.launch("/usr/bin/dde-dock", true)
	manager.launch("/usr/bin/dde-dock-applets", false)

	if showGuide {
		manager.launch("/usr/bin/dde-launcher", true, "--hidden")
	}

	manager.setPropStage(SessionStageCoreEnd)

	manager.setPropStage(SessionStageAppsBegin)

	if !debug {
		startAutostartProgram()
	}
	manager.setPropStage(SessionStageAppsEnd)
}

func (m *SessionManager) ShowGuideOnce() bool {
	path := os.ExpandEnv("$HOME/.config/not_first_run_dde")
	_, err := os.Stat(path)
	if err != nil {
		f, err := os.Create(path)
		defer f.Close()
		if err != nil {
			Logger.Error("Can't initlize first dde", err)
			return false
		}

		m.launch("/usr/lib/deepin-daemon/dde-guide", true)
		return true
	}

	return false
}
