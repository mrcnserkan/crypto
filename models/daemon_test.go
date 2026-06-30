package models

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDaemonState_WriteReadRemovePID(t *testing.T) {
	dir := t.TempDir()
	d := NewDaemonState(dir)

	if err := d.WritePID(12345); err != nil {
		t.Fatalf("WritePID() error = %v", err)
	}

	pid, err := d.ReadPID()
	if err != nil || pid != 12345 {
		t.Fatalf("ReadPID() = (%d, %v)", pid, err)
	}

	if err := d.RemovePID(); err != nil {
		t.Fatalf("RemovePID() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, alertPIDFileName)); !os.IsNotExist(err) {
		t.Fatal("expected pid file removed")
	}
}
