package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
)

func main() {
	fmt.Println("Welcome to seven, the rules are simple. You bet if the sum of two dice rolls is over/under or equal with seven. Press ctrl+C")
	for {
		bet := getBets()
		dices := throwDices()
		fmt.Printf("Dice result: %d\n", dices)

		res := checkResult(dices)
		if bet == res {
			fmt.Println("You won!")
		} else {
			fmt.Println("You lost!")
		}
	}

}

func throwDices() int {
	dice := rand.Intn(5) + 1
	dice2 := rand.Intn(5) + 1
	return dice + dice2
}

func checkResult(sum int) Bet {
	if sum < 7 {
		return Under
	} else if sum == 7 {
		return Seven
	} else {
		return Over
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
	for {
		var picked string

		fmt.Println("Seven, Under or Over? (s/u/o)")
		_, err := fmt.Scanln(&picked)
		if err != nil && err.Error() != "unexpected newline" {

			log.Fatal("program failed: failed to get user input:", err)
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
