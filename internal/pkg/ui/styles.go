package ui

import "github.com/charmbracelet/lipgloss"

const (
	ColorBg        = "#1a1b26"
	ColorFg        = "#a9b1d6"
	ColorHighlight = "#7aa2f7"
	ColorSelected  = "#364a82"
	ColorAccent    = "#bb9af7"
	ColorMono      = "#e4e0d7"
	ColorStack     = "#f5c542"
	ColorSuccess   = "#9ece6a"
	ColorError     = "#f7768e"
	ColorWarning   = "#e0af68"
	ColorSubtle    = "#565f89"
	ColorBorder    = "#414868"
	ColorCyan      = "#7dcfff"
	ColorOrange    = "#ff9e64"
	ColorIndigo    = "#6366f1"
	ColorRose      = "#fb7185"
	ColorEmerald   = "#34d399"
	ColorAmber     = "#fbbf24"
)

const (
	IconAhead  = "↑"
	IconBehind = "↓"
	IconDirty  = "✗"
	IconClean  = "✓"
	IconSpace  = " "
	IconFetch  = "⟳"
	IconBullet = "•"
)

var SpinnerFrames = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}

var (
	ColorBgLip        = lipgloss.Color(ColorBg)
	ColorFgLip        = lipgloss.Color(ColorFg)
	ColorHighlightLip = lipgloss.Color(ColorHighlight)
	ColorAccentLip    = lipgloss.Color(ColorAccent)
	ColorSuccessLip   = lipgloss.Color(ColorSuccess)
	ColorErrorLip     = lipgloss.Color(ColorError)
	ColorSubtleLip    = lipgloss.Color(ColorSubtle)
	ColorBorderLip    = lipgloss.Color(ColorBorder)
	ColorCyanLip      = lipgloss.Color(ColorCyan)
	ColorWarningLip   = lipgloss.Color(ColorWarning)
	ColorMonoLip      = lipgloss.Color(ColorMono)
	ColorStackLip     = lipgloss.Color(ColorStack)

	BrandMonoStyle = lipgloss.NewStyle().
			Foreground(ColorMonoLip).
			Bold(true)
	BrandStackStyle = lipgloss.NewStyle().
			Foreground(ColorStackLip).
			Bold(true)
)
