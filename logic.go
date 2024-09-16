package main

import (
	"crypto/rand"
	"fmt"
	"github.com/fatih/color"
	"math/big"
	"strings"
)

type Result struct {
	DiceA, DiceB uint
	Lost         bool
}

func (p Player) playRound(bet Bet, betAmount uint) (Result, Player) {
	res := Result{}
	res.DiceA, res.DiceB = throwDices()
	p, won, _ := handleWinnings(p, bet, betAmount, res.DiceA+res.DiceB)
	res.Lost = !won
	return res, p
}

func throwDices() (uint, uint) {
	dice, err := rand.Int(rand.Reader, big.NewInt(6))
	dice2, err := rand.Int(rand.Reader, big.NewInt(6))
	if err != nil {
		panic("failed to generate random number: " + err.Error())
	}
	return uint(1 + dice.Int64()), uint(1 + dice2.Int64())
}

func handleWinnings(p Player, bet Bet, betAmount uint, res uint) (updatedP Player, won bool, gameover bool) {
	// bet the amount, removing it from the player
	p, ok := p.Bet(betAmount)
	if !ok {
		panic("inconsistent state, bet amount which player does not have funding for")
	}

	// check if bet was correct
	if matchingBet := ToBet(res); matchingBet == bet {
		switch matchingBet {
		case Seven:
			p = p.WinAmount(betAmount * 4)
		case Under, Over:
			p = p.WinAmount(betAmount * 2)
		}
		return p, true, false
	}

	// return player, which has lost and check if player has lost
	return p, false, p.HasLost()
}

type Bet int

const (
	Seven Bet = iota
	Under
	Over
	None
)

var Bets = []Bet{Under, Seven, Over}

var betNames = map[Bet]string{
	Seven: "seven",
	Under: "under",
	Over:  "over",
	None:  "none",
}

func (b Bet) String() string {
	name := betNames[b]
	return strings.ToUpper(string(name[0])) + name[1:]
}

func ToBet(sum uint) Bet {
	switch {
	case 0 < sum && sum < 7:
		return Under
	case 7 < sum && sum < 13:
		return Over
	case sum == 7:
		return Seven
	default:
		panic("bad sum, got: " + string(sum))
	}
}

func getBets() Bet {
	c := color.New().Add(color.Underline)
	for {
		var picked string

		c.Printf("\n\tSeven, Under or Over? [(s)even/(u)nder/(o)ver]\n\t")
		_, err := fmt.Scanln(&picked)
		if err != nil && err.Error() != "unexpected newline" {
			panic("program failed: failed to get user input:" + err.Error())
		}

		// converting, if none match redo
		switch strings.ToLower(picked) {
		case "seven", "s":
			return Seven
		case "under", "u":
			return Under
		case "over", "o":
			return Over
		default:
		}
	}
}

type Player uint

func (p Player) String() string {
	return string(p)
}

func (p Player) Bet(amount uint) (Player, bool) {
	if left := p - Player(amount); left >= 0 {
		return left, true
	}
	return p, false
}

func (p Player) CanBet(amount uint) bool {
	if Player(amount) > p {
		return false
	}

	if left := p - Player(amount); left >= 0 {
		return true
	}
	return false
}

func (p Player) WinAmount(amount uint) Player {
	return p + Player(amount)
}

func (p Player) HasLost() bool {
	return p == 0
}
