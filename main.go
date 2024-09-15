package main

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"os"
	"strconv"
	"strings"
)

func newModel() model {
	betInp := textinput.New()
	betInp.Width = 25
	//betInp.Placeholder = "gamba amount"
	betInp.CompletionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("32"))
	betInp.Validate = func(s string) error {
		if strings.TrimSpace(s) == "" {
			return errors.New("no bet no gamba")
		}

		amount, err := strconv.Atoi(s)
		if err != nil {
			return errors.New("we don't gamble with words")
		} else if amount == 0 {
			return errors.New("no gamba without money")
		} else if amount < 0 {
			return errors.New("gamba without money?")
		}
		assert(err == nil, "have to have a valid gamba amount")
		return nil
	}

	return model{
		bet:       None,
		player:    Player(25),
		betAmount: 1,
		betInp:    betInp,
	}
}

type model struct {
	focused       int
	width, height int

	// fields for placing the bet
	bet       Bet
	betcursor int
	betErr    error

	// fields for setting bet amount
	player    Player
	betAmount uint
	betInp    textinput.Model

	isGamba bool
	game    Result
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (model tea.Model, cmd tea.Cmd) {
	if m.player.HasLost() {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.focused++
			m.focused %= 3
			return m, nil
		case "shift+tab":
			m.focused--
			m.focused %= 3
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	switch m.focused {
	case 0:
		return m.takeBet(msg)
	case 1:
		m.betInp.Focus()
		m, cmd = m.placeBet(msg)
		m.betInp.Blur()
		return m, cmd
	case 2:
		return m.gamba(msg)
	default:
		panic("bad focus: " + string(m.focused))
	}
}

var bgStyle = lipgloss.NewStyle().Background(lipgloss.Color("34"))
var errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("160"))
var activeColor = color.New(color.FgCyan)

const (
	totalWidth = 80
	leftWidth  = 30
	rightWidth = totalWidth - leftWidth - 4
)

func (m model) View() string {
	withBorder := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	leftSide := withBorder.Width(leftWidth)

	var out strings.Builder
	out.WriteString(
		withBorder.Width(totalWidth).Render(
			lipgloss.JoinVertical(lipgloss.Left,
				lipgloss.Place(totalWidth, 7, lipgloss.Center, lipgloss.Center,
					lipgloss.NewStyle().Foreground(lipgloss.Color("34")).Width(62).Render(
						`
                                              ____
███████╗███████╗██╗   ██╗███████╗███╗   ██╗  /\' .\    _____
██╔════╝██╔════╝██║   ██║██╔════╝████╗  ██║ /: \___\  / .  /\
███████╗█████╗  ██║   ██║█████╗  ██╔██╗ ██║ \' / . / /____/..\
╚════██║██╔══╝  ╚██╗ ██╔╝██╔══╝  ██║╚██╗██║  \/___/  \'  '\  /
███████║███████╗ ╚████╔╝ ███████╗██║ ╚████║           \'__'\/
╚══════╝╚══════╝  ╚═══╝  ╚══════╝╚═╝  ╚═══╝
`),
				),

				lipgloss.NewStyle().MarginTop(2).MarginLeft(3).MarginRight(3).Render(
					lipgloss.JoinHorizontal(lipgloss.Left,
						lipgloss.JoinVertical(lipgloss.Center,
							leftSide.Render(m.renderTakeBet()),
							leftSide.Render(m.renderPlaceBet()),
						),
						lipgloss.Place(rightWidth, 15, lipgloss.Center, lipgloss.Center,
							m.renderGamba(),
						),
					),
				),
			),
		),
	)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center,
		lipgloss.Center,

		out.String(),
	)
}

func (m model) takeBet(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.betcursor > 0 {
				m.betcursor--
			}
		case "down", "j":
			if m.betcursor < len(Bets)-1 {
				m.betcursor++
			}
		case "enter", " ":
			m.bet = Bets[m.betcursor]
			m.betErr = nil
		}
	}
	return m, nil
}

func (m model) renderTakeBet() string {
	bet := "What are you betting for?\n\n"
	for i, choice := range Bets {
		cursor := " "
		if m.betcursor == i {
			cursor = ">"
		}
		checked := " "
		if choice == m.bet {
			checked = "x"
		}
		row := fmt.Sprintf("%s [%s] %s", cursor, checked, choice.String())
		if m.betcursor == i && m.focused == 0 {
			bet += activeColor.Sprintf(row)
		} else {
			bet += row
		}
		bet += "\n"
	}
	if m.betErr != nil {
		bet += errStyle.Render(fmt.Sprintf("  *%v", m.betErr))
	} else {
		bet += ""
	}
	return bet
}

func (m model) placeBet(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	m.betInp, cmd = m.betInp.Update(msg)
	if m.betInp.Err == nil {
		val := m.betInp.Value()
		if val == "" { // checking for empty input, will cause ParseError otherwise
			return m, cmd
		}
		amount, err := strconv.Atoi(val)

		if !m.player.CanBet(uint(amount)) { // check if player has enough funds
			m.betInp.Err = fmt.Errorf("you have %d money", m.player)
			return m, cmd
		}
		assert(err == nil, "bad amount got validated")
		assert(amount != 0, "zero amount placed; "+val)
		m.betAmount = uint(amount)
	}
	return m, cmd
}

func (m model) renderPlaceBet() string {
	inp := m.betInp
	betted := inp.View()

	if m.focused == 1 {
		betted = activeColor.Sprint(betted)
	}
	betted += "\n"

	if inp.Err != nil {
		betted += errStyle.Render("  *" + inp.Err.Error())
	}

	return betted
}

func (m model) gamba(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			var bad bool

			// if no bet has been placed yet
			if m.bet == None {
				m.betErr = errors.New("pls bet on something")
				bad = true
			}

			// if the bet amount is invalid
			if m.betInp.Err != nil {
				return m, nil
			}
			// if no bet amount has been placed
			if m.betInp.Value() == "" {
				m.betInp.Err = errors.New("please place a bet")
				bad = true
			}

			// How to do this?
			if !bad {
				m.game, m.player = m.player.playRound(m.bet, m.betAmount)
			}
			return m, nil
		}
	}
	return m, nil
}

func (m model) renderGamba() string {
	msg := color.HiGreenString("You won!")
	if m.game.Lost {
		msg = color.HiRedString("You Lost!")
	}
	msg = fmt.Sprintf("   %10s", msg)

	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Center,
			func(prev lipgloss.Style, active bool) lipgloss.Style {
				if active {
					activeColor := lipgloss.Color("#00E6E6")
					prev = prev.Border(lipgloss.ThickBorder()).BorderForeground(activeColor).Foreground(activeColor)
				} else {
					prev = prev.Border(lipgloss.ThickBorder())
				}
				return prev
			}(lipgloss.NewStyle(), m.focused == 2).Render(
				"gamba",
			),
			msg,
		),
		fmt.Sprintf(`
 Dice 1 | Dice 2 | Total
--------+--------+-------
 %2d     | %2d     | %2d    `, m.game.DiceA, m.game.DiceB, m.game.DiceA+m.game.DiceB),
		lipgloss.NewStyle().MarginTop(1).Render(
			color.New(color.FgYellow).Sprintf("You have %d money", m.player),
		),
	)
}

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func assert(statment bool, msg string) {
	if !statment {
		panic(msg)
	}
}
