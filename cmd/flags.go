package cmd

import (
	"github.com/fatih/color"
	"github.com/mrcnserkan/crypto/v2/service"
	"github.com/mrcnserkan/crypto/v2/utils"
	"github.com/spf13/cobra"
)

func getCurrencyFlag(cmd *cobra.Command) string {
	if rootCmd.PersistentFlags().Changed("currency") {
		currency, _ := rootCmd.PersistentFlags().GetString("currency")
		return utils.NormalizeCurrency(currency)
	}
	if configStore != nil {
		return configStore.CurrencyOrDefault(service.DEFAULT_CURRENCY)
	}
	return utils.NormalizeCurrency(service.DEFAULT_CURRENCY)
}

func disableColorsIfNeeded() {
	noColor, _ := rootCmd.PersistentFlags().GetBool("no-color")
	if noColor || (configStore != nil && configStore.Config.NoColor) {
		color.NoColor = true
	}
}
