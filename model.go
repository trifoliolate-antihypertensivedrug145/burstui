package main

import (
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	modeDir     = "dir"
	modeVhost   = "vhost"
	modeDNS     = "dns"
	maxLogLines = 200
)

type stage int

const (
	stageForm stage = iota
	stageBrowse
	stageScanning
)

type logLineMsg string
type scanFinishedMsg struct{ err error }
type scanProgressMsg float64
type scanCountMsg struct {
	current int
	total   int
}
type tickMsg time.Time
type clearErrorMsg struct{}

type model struct {
	stage       stage
	modeIndex   int
	modeOptions []string
	focusIndex  int

	targetInput          textinput.Model
	statusCodesInput     textinput.Model
	threadsInput         textinput.Model
	customDNSServerInput textinput.Model
	wordlistInput        textinput.Model
	outputFileInput      textinput.Model
	picker               filepicker.Model

	spinner      spinner.Model
	logs         []string
	errorMsg     string
	logSaveError string
	pickerErr    error

	cmd             *exec.Cmd
	quiting         bool
	width           int
	height          int
	lastStatus      string
	scanFinished    bool
	gradientStep    int
	progressCurrent int
	progressTotal   int
	resultFocused   bool
	autoFollowLog   bool
	viewport        viewport.Model
	viewportReady   bool
}

func initialModel() model {
	targetInput := textinput.New()
	targetInput.Placeholder = "http://"
	targetInput.Prompt = "> "
	targetInput.CharLimit = 512
	targetInput.Width = 60

	wordlistInput := textinput.New()
	wordlistInput.Placeholder = "/usr/share/wordlists/dirb/common.txt"
	wordlistInput.Prompt = "> "
	wordlistInput.CharLimit = 1024
	wordlistInput.Width = 60

	statusCodesInput := textinput.New()
	statusCodesInput.Placeholder = "200"
	statusCodesInput.Prompt = "> "
	statusCodesInput.CharLimit = 256
	statusCodesInput.Width = 60

	threadsInput := textinput.New()
	threadsInput.Placeholder = "10"
	threadsInput.Prompt = "> "
	threadsInput.CharLimit = 32
	threadsInput.Width = 60

	customDNSServerInput := textinput.New()
	customDNSServerInput.Placeholder = "(optional)"
	customDNSServerInput.Prompt = "> "
	customDNSServerInput.CharLimit = 256
	customDNSServerInput.Width = 60

	outputFileInput := textinput.New()
	outputFileInput.Placeholder = "./burstui-output.log"
	outputFileInput.Prompt = "> "
	outputFileInput.CharLimit = 1024
	outputFileInput.Width = 60

	picker := filepicker.New()
	picker.CurrentDirectory = "/usr/share/wordlists"
	picker.ShowPermissions = false
	picker.ShowHidden = false
	picker.SetHeight(16)

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot

	vp := viewport.New(0, 0)
	vp.SetContent("")

	gobusterPath, gobusterVersion := detectGobusterInfo()

	m := model{
		stage:       stageForm,
		modeIndex:   0,
		modeOptions: []string{modeDir, modeVhost, modeDNS},
		focusIndex:  1,

		targetInput:          targetInput,
		statusCodesInput:     statusCodesInput,
		threadsInput:         threadsInput,
		customDNSServerInput: customDNSServerInput,
		wordlistInput:        wordlistInput,
		outputFileInput:      outputFileInput,
		picker:               picker,

		spinner: sp,
		logs: []string{
			"Gobuster path: " + gobusterPath,
			"Gobuster version: " + gobusterVersion + " (recommended >= 3.8.2)",
			"Ready. Fill out the form and press Enter on Start Scan.",
		},
		lastStatus:    "Idle.",
		viewport:      vp,
		autoFollowLog: false,
	}
	m.syncFocus()
	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.picker.Init(), m.spinner.Tick, tickCmd())
}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

func detectGobusterInfo() (string, string) {
	path, err := exec.LookPath("gobuster")
	if err != nil {
		return "not found in PATH", "unavailable"
	}

	output, err := exec.Command(path, "--version").CombinedOutput()
	version := strings.TrimSpace(string(output))
	parts := strings.SplitN(version, "\n", 2)
	version = strings.TrimSpace(parts[0])
	if err != nil {
		if version == "" {
			version = "unavailable"
		} else {
			version += " (version check failed)"
		}
	}
	if version == "" {
		version = "unavailable"
	}

	return path, version
}
