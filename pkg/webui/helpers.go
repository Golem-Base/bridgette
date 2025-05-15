package webui

import (
	"fmt"
	"time"

	"github.com/a-h/templ"
)

// shortenAddress shortens an Ethereum address for display
func shortenAddress(address string) string {
	if len(address) < 10 {
		return address
	}
	return address[0:6] + "..." + address[len(address)-4:]
}

// formatTime formats a time for display
func formatTime(t time.Time) string {
	return t.Format("Jan 02, 2006 15:04:05")
}

// formatTimeDiff formats a time difference in seconds for display
func formatTimeDiff(seconds int64) string {
	if seconds < 60 {
		return fmt.Sprintf("%d seconds", seconds)
	} else if seconds < 3600 {
		return fmt.Sprintf("%.1f minutes", float64(seconds)/60)
	} else {
		return fmt.Sprintf("%.1f hours", float64(seconds)/3600)
	}
}

// safeURL creates a safe URL for templ
func safeURL(url string) templ.SafeURL {
	return templ.URL(url)
}

// etherscanURL creates a safe Etherscan URL
func etherscanURL(txHash string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("https://etherscan.io/tx/%s", txHash))
}

// explorerURL creates a safe Golem explorer URL
func explorerURL(txHash string) templ.SafeURL {
	return templ.URL(fmt.Sprintf("https://explorer.golem.network/tx/%s", txHash))
}
