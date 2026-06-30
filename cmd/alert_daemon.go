package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var alertStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start alert checker in background",
	Long: `Start a background daemon that monitors price alerts.

The daemon writes its PID to ~/.crypto/alert.pid and checks alerts
every 5 minutes (configurable via config.json).

EXAMPLE:
  crypto alert start`,
	Run: func(cmd *cobra.Command, args []string) {
		if running, pid := daemonState.IsRunning(); running {
			fmt.Printf("Alert daemon already running (PID %d)\n", pid)
			return
		}
		if len(alertManager.GetAlerts()) == 0 {
			fmt.Println("No active alerts to monitor")
			return
		}

		executable, err := os.Executable()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		logFile := filepath.Join(configDir, "alert.log")
		logOut, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			fmt.Printf("Error creating log file: %v\n", err)
			os.Exit(1)
		}

		child := exec.Command(executable, "alert", "watch")
		child.Stdout = logOut
		child.Stderr = logOut
		child.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

		if err := child.Start(); err != nil {
			_ = logOut.Close()
			fmt.Printf("Error starting daemon: %v\n", err)
			os.Exit(1)
		}
		_ = logOut.Close()

		if err := daemonState.WritePID(child.Process.Pid); err != nil {
			fmt.Printf("Error writing PID file: %v\n", err)
			os.Exit(1)
		}

		titleColor := color.New(color.FgHiCyan, color.Bold).SprintFunc()
		fmt.Printf("\n%s Alert daemon started (PID %d)\n", titleColor("🔔"), child.Process.Pid)
		fmt.Printf("Logs: %s\n", logFile)
	},
}

var alertStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop background alert daemon",
	Run: func(cmd *cobra.Command, args []string) {
		running, pid := daemonState.IsRunning()
		if !running {
			_ = daemonState.RemovePID()
			fmt.Println("Alert daemon is not running")
			return
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if err := process.Signal(syscall.SIGTERM); err != nil {
			fmt.Printf("Error stopping daemon: %v\n", err)
			os.Exit(1)
		}

		for i := 0; i < 10; i++ {
			if running, _ = daemonState.IsRunning(); !running {
				break
			}
			time.Sleep(200 * time.Millisecond)
		}
		_ = daemonState.RemovePID()

		titleColor := color.New(color.FgHiCyan, color.Bold).SprintFunc()
		fmt.Printf("\n%s Alert daemon stopped\n", titleColor("🔔"))
	},
}

var alertStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show alert daemon status",
	Run: func(cmd *cobra.Command, args []string) {
		running, pid := daemonState.IsRunning()
		alerts := alertManager.GetAlerts()

		titleColor := color.New(color.FgHiCyan, color.Bold).SprintFunc()
		fmt.Printf("\n%s Alert Status\n\n", titleColor("🔔"))
		if running {
			fmt.Printf("Daemon: running (PID %d)\n", pid)
		} else {
			fmt.Println("Daemon: stopped")
		}
		fmt.Printf("Active alerts: %d\n", len(alerts))
		if len(alerts) > 0 && !running {
			fmt.Println("\nTip: run `crypto alert watch` or `crypto alert start` to monitor alerts")
		}
	},
}

func init() {
	alertCmd.AddCommand(alertStartCmd)
	alertCmd.AddCommand(alertStopCmd)
	alertCmd.AddCommand(alertStatusCmd)
}

func minutesToDuration(minutes int) time.Duration {
	return time.Duration(minutes) * time.Minute
}
