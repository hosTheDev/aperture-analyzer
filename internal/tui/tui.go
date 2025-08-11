package tui

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/hosTheDev/aperture-analyzer/internal/capture"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// We define a new command that waits for a packet message from a channel.
func waitForPacket(c <-chan string) tea.Cmd {
	return func() tea.Msg {
		return packetMsg(<-c)
	}
}

// packetMsg is the message we'll send when a new packet is captured.
type packetMsg string

// model holds the entire state of our TUI application.
type model struct {
	viewport   viewport.Model
	ready      bool
	packets    []string
	packetChan <-chan string
}

// Start is the main entry point for the TUI.
func Start(packetChan chan string) {
	m := model{
		packets:    make([]string, 0),
		packetChan: packetChan,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}

// Init is the first command that's run. We'll start listening for packets here.
func (m model) Init() tea.Cmd {
	return waitForPacket(m.packetChan)
}

// Update handles all events and updates the model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case packetMsg:
		m.packets = append(m.packets, string(msg))
		m.viewport.SetContent(strings.Join(m.packets, "\n"))
		m.viewport.GotoBottom()
		// After processing a packet, immediately listen for the next one.
		return m, waitForPacket(m.packetChan)
	case tea.WindowSizeMsg:
		headerHeight, footerHeight := 3, 3
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-(headerHeight+footerHeight))
			m.viewport.YPosition = headerHeight
			m.viewport.SetContent("Waiting for packets...")
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - (headerHeight + footerHeight)
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// View renders the UI.
func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	footerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))

	header := headerStyle.Render("Aperture Packet Capture")
	footer := footerStyle.Render(fmt.Sprintf("Total Packets: %d | Press 'q' to quit", len(m.packets)))

	return fmt.Sprintf("%s\n%s\n%s", header, m.viewport.View(), footer)
}

// PromptForDeviceSelection handles the initial CLI prompt to choose a device.
func PromptForDeviceSelection(devices []capture.DeviceInfo) (capture.DeviceInfo, error) {
	fmt.Printf("\n[Aperture] Success! Found %d network devices:\n", len(devices))
	fmt.Println("-------------------------------------------------")
	for i, device := range devices {
		fmt.Printf("  %d: %s\n    %s\n", i+1, device.Name, device.Description)
	}
	fmt.Println("-------------------------------------------------")
	for {
		fmt.Printf("\nEnter the number of the device to start analysis (1-%d): ", len(devices))
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		choice, err := strconv.Atoi(input)
		if err == nil && choice >= 1 && choice <= len(devices) {
			return devices[choice-1], nil
		}
		fmt.Printf("Invalid input. Please enter a number between 1 and %d.\n", len(devices))
	}
}