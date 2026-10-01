# PokerGame Architecture Flow

_Prepared with Codex._

This is the small mental model for the engine.

PokerGame does not decide which poker hand is strongest. PokerAlgo already does that.

PokerGame decides everything around the hand:

```text
who is playing
who acts next
what moves are legal
how much each player has bet
when a betting round is over
what pots exist
who gets paid
```

## Big Picture

The project is basically three layers:

```text
Player
  owns stack, bet, cards, folded/acted state

Game
  owns the hand flow, betting loop, board, blinds, and turn order

PokerAlgo
  receives final cards and says who wins
```

The sandbox is just a console input layer sitting outside the engine.

## Main Hand Flow

A hand starts with `PlayHand`.

```text
reset deck
clear community cards
move dealer/blinds
post small blind and big blind
deal two cards to each player
run preflop betting
deal flop
run betting
deal turn
run betting
deal river
run betting
showdown if needed
reset hand state
```

At every street, the important question is:

```text
Do we need more player actions,
or is this betting round done?
```

## Betting Loop

The betting loop looks at the current player and decides what to do.

Before asking for input:

```text
if player folded:
  skip them

if player is all-in:
  skip them

if only one player has not folded:
  end the hand and pay that player

if fewer than two players can act and bets are settled:
  stop betting and run to showdown

if everyone has acted and bets are settled:
  end this betting round
```

If none of those stops happen, the engine asks the `ActionSource` for a move.

## Legal Moves

Legal moves come from the price to call:

```text
toCall = CurrentBet - player.Bet
```

If `toCall > 0`, the player can:

```text
Call
Fold
Raise, if raises are still allowed
```

If `toCall == 0`, the player can:

```text
Check
Raise, if raises are still allowed
```

`GameState()` returns these choices in `GameState.LegalActions`. The action source returns an `Action` with `Type` (`Fold`, `Check`, `Call`, or `Raise`) and `Amount`; the engine rejects types outside that list.

## Why ActionSource Exists

`ActionSource` is the input boundary.

The engine says:

```text
here is the current state
here are the moves this player can make
give me one action
```

The console sandbox answers through `ConsoleActionSource.NextAction`, which reads stdin. Its `SetEngine` reference is currently used for development output.

Later, the same interface can be used by AI, a UI, or multiplayer networking without changing the betting rules.

## Showdown Flow

If more than one player survives to showdown:

```text
build pots from player bets
for each pot:
  find players eligible for that pot
  ask PokerAlgo who wins among those players
  map PokerAlgo winners back to engine Players
  split and pay the pot
```

PokerGame cares about pot eligibility.

PokerAlgo cares about card strength.

That separation is the main design idea.

## Pot Splitting

The pot algorithm repeatedly peels off the smallest remaining non-folded bet.

Example idea:

```text
each recursion level creates one pot
smallest active bet decides the pot layer
folded players add dead chips but cannot win
non-folded players in that layer can win that pot
remaining chips become the next side pot
```

So unequal all-in bets naturally become main pots and side pots.

## Future Shape

The next version of the engine should make the hand flow observable and replayable. `Event` is currently empty and `EventSink.OnEvent(Event)` is only an interface declaration; the game does not emit events yet.

The planned pieces are:

```text
deep engine logs
tests
replay files
replay ActionSource
between-hand cleanup
game-end reporting
```

The replay idea fits the current architecture nicely because input already goes through `ActionSource`.

Live game:

```text
PokerGame asks for action
Console/AI/UI ActionSource returns action
PokerGame validates and applies it
```

Replay game:

```text
PokerGame asks for action
Replay ActionSource reads the next saved move
Replay ActionSource checks the request matches the file
PokerGame validates and applies it
```

A replay file should store the stuff needed to recreate the hand:

```text
engine version
game options
starting players and stacks
deck seed
dealer/blind positions
player moves
important results
```

Deep logs and replay data should tell the same story at different levels. Logs explain what the engine did. Replays give the engine the same inputs again.

Tests can then use small fake action sources for focused cases and replay files for bigger regression tests.

## Short Version

PokerGame is the table manager.

```text
Player stores mutable player state.
Game runs the hand and betting loop.
ActionSource provides decisions.
Pot logic turns committed bets into payable pots.
PokerAlgo decides card winners at showdown.
```

The most important idea is that betting and payouts live here, but hand strength lives in PokerAlgo.
