package main

import (
	"fmt"
	"os"

	pokergame "pokergame"
)

func main() {
	engineOptions := pokergame.Options{BuyIn: 1000, BigBlind: 50, AdditionalRaises: 1, EnableDebug: true}
	actionSource := NewConsoleActionSource()
	engine := pokergame.New(engineOptions, actionSource)
	actionSource.SetEngine(engine)

	playersInfo := []pokergame.PlayerSpec{
		pokergame.NewPlayerInfo("Alpha"),
		pokergame.NewPlayerInfo("Tango"),
		pokergame.NewPlayerInfo("Sierra"),
		pokergame.NewPlayerInfo("Quebec"),
		pokergame.NewPlayerInfo("Zulu"),
	}

	if err := engine.SeatPlayers(playersInfo); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := engine.PlayHand(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := engine.PlayHand(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

/*
! ISSUES:
!

TODO
TODO: Add detailed and standardized logging, log to file as well. Deep engine logging.
TODO: Add end hand and end game logic and reporting
TODO: Implement replay functionality
TODO: Unit and integration tests
TODO: Account for flexible ruleset so I can make changes later

? Future Ideas
?

* Notes
*

* Changes
* have codex generate architectural docs for future reference
* renamed a bunch of symbols
* redid errors
*/
