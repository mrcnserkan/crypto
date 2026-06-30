package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var alertWatchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch price alerts in foreground",
	Long: `Continuously monitor active price alerts in the foreground.

Checks alert conditions at the configured interval (default: 5 minutes)
and prints terminal notifications when targets are hit.

Press Ctrl+C to stop.

EXAMPLE:
  crypto alert watch
  crypto alert watch --currency eur`,
	Run: func(cmd *cobra.Command, args []string) {
		alerts := alertManager.GetAlerts()
		if len(alerts) == 0 {
			fmt.Println("No active alerts to watch")
			return
		}

		intervalMin := configStore.AlertIntervalOrDefault(5)
		alertChecker.SetInterval(minutesToDuration(intervalMin))

		titleColor := color.New(color.FgHiCyan, color.Bold).SprintFunc()
		fmt.Printf("\n%s Watching %d alert(s). Press Ctrl+C to stop.\n\n",
			titleColor("🔔"), len(alerts))

		alertChecker.Start()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		fmt.Println("\nStopping alert watcher...")
		alertChecker.Stop()
	},
}

func init() {
	alertCmd.AddCommand(alertWatchCmd)
}
