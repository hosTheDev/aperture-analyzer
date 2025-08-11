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

	"time"
)

// We define a new command that waits for a packet message from a channel.
func waitForPacket(c <-chan capture.PacketInfo) tea.Cmd {
	return func() tea.Msg {
		return packetMsg(<-c)
	}
}

// packetMsg is the message we'll send when a new packet is captured.
type packetMsg capture.PacketInfo

// model holds the entire state of our TUI application.
type model struct {
	viewport   viewport.Model
	ready      bool
	packets    []string
	packetChan <-chan capture.PacketInfo

	startTime   time.Time
	packetCount int
	tcpCount    int
	udpCount    int
	otherCount  int
	totalBytes  int
}

// Start is the main entry point for the TUI.
func Start(packetChan chan capture.PacketInfo) {
	m := model{
		packets:    make([]string, 0),
		packetChan: packetChan,
		startTime:  time.Now(),
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
		m.packetCount++
		m.totalBytes += msg.Length
		switch msg.Protocol {
		case "TCP":
			m.tcpCount++
		case "UDP":
			m.udpCount++
		default:
			m.otherCount++
		}

		m.packets = append(m.packets, msg.Summary)
		m.viewport.SetContent(strings.Join(m.packets, "\n"))
		m.viewport.GotoBottom()
		return m, waitForPacket(m.packetChan)
		// m.packets = append(m.packets, string(msg))
		// m.viewport.SetContent(strings.Join(m.packets, "\n"))
		// m.viewport.GotoBottom()
		// // After processing a packet, immediately listen for the next one.
		// return m, waitForPacket(m.packetChan)
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

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// --- Panel Styling ---
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2)

	// --- Stats Panel Content ---
	elapsed := time.Since(m.startTime)
	pps := 0.0
	if elapsed.Seconds() > 0 {
		pps = float64(m.packetCount) / elapsed.Seconds()
	}

	totalData := float64(m.totalBytes)
	dataUnit := "B"
	if totalData > 1024 {
		totalData /= 1024
		dataUnit = "KB"
	}
	if totalData > 1024 {
		totalData /= 1024
		dataUnit = "MB"
	}

	stats := fmt.Sprintf(
		"Time Elapsed: %s\n\nTotal Packets: %d\nPackets/sec: %.2f\n\nData Captured: %.2f %s\n\nProtocol Breakdown:\n  TCP: %d\n  UDP: %d\n  Other: %d",
		elapsed.Round(time.Second),
		m.packetCount,
		pps,
		totalData,
		dataUnit,
		m.tcpCount,
		m.udpCount,
		m.otherCount,
	)

	// --- Layout ---
	// Set the width of the stats panel
	statsPanelWidth := 30
	// The packet list viewport takes the remaining space
	m.viewport.Width = m.viewport.Width - statsPanelWidth - 3 // Adjust for border/padding

	statsPanel := panelStyle.
		Width(statsPanelWidth).
		Render(stats)

	packetPanel := panelStyle.
		Width(m.viewport.Width).
		Render(m.viewport.View())

	// Join the panels horizontally
	mainView := lipgloss.JoinHorizontal(lipgloss.Top, packetPanel, statsPanel)

	// --- Header and Footer ---
	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("Aperture Packet Capture")
	footer := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render(fmt.Sprintf("Press 'q' to quit"))

	return fmt.Sprintf("%s\n%s\n%s", header, mainView, footer)
}

// View renders the UI.
// func (m model) View() string {
// 	if !m.ready {
// 		return "Initializing..."
// 	}
// 	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
// 	footerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))

// 	header := headerStyle.Render("Aperture Packet Capture")
// 	footer := footerStyle.Render(fmt.Sprintf("Total Packets: %d | Press 'q' to quit", len(m.packets)))

// 	return fmt.Sprintf("%s\n%s\n%s", header, m.viewport.View(), footer)
// }

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