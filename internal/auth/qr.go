package auth

import (
	"fmt"
	"strings"

	"github.com/skip2/go-qrcode"
)

func GenerateQRCode(data string) (string, error) {
	qr, err := qrcode.New(data, qrcode.Medium)
	if err != nil {
		return "", fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Convert QR code to ASCII art for terminal display
	return qr.ToSmallString(false), nil
}

func FormatQRForTerminal(qrString string) string {
	lines := strings.Split(qrString, "\n")
	var result strings.Builder

	result.WriteString("┌" + strings.Repeat("─", len(lines[0])+2) + "┐\n")

	for _, line := range lines {
		if line != "" {
			result.WriteString("│ " + line + " │\n")
		}
	}

	result.WriteString("└" + strings.Repeat("─", len(lines[0])+2) + "┘\n")

	return result.String()
}
