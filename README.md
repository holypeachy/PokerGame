# PokerGame

A WIP Texas Hold'em game engine written in Go, with a simple and mechanical interface.

**Status:** Functional hand flow, active development, slow  
**Built with:** C# => Go | [PokerAlgo](https://github.com/holypeachy/PokerAlgo)

## Overview

PokerGame represents a game of Texas Hold'em. The main `PokerGame` type owns the players, deck, community cards, table positions, current bets, and the progression of each hand from the blinds through showdown.

The engine was built on top of my [PokerAlgo](https://github.com/holypeachy/PokerAlgo), which handles hand evaluation and winner determination. PokerGame is responsible for everything around it: rotating positions (aka who plays around the table and blinds), dealing cards, generating legal moves, processing player actions, building pots, and paying winners.

Input is kept outside the game logic through an `ActionSource`. The engine provides the current game state and valid moves, then receives a player action in return. The current sandbox uses console input, but the same boundary can later support AI players, a graphical interface, or network clients.

The current implementation is a Go port of the original C# project. The port preserves the existing behavior so I can continue development in Go, but the engine is still unfinished and will be reviewed as work continues.

## Features

- Runs a hand through pre-flop, flop, turn, river, and showdown.
- Rotates the dealer and blinds and tracks each player's stack, bet, and hand state.
- Generates and validates legal check, call, fold, and raise actions.
- Handles folded players, all-in players, early wins, raise limits, and automatic progression to showdown.
- Builds main and side pots from unequal contributions and determines the eligible winner or winners of each pot.
- Separates game logic from player input through game-state snapshots and an action interface.

## Getting Started

PokerGame requires Go 1.26 or later and the current Go version of PokerAlgo. The module currently expects PokerAlgo in a neighboring directory named `pokeralgo-go`.

```sh
mkdir poker-project
cd poker-project

git clone https://github.com/holypeachy/PokerAlgo.git pokeralgo-go
git clone https://github.com/holypeachy/PokerGame.git pokergame-go

cd pokergame-go
go build ./...
go run ./cmd/pokergame-sandbox
```

The sandbox creates five players and runs interactive hands through the console. It is primarily a development tool for inspecting state transitions and game behavior. It is not pretty at all.

## Usage

PokerGame receives player decisions through the `ActionSource` interface:

```go
type ActionSource interface {
	NextAction(gameState GameState) (PlayerAction, error)
}
```

Once an action source is available, a table can be initialized and a hand started:

```go
options := pokergame.PokerEngineOptions{
	BuyIn:            1_000,
	BigBlind:         50,
	AdditionalRaises: 1,
}

game := pokergame.NewPokerGame(options, actionSource)

players := []pokergame.PlayerInfoDto{
	pokergame.NewPlayerInfoDto("Alice"),
	pokergame.NewPlayerInfoDto("Bob"),
	pokergame.NewPlayerInfoDto("Charlie"),
}

if err := game.InitializeTable(players); err != nil {
	return err
}

if err := game.StartHand(); err != nil {
	return err
}
```

Each call to `NextAction` receives a `GameState` containing the current table state, the player who must act, the amount required to call, and the legal moves available to that player.

## Planned Work

- Review the port before I continue to implement features.
- Add structured logging, hand histories, and replay support.
- Complete the between-hand lifecycle, including busted-player removal and game-end reporting.
- Add unit and integration tests for betting, pot construction, and full-hand behavior.
