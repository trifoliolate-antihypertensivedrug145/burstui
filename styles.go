package main

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var progressRE = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)%`)
var progressCountRE = regexp.MustCompile(`(?i)(\d+)\s*/\s*(\d+)`)
var statusCodeInlineRE = regexp.MustCompile(`(?i)(status:\s*)(\d{3})`)

var (
	titleStyle         = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	subtitleStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	versionStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("109"))
	sectionTitleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230"))
	boxStyle           = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(1, 2)
	activeLabelStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	inactiveLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	helpStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	errorStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
	successStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	infoStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("117"))
	status2xxStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
	status3xxStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	status4xxStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("208"))
	status5xxStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	footerSpinnerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
)

func renderLogLine(line string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return ""
	}

	prefix := infoStyle.Render("• ")
	lower := strings.ToLower(trimmed)

	if strings.Contains(lower, "error") || strings.Contains(lower, "failed") {
		return prefix + errorStyle.Render(trimmed)
	}
	if strings.Contains(lower, "success") || strings.Contains(lower, "finished successfully") {
		return prefix + successStyle.Render(trimmed)
	}

	rendered := trimmed
	rendered = statusCodeInlineRE.ReplaceAllStringFunc(rendered, func(s string) string {
		parts := statusCodeInlineRE.FindStringSubmatch(s)
		if len(parts) != 3 {
			return s
		}
		return parts[1] + colorizeStatusCode(parts[2])
	})

	return prefix + rendered
}

func colorizeStatusCode(code string) string {
	switch {
	case strings.HasPrefix(code, "2"):
		return status2xxStyle.Render(code)
	case strings.HasPrefix(code, "3"):
		return status3xxStyle.Render(code)
	case strings.HasPrefix(code, "4"):
		return status4xxStyle.Render(code)
	case strings.HasPrefix(code, "5"):
		return status5xxStyle.Render(code)
	default:
		return code
	}
}

func renderGradientText(text string, offset int) string {
	palette := []lipgloss.Color{
		lipgloss.Color("39"),
		lipgloss.Color("45"),
		lipgloss.Color("51"),
		lipgloss.Color("87"),
		lipgloss.Color("123"),
		lipgloss.Color("159"),
	}

	runes := []rune(text)
	var b strings.Builder
	for i, r := range runes {
		if r == ' ' {
			b.WriteRune(r)
			continue
		}
		color := palette[(i+offset)%len(palette)]
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(color).Render(string(r)))
	}
	return b.String()
}

func gradientBorderColor(offset int) lipgloss.Color {
	palette := []lipgloss.Color{
		lipgloss.Color("39"),
		lipgloss.Color("45"),
		lipgloss.Color("51"),
		lipgloss.Color("87"),
		lipgloss.Color("123"),
		lipgloss.Color("159"),
	}
	return palette[offset%len(palette)]
}
