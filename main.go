package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

func main() {
	fmt.Println(`
Welcome to seven!
The rules are simple, you bet if the sum of two dice rolls is over/under or equal with seven.
You start with 25 money, your goal is to survive as many rounds as possible.
A correct bet on Over/Under gives you that amount while a bet on seven returns three times the bet amount.

Press Ctrl+C to exit.`)

	p, rounds := Player(25), 0
	for {
		bet := getBets()
		amount := getBettingAmount(p)
		dices := throwDices()
		fmt.Printf("\n\tDice result: %d\n\t", dices)

		var won, gameOver bool
		if p, won, gameOver = handleWinnings(p, bet, amount, dices); won {
			color.Green("\n\tWon!")
		} else {
			color.Red("\n\tLost!")
		}

		if gameOver {
			fmt.Printf("\nYou have no funding left, you lose!\nYou have survived %d rounds.", rounds)
			return
		}

		rounds++
	}

}

func throwDices() uint {
	dice, err := rand.Int(rand.Reader, big.NewInt(6))
	dice2, err := rand.Int(rand.Reader, big.NewInt(6))
	if err != nil {
		panic("failed to generate random number: " + err.Error())
	}
	return uint(2 + dice.Int64() + dice2.Int64())
}

func handleWinnings(p Player, bet Bet, betAmount uint, res uint) (updatedP Player, won bool, gameover bool) {
	// bet the amount, removing it from the player
	p, ok := p.Bet(betAmount)
	if !ok {
		panic("inconsistent state, bet amount which player does not have funding for")
	}

	// check if bet was correct
	if matchingBet := checkResult(res); matchingBet == bet {
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

func checkResult(sum uint) Bet {
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

type Bet int

const (
	Seven Bet = iota
	Under
	Over
)

var betNames = map[Bet]string{
	Seven: "seven",
	Under: "under",
	Over:  "over",
}

func (b Bet) String() string {
	return betNames[b]
}

func getBettingAmount(p Player) uint {
	c := color.New().Add(color.Underline)
	for {
		var picked string

		c.Printf("\n\n\t\tWhat amount do you want to bet? You've got %d money\n\t\t", p)
		_, err := fmt.Scanln(&picked)
		if err != nil && err.Error() != "unexpected newline" {
			panic("program failed: failed to get user input:" + err.Error())
		}

		amount, err := strconv.Atoi(picked)
		switch {
		case err != nil:
		case amount <= 0:
			fmt.Printf("\n\t\tYou have to bet at least one money!\n\t\t")
		case !p.CanBet(uint(amount)):
			fmt.Printf("\n\t\tTo little money for that bet, you've got %d money\n\t\t", p)
		default:
			return uint(amount)
		}
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

func (p Player) Bet(amount uint) (Player, bool) {
	if left := p - Player(amount); left >= 0 {
		return left, true
	}
	return p, false
}

func (p Player) CanBet(amount uint) bool {
	if Player(amount) > p {
		color.Blue("\n\nYou have found an easter egg, a buffer overflow!\n\n")
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
