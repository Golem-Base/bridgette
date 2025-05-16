package webui

import (
	"fmt"
	"time"
)

// shortenAddress shortens an Ethereum address for display
func shortenAddress(address string) string {
	// TODO: Implement address shortening when we add links to explorer
	return address
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
