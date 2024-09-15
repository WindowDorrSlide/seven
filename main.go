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
	betInp.Placeholder = "gamba amount"
	betInp.CompletionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("32"))

	// set a validater function for the input
	betInp.Validate = func(s string) error {
		// check if no input
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

	game Result
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (model tea.Model, cmd tea.Cmd) {
	// check if player can play another round
	if m.player.HasLost() {
		return m, tea.Quit
	}

	// handle switching focus and quiting
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

	// let focused item handle functionality
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

var errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("160"))
var activeColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#804000"))

const (
	totalWidth = 80
	leftWidth  = 30
	rightWidth = totalWidth - leftWidth - 4
)

func (m model) View() string {
	withBorder := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	leftSide := withBorder.Width(leftWidth)

	// chaos
	var out strings.Builder
	out.WriteString(
		withBorder.Border(lipgloss.BlockBorder()).Width(totalWidth).Render(
			lipgloss.JoinVertical(lipgloss.Left,
				lipgloss.Place(
					totalWidth, 7,
					lipgloss.Center, lipgloss.Center,
					lipgloss.NewStyle().Foreground(lipgloss.Color("#005c78")).Width(62).Render(
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
					lipgloss.JoinHorizontal(lipgloss.Center,
						lipgloss.JoinVertical(lipgloss.Center,
							leftSide.Render(m.renderTakeBet(m.focused == 0)),
							leftSide.Render(m.renderPlaceBet(m.focused == 1)),
						),
						lipgloss.Place(rightWidth, 15, lipgloss.Center, lipgloss.Center,
							lipgloss.NewStyle().BorderForeground(lipgloss.Color("#f22453")).Render(
								m.renderGamba(m.focused == 2),
							),
						),
					),
				),
			),
		),
	)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
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

func (m model) renderTakeBet(active bool) string {
	bet := "What are you betting for?\n\n"

	// render all options
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
		if m.betcursor == i && active {
			bet += activeColor.Render(row)
		} else {
			bet += row
		}
		bet += "\n"
	}

	// render error if there is one
	if m.betErr != nil {
		bet += errStyle.Render(fmt.Sprintf("  *%v", m.betErr))
	}
	return bet
}

func (m model) placeBet(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	m.betInp, cmd = m.betInp.Update(msg)
	if m.betInp.Err != nil {
		return m, cmd
	}

	// getting value and converting it to int
	val := m.betInp.Value()
	if val == "" { // checking for empty input, will cause ParseError otherwise
		return m, cmd
	}
	amount, err := strconv.Atoi(val)

	// check if player has enough funds
	if !m.player.CanBet(uint(amount)) {
		m.betInp.Err = fmt.Errorf("you have %d money", m.player)
		return m, cmd
	}
	assert(err == nil, "bad amount got validated")
	assert(amount != 0, "zero amount placed; "+val)

	// update the amount betted
	m.betAmount = uint(amount)
	return m, cmd
}

func (m model) renderPlaceBet(active bool) string {
	// render the view
	betted := m.betInp.View()

	// change text color if it is active
	if active {
		betted = activeColor.Render(betted)
	}
	// add to item after so it doesnt get colored, causing chaos
	betted += "\n"

	// print error
	if m.betInp.Err != nil {
		betted += errStyle.Render("  *" + m.betInp.Err.Error())
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
				// check if player has enough funds, if the amount is not changed it won't validate
				// meaning the player could spend more then he has funding to
				if !m.player.CanBet(m.betAmount) {
					m.betInp.Err = fmt.Errorf("you have %d money", m.player)
					return m, nil
				}
				m.game, m.player = m.player.playRound(m.bet, m.betAmount)
			}
			return m, nil
		}
	}
	return m, nil
}

func (m model) renderGamba(active bool) string {
	// victory msg
	msg := color.HiGreenString("You won!")
	if m.game.DiceA == 0 { // If no game has been played yet
		msg = ""
	} else if m.game.Lost { // Set text if lost
		msg = color.HiRedString("You Lost!")
	}
	msg = fmt.Sprintf("   %10s", msg)

	// render
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Center,
			func(prev lipgloss.Style, active bool) lipgloss.Style {
				if active {
					activeColor := lipgloss.Color("#804000")
					prev = prev.Border(lipgloss.RoundedBorder()).BorderForeground(activeColor)
				} else {
					prev = prev.Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#f22453"))
				}
				return prev
			}(lipgloss.NewStyle().Foreground(lipgloss.Color("#f22453")), active).Render(
				"gamba",
			),
			msg,
		),
		fmt.Sprintf(`
 Dice A │ Dice B │ Total
────────┼────────┼───────
 %2d     │ %2d     │ %2d    `, m.game.DiceA, m.game.DiceB, m.game.DiceA+m.game.DiceB),
		lipgloss.Place(25, 1, lipgloss.Right, lipgloss.Center,
			lipgloss.NewStyle().MarginTop(1).Foreground(lipgloss.Color("#dbcf60")).Render(
				fmt.Sprintf("%d \u00A9", m.player),
			),
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
