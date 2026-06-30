package models

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const alertPIDFileName = "alert.pid"

// DaemonState manages background process PID file.
type DaemonState struct {
	pidFile string
}

func NewDaemonState(configDir string) *DaemonState {
	return &DaemonState{pidFile: filepath.Join(configDir, alertPIDFileName)}
}

func (d *DaemonState) PIDFile() string {
	return d.pidFile
}

func (d *DaemonState) WritePID(pid int) error {
	return atomicWriteFile(d.pidFile, []byte(strconv.Itoa(pid)))
}

func (d *DaemonState) ReadPID() (int, error) {
	data, err := os.ReadFile(d.pidFile)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("invalid pid file")
	}
	return pid, nil
}

func (d *DaemonState) RemovePID() error {
	err := os.Remove(d.pidFile)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (d *DaemonState) IsRunning() (bool, int) {
	pid, err := d.ReadPID()
	if err != nil {
		return false, 0
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, pid
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil, pid
}
