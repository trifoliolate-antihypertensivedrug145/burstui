package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.stage {
		case stageForm:
			if m.resultFocused {
				return m.updateResultPane(msg)
			}
			return m.updateForm(msg)
		case stageBrowse:
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			if msg.String() == "esc" {
				m.stage = stageForm
				m.syncFocus()
				return m, nil
			}
		case stageScanning:
			if m.resultFocused {
				return m.updateResultPane(msg)
			}
			return m.updateScanningKeys(msg)
		}
	case clearErrorMsg:
		m.pickerErr = nil
		return m, nil
	case logLineMsg:
		line := string(msg)
		m.appendLog(line)
		if current, total, ok := extractProgressCounts(line); ok {
			m.progressCurrent = current
			m.progressTotal = total
		}
		return m, nil
	case scanProgressMsg:
		return m, nil
	case scanCountMsg:
		m.progressCurrent = msg.current
		m.progressTotal = msg.total
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if m.stage == stageScanning {
			if m.viewportReady {
				m.viewport.SetContent(m.rightPanelView())
			}
			return m, tea.Batch(cmd, m.spinner.Tick)
		}
		return m, cmd
	case tickMsg:
		m.gradientStep = (m.gradientStep + 1) % 6
		return m, tickCmd()
	case scanFinishedMsg:
		m.scanFinished = true
		m.stage = stageForm
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
			m.lastStatus = "Last scan finished with an error."
			m.appendLog("Scan failed: " + msg.err.Error())
		} else {
			m.errorMsg = ""
			m.lastStatus = "Last scan completed successfully."
			m.appendLog("Scan finished successfully.")
		}
		if m.progressTotal > 0 {
			m.progressCurrent = m.progressTotal
		}
		m.cmd = nil
		m.focusIndex = 4
		m.syncFocus()
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		rightWidth := 56
		if m.width > 0 {
			usable := m.width - 6
			if usable > 40 {
				rightWidth = usable - (usable / 2)
			}
		}

		viewportWidth := rightWidth - 4
		viewportHeight := m.height - 12
		if viewportHeight < 8 {
			viewportHeight = 8
		}

		browseHeight := m.height - 14
		if browseHeight < 8 {
			browseHeight = 8
		}
		m.picker.SetHeight(browseHeight)

		if !m.viewportReady {
			m.viewport = viewport.New(viewportWidth, viewportHeight)
			m.viewportReady = true
		} else {
			m.viewport.Width = viewportWidth
			m.viewport.Height = viewportHeight
		}
		m.viewport.SetContent(m.rightPanelView())

		return m, nil
	}

	if m.stage == stageBrowse {
		var cmd tea.Cmd
		m.picker, cmd = m.picker.Update(msg)

		if didSelect, path := m.picker.DidSelectFile(msg); didSelect {
			m.wordlistInput.SetValue(path)
			m.pickerErr = nil
			m.appendLog("Selected wordlist: " + path)
			m.stage = stageForm
			m.focusIndex = 2
			m.syncFocus()
			return m, nil
		}

		if didSelect, path := m.picker.DidSelectDisabledFile(msg); didSelect {
			m.pickerErr = errors.New(path + " is not valid.")
			return m, tea.Batch(cmd, clearErrorAfter(2*time.Second))
		}

		return m, cmd
	}

	m.viewport.SetContent(m.rightPanelView())
	return m, nil
}

func (m model) updateResultPane(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "left", "esc":
		m.resultFocused = false
		return m, nil
	case "tab", "shift+tab":
		m.resultFocused = false
		return m, nil
	}

	m.viewport, cmd = m.viewport.Update(msg)
	if m.stage == stageScanning {
		m.autoFollowLog = m.viewport.ScrollPercent() >= 0.999
	}
	return m, cmd
}

func (m model) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "tab", "shift+tab":
		m.resultFocused = true
		return m, nil
	case "up", "down":
		order := []int{0, 1, 2, 3, 4, 7, 8, 6}
		if m.modeOptions[m.modeIndex] == modeDNS {
			order = []int{0, 1, 4, 3, 5, 7, 8, 6}
		}
		if !m.scanFinished {
			if m.modeOptions[m.modeIndex] == modeDNS {
				order = []int{0, 1, 4, 3, 5, 7, 8}
			} else {
				order = []int{0, 1, 2, 3, 4, 7, 8}
			}
		}

		currentPos := 0
		for i, idx := range order {
			if idx == m.focusIndex {
				currentPos = i
				break
			}
		}

		if msg.String() == "up" {
			currentPos--
		} else {
			currentPos++
		}

		if currentPos < 0 {
			currentPos = len(order) - 1
		}
		if currentPos >= len(order) {
			currentPos = 0
		}

		m.focusIndex = order[currentPos]
		m.syncFocus()
		return m, nil
	case "left", "right":
		if m.focusIndex == 0 {
			if msg.String() == "left" {
				m.modeIndex--
			} else {
				m.modeIndex++
			}
			if m.modeIndex < 0 {
				m.modeIndex = len(m.modeOptions) - 1
			}
			if m.modeIndex >= len(m.modeOptions) {
				m.modeIndex = 0
			}
			if m.modeOptions[m.modeIndex] == modeVhost {
				m.statusCodesInput.Placeholder = "400"
			} else {
				m.statusCodesInput.Placeholder = "200"
			}
			if m.modeOptions[m.modeIndex] == modeDNS {
				m.targetInput.Placeholder = "example.com"
			} else {
				m.targetInput.Placeholder = "http://"
			}
			return m, nil
		}
		if msg.String() == "right" {
			if m.focusIndex == 1 && strings.TrimSpace(m.targetInput.Value()) == "" {
				m.targetInput.SetValue(m.targetInput.Placeholder)
				return m, nil
			}
			if m.focusIndex == 2 && m.modeOptions[m.modeIndex] != modeDNS && strings.TrimSpace(m.statusCodesInput.Value()) == "" {
				m.statusCodesInput.SetValue(m.statusCodesInput.Placeholder)
				return m, nil
			}
			if m.focusIndex == 3 && strings.TrimSpace(m.threadsInput.Value()) == "" {
				m.threadsInput.SetValue(m.threadsInput.Placeholder)
				return m, nil
			}
			if m.focusIndex == 4 && m.modeOptions[m.modeIndex] == modeDNS && strings.TrimSpace(m.customDNSServerInput.Value()) == "" {
				m.customDNSServerInput.SetValue(m.customDNSServerInput.Placeholder)
				return m, nil
			}
			if ((m.modeOptions[m.modeIndex] == modeDNS && m.focusIndex == 5) || (m.modeOptions[m.modeIndex] != modeDNS && m.focusIndex == 4)) && strings.TrimSpace(m.wordlistInput.Value()) == "" {
				m.wordlistInput.SetValue(m.wordlistInput.Placeholder)
				return m, nil
			}
			if m.scanFinished && m.focusIndex == 6 && strings.TrimSpace(m.outputFileInput.Value()) == "" {
				m.outputFileInput.SetValue(m.outputFileInput.Placeholder)
				return m, nil
			}

			atFarRight := false
			switch m.focusIndex {
			case 1:
				atFarRight = m.targetInput.Position() >= len(m.targetInput.Value())
			case 2:
				atFarRight = m.modeOptions[m.modeIndex] != modeDNS && m.statusCodesInput.Position() >= len(m.statusCodesInput.Value())
			case 3:
				atFarRight = m.threadsInput.Position() >= len(m.threadsInput.Value())
			case 4:
				if m.modeOptions[m.modeIndex] == modeDNS {
					atFarRight = m.customDNSServerInput.Position() >= len(m.customDNSServerInput.Value())
				} else {
					atFarRight = m.wordlistInput.Position() >= len(m.wordlistInput.Value())
				}
			case 5:
				if m.modeOptions[m.modeIndex] == modeDNS {
					atFarRight = m.wordlistInput.Position() >= len(m.wordlistInput.Value())
				}
			case 6:
				atFarRight = m.scanFinished && m.outputFileInput.Position() >= len(m.outputFileInput.Value())
			case 7, 8:
				atFarRight = true
			}

			if atFarRight {
				m.resultFocused = true
				return m, nil
			}
		}
	case "enter":
		switch m.focusIndex {
		case 7:
			m.stage = stageBrowse
			m.appendLog("Browse mode opened. Use arrow keys and Enter to select a file. Press esc to return.")
			return m, m.picker.Init()
		case 8:
			if m.stage == stageScanning {
				return m, nil
			}
			if strings.TrimSpace(m.targetInput.Value()) == "" {
				m.appendLog("Please enter a target URL before starting.")
				return m, nil
			}
			if strings.TrimSpace(m.wordlistInput.Value()) == "" {
				m.appendLog("Please enter or select a wordlist file before starting.")
				return m, nil
			}
			return m.startScan()
		case 6:
			if m.scanFinished {
				return m.saveOutputLog()
			}
		}
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	if m.focusIndex == 1 {
		m.targetInput, cmd = m.targetInput.Update(msg)
		cmds = append(cmds, cmd)
	}
	if m.focusIndex == 2 {
		m.statusCodesInput, cmd = m.statusCodesInput.Update(msg)
		cmds = append(cmds, cmd)
	}
	if m.focusIndex == 3 {
		m.threadsInput, cmd = m.threadsInput.Update(msg)
		cmds = append(cmds, cmd)
	}
	if m.focusIndex == 4 {
		if m.modeOptions[m.modeIndex] == modeDNS {
			m.customDNSServerInput, cmd = m.customDNSServerInput.Update(msg)
		} else {
			m.wordlistInput, cmd = m.wordlistInput.Update(msg)
		}
		cmds = append(cmds, cmd)
	}
	if m.focusIndex == 5 && m.modeOptions[m.modeIndex] == modeDNS {
		m.wordlistInput, cmd = m.wordlistInput.Update(msg)
		cmds = append(cmds, cmd)
	}
	if m.focusIndex == 6 && m.scanFinished {
		m.outputFileInput, cmd = m.outputFileInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) updateScanningKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.stopScan()
		m.quiting = true
		return m, tea.Quit
	case "left":
		m.resultFocused = false
		return m, nil
	case "right":
		m.resultFocused = true
		return m, nil
	}
	return m, nil
}

func (m *model) syncFocus() {
	if m.focusIndex == 1 {
		m.targetInput.Focus()
	} else {
		m.targetInput.Blur()
	}
	if m.focusIndex == 2 && m.modeOptions[m.modeIndex] != modeDNS {
		m.statusCodesInput.Focus()
	} else {
		m.statusCodesInput.Blur()
	}
	if m.focusIndex == 3 {
		m.threadsInput.Focus()
	} else {
		m.threadsInput.Blur()
	}
	if m.focusIndex == 4 && m.modeOptions[m.modeIndex] == modeDNS {
		m.customDNSServerInput.Focus()
	} else {
		m.customDNSServerInput.Blur()
	}
	if (m.focusIndex == 4 && m.modeOptions[m.modeIndex] != modeDNS) || (m.focusIndex == 5 && m.modeOptions[m.modeIndex] == modeDNS) {
		m.wordlistInput.Focus()
	} else {
		m.wordlistInput.Blur()
	}
	if m.scanFinished && m.focusIndex == 6 {
		m.outputFileInput.Focus()
	} else {
		m.outputFileInput.Blur()
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func extractPercent(line string) (float64, bool) {
	match := progressRE.FindStringSubmatch(line)
	if len(match) < 2 {
		return 0, false
	}
	var pct float64
	_, err := fmt.Sscanf(match[1], "%f", &pct)
	if err != nil {
		return 0, false
	}
	return clamp(pct/100.0, 0, 1), true
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
