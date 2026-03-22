package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) startScan() (tea.Model, tea.Cmd) {
	target := strings.TrimSpace(m.targetInput.Value())
	statusCodes := strings.TrimSpace(m.statusCodesInput.Value())
	threads := strings.TrimSpace(m.threadsInput.Value())
	customDNSServer := strings.TrimSpace(m.customDNSServerInput.Value())
	wordlistPath := strings.TrimSpace(m.wordlistInput.Value())
	wordCount, wordCountErr := countWordlistLines(wordlistPath)
	if target == "" {
		m.appendLog("Please enter a target URL before starting.")
		return m, nil
	}
	if wordlistPath == "" {
		m.appendLog("Please enter or select a wordlist file before starting.")
		return m, nil
	}
	if statusCodes == "" {
		if m.modeOptions[m.modeIndex] == modeVhost {
			statusCodes = "400"
		} else {
			statusCodes = "200"
		}
	}
	if threads == "" {
		threads = m.threadsInput.Placeholder
	}

	m.stage = stageScanning
	m.resultFocused = true
	m.autoFollowLog = true
	if m.viewportReady {
		m.viewport.GotoBottom()
	}
	m.scanFinished = false
	m.progressCurrent = 0
	m.progressTotal = wordCount
	m.errorMsg = ""
	m.logSaveError = ""
	m.lastStatus = "Scan currently in progress."
	m.logs = []string{
		"Starting gobuster...",
		"Mode: " + m.modeOptions[m.modeIndex],
		"Threads: " + threads,
		"Wordlist: " + wordlistPath,
	}
	if m.modeOptions[m.modeIndex] == modeDNS {
		m.logs = append(m.logs, "Domain: "+target)
		if customDNSServer != "" && customDNSServer != m.customDNSServerInput.Placeholder {
			m.logs = append(m.logs, "Custom DNS server: "+customDNSServer)
		}
	} else {
		m.logs = append(m.logs, "Target: "+target)
		if m.modeOptions[m.modeIndex] == modeVhost {
			m.logs = append(m.logs, "Append domain: true")
			m.logs = append(m.logs, "Excluded status codes: "+statusCodes)
		} else {
			m.logs = append(m.logs, "Status codes: "+statusCodes)
		}
	}
	if wordCountErr == nil && wordCount > 0 {
		m.logs = append(m.logs, fmt.Sprintf("Wordlist entries: %d", wordCount))
	} else if wordCountErr != nil {
		m.logs = append(m.logs, "Could not count wordlist entries: "+wordCountErr.Error())
	}

	args := []string{m.modeOptions[m.modeIndex], "-u", target, "-w", wordlistPath, "-s", statusCodes, "-t", threads, "--status-codes-blacklist", ""}
	if m.modeOptions[m.modeIndex] == modeVhost {
		args = []string{m.modeOptions[m.modeIndex], "-u", target, "-w", wordlistPath, "-t", threads, "-xs", statusCodes, "--append-domain"}
	}
	if m.modeOptions[m.modeIndex] == modeDNS {
		args = []string{m.modeOptions[m.modeIndex], "-do", target, "-w", wordlistPath, "-t", threads}
		if customDNSServer != "" && customDNSServer != m.customDNSServerInput.Placeholder {
			args = append(args, "--resolver", customDNSServer)
		}
	}

	cmd := exec.Command("gobuster", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		m.stage = stageForm
		m.errorMsg = err.Error()
		m.lastStatus = "Failed to prepare scan output pipes."
		m.appendLog("Failed to prepare stdout: " + err.Error())
		return m, nil
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		m.stage = stageForm
		m.errorMsg = err.Error()
		m.lastStatus = "Failed to prepare scan error output."
		m.appendLog("Failed to prepare stderr: " + err.Error())
		return m, nil
	}
	if err := cmd.Start(); err != nil {
		m.stage = stageForm
		m.errorMsg = err.Error()
		m.lastStatus = "Failed to start gobuster."
		m.appendLog("Failed to start gobuster: " + err.Error())
		return m, nil
	}

	m.cmd = cmd

	return m, tea.Batch(
		m.spinner.Tick,
		m.streamOutput(stdout),
		m.streamOutput(stderr),
		m.waitForScan(),
	)
}

func (m model) saveOutputLog() (tea.Model, tea.Cmd) {
	path := strings.TrimSpace(m.outputFileInput.Value())
	if path == "" {
		path = m.outputFileInput.Placeholder
		m.outputFileInput.SetValue(path)
	}

	if _, err := os.Stat(path); err == nil {
		m.logSaveError = "output file already exists"
		m.lastStatus = "Failed to save output log."
		m.appendLog("Failed to save output log: file already exists: " + path)
		return m, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		m.logSaveError = err.Error()
		m.lastStatus = "Failed to save output log."
		m.appendLog("Failed to check output file: " + err.Error())
		return m, nil
	}

	content := strings.Join(m.logs, "\n")
	if content != "" {
		content += "\n"
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		m.logSaveError = err.Error()
		m.lastStatus = "Failed to save output log."
		m.appendLog("Failed to save output log: " + err.Error())
		return m, nil
	}

	m.logSaveError = ""
	m.lastStatus = "Output log saved successfully."
	m.appendLog("Output log saved to " + path)
	return m, nil
}

func (m model) streamOutput(r io.Reader) tea.Cmd {
	return func() tea.Msg {
		scanner := bufio.NewScanner(r)
		scanner.Split(scanLinesOrCR)
		if scanner.Scan() {
			line := scanner.Text()
			cmds := make([]tea.Cmd, 0, 4)
			if current, total, ok := extractProgressCounts(line); ok {
				cmds = append(cmds, func() tea.Msg { return scanCountMsg{current: current, total: total} })
			}
			if p, ok := extractPercent(line); ok {
				cmds = append(cmds, func() tea.Msg { return scanProgressMsg(p) })
			}
			cmds = append(cmds, func() tea.Msg { return logLineMsg(line) })
			cmds = append(cmds, m.streamOutput(r))
			return tea.Batch(cmds...)()
		}
		if err := scanner.Err(); err != nil {
			return logLineMsg("Output error: " + err.Error())
		}
		return nil
	}
}

func scanLinesOrCR(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	for i, b := range data {
		if b == '\n' || b == '\r' {
			if i == 0 {
				return 1, nil, nil
			}
			return i + 1, data[:i], nil
		}
	}

	if atEOF {
		return len(data), data, nil
	}

	return 0, nil, nil
}

func extractProgressCounts(line string) (int, int, bool) {
	match := progressCountRE.FindStringSubmatch(line)
	if len(match) < 3 {
		return 0, 0, false
	}
	var current, total int
	if _, err := fmt.Sscanf(match[1], "%d", &current); err != nil {
		return 0, 0, false
	}
	if _, err := fmt.Sscanf(match[2], "%d", &total); err != nil {
		return 0, 0, false
	}
	if total <= 0 {
		return 0, 0, false
	}
	return current, total, true
}

func countWordlistLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			count++
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return count, nil
}

func (m model) waitForScan() tea.Cmd {
	cmd := m.cmd
	return func() tea.Msg {
		if cmd == nil {
			return scanFinishedMsg{err: fmt.Errorf("scan command was not initialized")}
		}
		err := cmd.Wait()
		return scanFinishedMsg{err: err}
	}
}

func (m *model) stopScan() {
	if m.cmd != nil && m.cmd.Process != nil {
		_ = m.cmd.Process.Kill()
	}
}

func (m *model) appendLog(line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	m.logs = append(m.logs, line)
	if len(m.logs) > maxLogLines {
		m.logs = m.logs[len(m.logs)-maxLogLines:]
	}

	if m.viewportReady {
		m.viewport.SetContent(m.rightPanelView())
		if m.stage == stageScanning && m.autoFollowLog {
			m.viewport.GotoBottom()
		}
	}
}
