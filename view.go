package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	var leftContent string

	switch m.stage {
	case stageBrowse:
		leftContent = m.browseView()
	default:
		leftContent = m.formView()
	}

	headerTitle := lipgloss.JoinHorizontal(
		lipgloss.Center,
		titleStyle.Render("BurstUI"),
		" ",
		versionStyle.Render("v0.1.0"),
	)
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		headerTitle,
		subtitleStyle.Render("A terminal UI for Gobuster"),
	)

	leftWidth := 48
	rightWidth := 56
	if m.width > 0 {
		usable := m.width - 6
		if usable > 40 {
			leftWidth = usable / 2
			rightWidth = usable - leftWidth
		}
	}
	if m.stage == stageBrowse {
		leftWidth = rightWidth
	}

	leftStyle := boxStyle
	rightStyle := boxStyle

	if m.stage == stageScanning {
		leftStyle = leftStyle.BorderForeground(lipgloss.Color("241"))
		rightStyle = rightStyle.BorderForeground(gradientBorderColor(m.gradientStep))
	} else {
		if !m.resultFocused {
			leftStyle = leftStyle.BorderForeground(lipgloss.Color("86"))
		}
		if m.resultFocused {
			rightStyle = rightStyle.BorderForeground(lipgloss.Color("86"))
		}
	}

	leftPanel := leftStyle.Width(leftWidth).Render(leftContent)

	viewportContent := m.viewport.View()
	footer := helpStyle.Render(fmt.Sprintf("Scroll: %3.0f%%", m.viewport.ScrollPercent()*100))
	if m.stage == stageScanning {
		spinnerText := footerSpinnerStyle.Render(m.spinner.View())
		left := helpStyle.Render(fmt.Sprintf("Scroll: %3.0f%%", m.viewport.ScrollPercent()*100))
		gap := rightWidth - 4 - lipgloss.Width(left) - lipgloss.Width(spinnerText)
		if gap < 1 {
			gap = 1
		}
		footer = left + strings.Repeat(" ", gap) + spinnerText
	}
	hint := helpStyle.Render("Press Tab to focus result pane")
	if m.resultFocused {
		hint = helpStyle.Render("Result pane focused • ↑/↓ scroll • ←/esc/Tab back to Configuration")
	} else if m.stage == stageScanning {
		hint = helpStyle.Render("Scanning • Press → to focus live results • ctrl+c quits")
	}
	content := strings.Join([]string{viewportContent, "", footer, hint}, "\n")
	rightPanel := rightStyle.Width(rightWidth).Render(content)

	columns := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	return lipgloss.JoinVertical(lipgloss.Left, header, columns)
}

func (m model) rightPanelView() string {
	sections := make([]string, 0, 2)

	switch {
	case m.stage == stageScanning:
		sections = append(sections, m.scanView())
	case m.errorMsg != "":
		sections = append(sections, m.doneView())
	case m.scanFinished:
		sections = append(sections, m.doneView())
	default:
		statusBody := helpStyle.Render(m.lastStatus)
		if m.logSaveError != "" {
			statusBody += "\n\n" + errorStyle.Render("Log save error: "+m.logSaveError)
		}
		sections = append(sections, sectionTitleStyle.Render("Status")+"\n\n"+statusBody)
	}

	var lines []string
	for _, line := range m.logs {
		lines = append(lines, renderLogLine(line))
	}
	if len(lines) == 0 {
		lines = []string{helpStyle.Render("No output yet.")}
	}

	sections = append(sections, sectionTitleStyle.Render("Output")+"\n\n"+strings.Join(lines, "\n"))
	return strings.Join(sections, "\n\n")
}

func (m model) formView() string {
	modeLine := focusedLabel(m.focusIndex == 0, "Mode") + "  " + m.modeOptions[m.modeIndex] + "  " + helpStyle.Render("(←/→ to change)")
	targetLabel := "Target URL"
	if m.modeOptions[m.modeIndex] == modeDNS {
		targetLabel = "Target Domain"
	}
	targetLine := focusedLabel(m.focusIndex == 1, targetLabel) + "  " + m.targetInput.View()
	statusCodesLabel := "Include Status Codes"
	if m.modeOptions[m.modeIndex] == modeVhost {
		statusCodesLabel = "Exclude Status Codes"
	}
	statusCodesLine := focusedLabel(m.focusIndex == 2, statusCodesLabel) + "  " + m.statusCodesInput.View()
	threadsLine := focusedLabel(m.focusIndex == 3, "Threads") + "  " + m.threadsInput.View()
	customDNSServerLine := focusedLabel(m.focusIndex == 4, "Custom DNS Server") + "  " + m.customDNSServerInput.View()
	wordlistFocusIndex := 4
	if m.modeOptions[m.modeIndex] == modeDNS {
		wordlistFocusIndex = 5
	}
	wordlistLine := focusedLabel(m.focusIndex == wordlistFocusIndex, "Wordlist Path") + "  " + m.wordlistInput.View()
	browseLine := focusedLabel(m.focusIndex == 7, "Browse Wordlist") + "  " + helpStyle.Render("Press Enter")
	startLine := startActionLabelStyled(m.stage == stageScanning, m.focusIndex == 8, m.gradientStep)
	help := helpStyle.Render("↑/↓ move fields • → fills placeholder • Tab/Shift+Tab switch panes • ctrl+c to quit")

	lines := []string{
		sectionTitleStyle.Render("Configuration"),
		"",
		modeLine,
		targetLine,
	}
	if m.modeOptions[m.modeIndex] != modeDNS {
		lines = append(lines, statusCodesLine)
	}
	if m.modeOptions[m.modeIndex] == modeDNS {
		lines = append(lines, customDNSServerLine)
	}
	lines = append(lines, threadsLine)
	lines = append(lines,
		wordlistLine,
		browseLine,
		"",
		startLine,
	)

	if m.scanFinished {
		outputFileLine := focusedLabel(m.focusIndex == 6, "Output Log File") + "  " + m.outputFileInput.View()
		lines = append(lines,
			"",
			outputFileLine,
			helpStyle.Render("Press Enter on Output Log File to save logs."),
		)
	}

	lines = append(lines,
		"",
		help,
	)

	return strings.Join(lines, "\n")
}

func startActionLabelStyled(scanning bool, active bool, gradientStep int) string {
	if scanning {
		return renderGradientText("▶ Scan in progress", gradientStep)
	}
	if active {
		return activeLabelStyle.Render("▶ Start Scan") + "  " + helpStyle.Render("Press Enter")
	}
	return inactiveLabelStyle.Render("  Start Scan") + "  " + helpStyle.Render("Press Enter")
}

func (m model) browseView() string {
	var b strings.Builder
	b.WriteString(sectionTitleStyle.Render("Browse Wordlist") + "\n\n")
	b.WriteString(helpStyle.Render("Use ↑/↓ and Enter to navigate/select • esc to return") + "\n\n")
	if m.pickerErr != nil {
		b.WriteString(errorStyle.Render(m.pickerErr.Error()))
		b.WriteString("\n\n")
	}
	b.WriteString(m.picker.View())
	return b.String()
}

func (m model) scanView() string {
	status := sectionTitleStyle.Render("Running gobuster " + m.modeOptions[m.modeIndex] + " scan")
	if m.errorMsg != "" {
		status += "\n\n" + errorStyle.Render(m.errorMsg)
	}

	spinnerLine := lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.spinner.View(),
		" ",
		helpStyle.Render("Scanning..."),
	)

	return strings.Join([]string{
		status,
		"",
		spinnerLine,
		"",
		helpStyle.Render("Press ctrl+c to stop and quit."),
	}, "\n")
}

func (m model) doneView() string {
	status := successStyle.Render("Scan complete.")
	if m.errorMsg != "" {
		status = errorStyle.Render("Scan finished with an error: " + m.errorMsg)
	}

	lines := []string{
		sectionTitleStyle.Render("Result"),
		"",
		status,
	}
	if m.logSaveError != "" {
		lines = append(lines, "", errorStyle.Render("Log save error: "+m.logSaveError))
	}
	lines = append(lines,
		"",
		helpStyle.Render("Configuration is editable again. Press Enter on Output Log File to save logs."),
	)
	return strings.Join(lines, "\n")
}

func focusedLabel(active bool, label string) string {
	if active {
		return activeLabelStyle.Render("▶ " + label)
	}
	return inactiveLabelStyle.Render("  " + label)
}
