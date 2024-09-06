package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/fatih/color"
)

func main() {
	fmt.Println(`
Welcome to seven!
The rules are simple, you bet if the sum of two dice rolls is over/under or equal with seven.

Press Ctrl+C to exit.`)

	for {
		bet := getBets()
		dices := throwDices()
		fmt.Printf("\n\tDice result: %d\n", dices)

		res := checkResult(dices)
		if bet == res {
			color.Green("\tYou won!")
		} else {
			color.Red("\tYou lost!")
		}
	}

}

func throwDices() int {
	dice, err := rand.Int(rand.Reader, big.NewInt(6))
	dice2, err := rand.Int(rand.Reader, big.NewInt(6))
	if err != nil {
		panic("failed to generate random number: " + err.Error())
	}
	return int(2 + dice.Int64() + dice2.Int64())
}

func checkResult(sum int) Bet {
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

func getBets() Bet {
	c := color.New().Add(color.Underline)
	for {
		var picked string

		c.Printf("\nSeven, Under or Over? [(s)even/(u)nder/(o)ver]\n\t")
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
