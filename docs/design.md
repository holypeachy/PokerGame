# PokerGame Technical Design Notes

_Prepared with Codex._

This document is a recovery note for the author of the current Go `PokerGame` repo. It uses the codebase as the source of truth, with the attached Code Tracker treated only as design context. The tracker was not modified.

## Project Shape

PokerGame is a work-in-progress Texas Hold'em game engine. It owns the flow around a poker hand: table setup, blinds, betting rounds, streets, legal moves, pot construction, and payouts. It delegates hand evaluation and showdown ranking to the sibling `PokerAlgo` module through the local `replace pokeralgo => ../PokerAlgo` directive in `go.mod`.

The root package contains the engine:

- `game.go`: main game state, hand lifecycle, betting loop, showdown, game-state snapshots, position rotation.
- `player.go`: mutable player state and actions that change stack, bet, folded/acted flags, and hole cards.
- `pot_algo.go`, `pot.go`: main/side-pot construction and payout splitting.
- `types.go`: public DTOs, options, game snapshots, and `ActionSource`.
- `player_move.go`, `output_type.go`: small enums for legal moves and output/request type.
- `errors.go`: exported `ErrGame` and `ErrInternal` sentinel errors.
- `events.go`, `debug.go`: early placeholders for event and debug infrastructure.

The only command is `cmd/sandbox`, an interactive console runner used for manual development.

## Relationship to PokerAlgo

PokerGame does not evaluate card strength itself. At showdown it converts eligible engine `Player` values into `pokeralgo.Player` values, calls `pokeralgo.DetermineWinners`, and maps the resulting winners back to engine players by ID.

That makes the boundary simple:

- PokerGame knows about stacks, bets, folds, blinds, turn order, pots, and input.
- PokerAlgo knows which poker hand wins once player cards and community cards are known.

`toAlgoPlayers` places each engine `Player.ID` in `pokeralgo.Player.Name`. `MapAlgoPlayers` matches those names back to engine IDs, which must be unique.

## Public Control Boundary

The game asks for player decisions through:

```go
type ActionSource interface {
	NextAction(gameState GameState) (Action, error)
}
```

`Game.GameState()` builds a `GameState` containing:

- `PlayerStates`: snapshots with `ID`, `Stack`, `Bet`, `HasFolded`, and hole cards,
- `CommunityCards`: a copy of the game's `Board`,
- `PlayerToAct`: a pointer into the snapshot's `PlayerStates` slice,
- `LegalActions`: available `ActionType` values,
- `ToCall`: the amount needed to match `CurrentBet`,
- `OutputType`: currently `InputRequest`.

The sandbox implements `ActionSource` through `ConsoleActionSource.NextAction`. Returned `Action` values contain `Type` and `Amount`. The sandbox's `SetEngine` reference supplies development output; decisions enter through `NextAction`.

`New(options Options, actionSource ActionSource)` constructs a game. `InitActionSource` can assign the source later if it is nil; it rejects reassignment. `SeatPlayers` accepts `[]PlayerSpec`, whose `ID` is also used as the engine player's ID. The constructor helper is still named `NewPlayerInfo`.

The input boundary can support later AI, UI, or network clients. Calls are synchronous: `PlayHand` runs the whole hand and waits for each `NextAction` to return.

## Hand Lifecycle

`SeatPlayers` requires an action source and at least two players. It resets the deck, sets the dealer to `-1`, creates each `Player` with the configured buy-in, and deals initial placeholder hole cards.

`PlayHand` runs one full hand:

1. Reset deck and community cards.
2. Advance dealer, small blind, big blind, and first-to-act indices.
3. Post blinds.
4. Deal each active player a new two-card hand.
5. Run preflop betting.
6. Burn and deal flop, then run betting.
7. Burn and deal turn, then run betting.
8. Burn and deal river, then run betting.
9. If more than one player remains, run showdown.
10. Reset hand flags/bets for the next hand.

If all but one player folds during betting, `onePlayerLeft` is set and the remaining player receives the committed bets. Later betting calls return immediately and community-card dealing and showdown are skipped. `PlayHand` still executes its later street print statements and burn-card draws.

If fewer than two players can act and everyone is settled at the current bet, the engine sets `runToShowdown`; later betting rounds return immediately while the board runs out to five cards.

## Betting Round Logic

`runBettingRound` is the core engine loop. For the current player, it first checks whether the player should be skipped or whether the round/hand should close:

- folded players are skipped,
- all-in players are skipped,
- one non-folded player ends the hand,
- fewer than two actors plus settled bets skips to showdown,
- everyone acted plus settled bets ends the betting round.

If action is needed, the engine:

1. Builds `GameState`.
2. Requests an action from `ActionSource`.
3. Verifies the move is currently legal.
4. Applies fold, check, call, or raise.
5. Updates `CurrentBet` when a bet/raise increases the price.
6. Rechecks all-in/settled conditions.
7. Advances to the next player index.

Legal moves are derived from `CurrentBet - player.Bet`:

- if `ToCall > 0`: call or fold are legal,
- if `ToCall == 0`: check is legal,
- raise is added when the raise cap is not reached and more than one player can still act.

The raise cap is represented by `AdditionalRaiseCount` and `Options.AdditionalRaises`.

Known quirk: when a raise amount is below `toCall`, the code currently forces `toCall + 10` instead of returning an invalid-action error. That is current behavior, not a recommendation.

## Errors

`ErrGame` identifies game input/setup failures; `ErrInternal` identifies engine invariant failures. These are package-level sentinel errors created with `errors.New`. Context is added with `fmt.Errorf("%w: ...", sentinel)` and callers classify errors with `errors.Is`:

```go
if err := game.PlayHand(); err != nil {
	if errors.Is(err, pokergame.ErrGame) {
		// Handle a game input/setup failure.
	}
}
```

This replaces the former `GameError` and `InternalError` structs and their constructor helpers. Messages include the sentinel prefix. Existing error conditions and panic-versus-return behavior are preserved; the busted-player panic now carries an error wrapping `ErrInternal`. Other string panics remain strings.

Errors received from PokerAlgo, the action source, and console parsing/IO are propagated unchanged, so they need not match either PokerGame sentinel. For example, PokerAlgo failures can be checked with `errors.Is(err, pokeralgo.ErrDuplicateCards)`.

## Player State Invariants

`Player` tracks:

- `Stack`: chips not currently committed,
- `Bet`: chips committed in the current hand,
- `Acted`: whether the player has acted in the current betting round,
- `Folded`: whether the player is out of the current hand,
- `HoleCards`: current private cards.

Facts from code:

- Bets cannot be negative.
- Paying a player cannot use a negative amount.
- Betting more than the stack puts the player all-in by moving the remaining stack into `Bet`.
- A player with `Stack == 0` and `Bet > 0` is considered all-in.
- A player with `Stack == 0` and `Bet == 0` causes an internal panic because busted players should not still be active.

`credit` adds chips to the stack. `postBlind` commits chips without marking the player as acted; `commitChips` also sets `Acted`. `deal` assigns hole cards. `resetForBettingRound` clears `Acted`, while `resetForNextHand` clears `Bet`, `Acted`, and `Folded`. The snapshot field remains named `PlayerState.HasFolded`.

The between-hand removal of busted players is not implemented yet.

## Pot Construction and Showdown

`buildPots` creates side pots from players' committed bets. It rejects states where every player folded, where nobody bet, or where a player has a negative bet.

The pot algorithm:

1. Convert each nonzero player bet into a `contribution` with `Player`, `Remaining`, and `HasFolded`.
2. `splitPots` uses `getMinBet` to find the smallest remaining non-folded contribution.
3. Pull that amount from each non-folded contribution and up to that amount from each folded contribution.
4. Folded players contribute chips but are not eligible for the pot.
5. Non-folded players contribute and become eligible for that pot.
6. Remove empty trackers and recurse until no chips remain.

Each `Pot` holds `EligiblePlayers`, `Amount`, and `Winners`; `Distribute` credits its winners.

At showdown:

- single-player pots are assigned to that player,
- multi-player pots call `pokeralgo.DetermineWinners` using only the players eligible for that pot,
- each pot pays its winners equally with integer division.

Known limitation: leftover chips from uneven splits are not explicitly handled.

## Tests and Verification

Current fact: there are no checked-in unit or integration tests.

`go test ./...` passes locally, but only verifies package compilation:

- `pokergame`: no test files,
- `cmd/sandbox`: no test files.

The Code Tracker and sandbox TODOs both indicate that tests are still planned. The highest-value future tests would cover betting-round closure, all-in skip-to-showdown behavior, legal move generation, side-pot construction, showdown payout, and full-hand flows.

## Known Limitations and Unfinished Work

Facts from code, README, tracker, and sandbox TODOs:

- The project is a WIP Go port from the C# implementation.
- Structured logging, hand history, and replay are not implemented.
- `Event` is empty and `EventSink` declares `OnEvent(event Event)`; neither is wired into the game yet.
- `Options.EnableDebug` and `Options.DebugVerbosity` (`DebugLevel`) are not read by the engine. Pot output is separately controlled by the internal `potDebugEnabled` variable, currently true.
- End-hand/end-game reporting is incomplete.
- Busted-player removal between hands is not implemented.
- There are no unit or integration tests yet.
- Flexible rulesets are planned but not implemented.
- Public `GameState` currently includes hole cards for every player, which is useful for debugging but may not be appropriate for real hidden-information clients.
- Output/debug printing is still mixed directly into engine methods.

Reasonable inferences:

- The engine is currently optimized for getting hand flow correct and inspectable, not for a stable public API.
- The attached Code Tracker is effectively the design checklist for `runBettingRound`; the current code closely follows it.
- The cleanest architectural seam is already present: `ActionSource` separates decision-making from game mechanics.

## Future Plans

Planned engine work from the Code Tracker, sandbox TODOs, and current project direction:

- Add deep engine logging. The useful logs are not just console messages; they should record meaningful state transitions: hand start, deck seed, blind movement, cards dealt, betting-round start/end, action requests, accepted moves, rejected moves, all-in transitions, street advancement, pot creation, showdown winners, payouts, busted players, and game end.
- Add unit and integration tests. The first test layer should pin small pure behaviors like legal move generation, `Player` betting/all-in behavior, and pot splitting. The second layer should run full scripted hands through a fake `ActionSource`.
- Add a replay system. A replay file should store enough information to deterministically reconstruct a hand or game: engine version, rules/options, player starting information, deck seed, blinds/positions, player moves, and important engine outputs.
- Add a replay `ActionSource`. Instead of asking a human for input, it reads the next recorded move from a replay file.
- Validate replay requests. When the replay `ActionSource` receives a `GameState`, it should check that the current request matches the replay file before returning the recorded move. At minimum, validate player-to-act, street/community state, `ToCall`, legal moves, and the recorded action.
- Use replays as tests. Once replay files exist, they can become regression fixtures: load a replay, run the engine, and verify the same state transitions/payouts occur.
- Complete the between-hand lifecycle. After a hand ends, mark players with `Stack == 0` as busted, remove them from the active table list, and end the game when fewer than two non-busted players remain.
- Add structured hand/game end reporting. This likely belongs in the same event/logging/replay model rather than as direct `fmt.Println` calls.
- Keep rules flexible enough to change later. Current obvious candidates are buy-in, blinds, raise cap, betting rules, and possibly payout/remainder behavior.

Reasonable replay shape:

```text
ReplayFile
  engineVersion
  options/rules
  initialPlayers
  hands[]
    deckSeed
    dealer/sb/bb positions
    startingStacks
    moves[]
      playerId
      expectedToCall
      expectedLegalMoves
      move
      amount
    expectedResult
      finalStacks
      pots
      winners
```

This would turn the current `ActionSource` boundary into the key replay seam: live games use a human/AI/network source, replay games use a file-backed source, and tests can use the same mechanism.
