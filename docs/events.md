# Event Implementation Guide

This is a checklist for implementing the types in `events.go`. Emission, retry behavior, and the complete game loop are planned work, not claims about what is already implemented.

Current code defines the event payloads and enums, including `LegalActions []ActionType`, `ToCall *int`, and `PlayerLeft`. `Street` and `EventType` have `String` methods; the newly added `PlayerLeft` case still needs to be added to `EventType.String()`. `Player.Left` exists but is not yet used by engine logic or included in `PlayerState`.

## Interface

The intended public flow is:

```go
game := New(options, actionSource, onEvent)
winner, err := game.Play(players)
```

- `ActionSource` supplies input through `ActionRequest`.
- `onEvent func(Event)` receives output. A nil callback disables delivery.
- The engine and callbacks are synchronous. Networking and asynchronous consumers live outside the engine.
- Logging, GUI updates, and replay verification consume events; the engine does not format logs or manage files.
- Hand execution and setup eventually become internal implementation details of `Play`.

The current constructor remains `New(options, actionSource)`. Callers use `SeatPlayers` and `PlayHand`; `PlayHand` and `InitActionSource` are temporary development APIs. `Play` will own setup, the hand loop, payouts, between-hand removal, and game completion.

## Build the Common Data Once

Use an internal `newEvent(kind EventType) Event` helper to capture the common fields: type, hand number, street, and copied player states. Fill event-specific fields at the call site. An `emit(event Event)` helper handles the nil callback check.

Illustrative usage:

```go
event := g.newEvent(ActionValid)
event.PlayerID = &playerID
event.Action = &action
g.emit(event)
```

Share player-snapshot construction with `actionRequest()` if useful, but keep event and input-request construction separate. Events do not require a current actor or legal-action list.

## Payload and Ownership

- `Players` is always supplied. It is a value snapshot of the internal player slice, in exactly the same order, including folded and all-in players.
- Every nil optional field means not applicable.
- `BlindIndices` indexes this event's `Players`, not a future roster.
- `Board` is a copied card slice. When applicable but empty, such as preflop, use a non-nil empty slice. Nil means no board context applies.
- `PlayerID` identifies the relevant player in `Players`; no duplicate player snapshot is needed.
- `LegalActions` is a copied action-type slice for `ActionRequested`, matching the request supplied to the action source. Nil means not applicable.
- `ToCall` is a snapshot-owned `*int` for `ActionRequested`. A pointer to zero means checking is available; nil means not applicable. `ActionRequest.ToCall` remains an `int`.
- `Action` is the submitted decision. For raises, its amount includes the call portion. Resulting player states show what was actually committed if the stack limited the amount.
- `Pots` contains independent `PotState` values. Copy nested `EligiblePlayerIDs` and `WinnerIDs` slices too. These are string IDs, not seat indices.
- `Err` carries rejection or failure information, without mutable engine references in custom error payloads.
- Optional pointers refer to snapshot-owned values, never live engine fields.
- Passing `Event` by value does not deep-copy slices or pointers. Capture independent data once; do not reuse its backing storage for later events.
- Consumers treat snapshots as read-only. Go does not enforce this; multiple consumers sharing a snapshot must respect the contract.
- Full snapshots include private hole cards. The networking layer must filter player-visible information rather than broadcast the internal snapshot unchanged.

`HandNumber` and `Street` are value fields, not optional pointers. Suggested convention: number hands from one, use zero before the first hand, and ignore `Street` on events without street context. If explicit absence becomes necessary, change that representation deliberately rather than interpreting the zero-valued `Preflop` as absence.

## Event Checklist

Every row includes the common snapshot fields. Additional payloads below are the event-specific information to populate. The timing below is a suggested consistent contract: successful transitions carry post-transition state; requests and rejections carry state before a successful action.

| Event | Timing | Additional payload |
| --- | --- | --- |
| `GameStarted` | Players are seated, before the first hand. | None required. Record options and seed separately in replay setup. |
| `HandStarted` | New hand state is initialized, before blinds and new hole cards. | Empty non-nil `Board`. Hole cards are not current until `HoleCardsDealt`. |
| `BlindsAdvanced` | Dealer and blind positions have been assigned. | `BlindIndices`. |
| `BlindsPosted` | Both blind commitments have been applied. | `BlindIndices`; `Players` shows actual committed amounts and remaining stacks. |
| `HoleCardsDealt` | All players have their new hole cards. | Cards are already in `Players`. |
| `StreetStarted` | The street's board and turn order are ready, before any action. | `Board`; `Street` identifies which street. |
| `ActionRequested` | Immediately before calling `NextAction`. | `PlayerID`, `Board`, copied `LegalActions`, and non-nil `ToCall`, matching `ActionRequest` for logging and replay validation. |
| `ActionValid` | The accepted action has been successfully applied, before closure events. | `PlayerID`, `Action`, `Board`; resulting player state is captured. |
| `ActionInvalid` | The submitted choice is rejected, before requesting another. | `PlayerID`, `Action`, `Err`, `Board`; game state is unchanged. |
| `StreetEnded` | Betting on this street is finished, before resetting acted flags. | `Board`. An explicit closure reason can be added later if needed. |
| `OnePlayerLeft` | Only one non-folded player remains, before payout. | `PlayerID`, `Board`. |
| `RunToShowdown` | No further betting is possible; emit once on that transition. | `Board`. |
| `ShowdownStarted` | The final board is ready and multiple players remain, before pot resolution. | `Board`. |
| `PotsCreated` | All pots and eligibility are known, before determining winners. | `Pots` with nil `WinnerIDs`, `Board`. |
| `WinnersDetermined` | Every pot has winners, before any payout. | `Pots` with populated `WinnerIDs`, `Board`. |
| `ChipsAwarded` | All payouts for this hand have completed. | `Pots` for showdown, or `PlayerID` for an uncontested win; `Board`. Player stacks show the result. |
| `HandEnded` | Payouts are complete, before clearing hand state or removing players. | `Board`, and `Pots` if applicable. |
| `PlayerBusted` | A player has zero chips after payout, before removal. | `PlayerID`; the player remains in this event's snapshot. |
| `PlayerLeft` | Suggested: a departure is accepted and marked for deferred removal; the trigger is still to be designed. | `PlayerID`; the player remains in this event's snapshot. Include `Action` and `Board` when applicable. |
| `GameEnded` | Between-hand cleanup has determined the game winner. | `PlayerID`; `Players` reflects the final roster. |
| `ErrorState` | A non-retryable failure occurs, before returning the error. | `Err`, plus player/action/board context where applicable. |

`ActionValid` means the action will be applied. Emitting it after successful application also gives consumers the resulting state without requiring a second action event. If it is instead emitted immediately after validation, document that its snapshot is pre-action.

Invalid choices should emit `ActionInvalid` and request another action without advancing the turn. A failure returned by the action source is a separate error path, not automatically an invalid choice to retry.

Currently, an action type outside `LegalActions` returns an `ErrGame` immediately; there is no retry or event emission. Raise amounts below `ToCall` are currently changed to `ToCall + 10` before committing chips, rather than rejected. Event work must distinguish this existing behavior from the planned rejection contract.

## Flow Details

- Emit `ActionValid` before leaving an action branch for an early win or all-in runout. Those branches currently break out of the loop, so placement matters.
- Emit one `StreetStarted` and `StreetEnded` for each actual street, including dealt runout streets with no action requests.
- Do not emit later street events after an uncontested win merely because `PlayHand` visits those code blocks.
- An uncontested win currently sums bets directly; it does not build pots. Its sequence can be `OnePlayerLeft`, `StreetEnded`, `ChipsAwarded`, `HandEnded`, without invented showdown or pot-construction events.
- At showdown, capture `PotsCreated` independently from `WinnersDetermined`. Adding winners later must not alter the earlier event.
- Keep final hand snapshots before resets so bets, cards, and final stacks remain inspectable.
- Emit `PlayerBusted` only after payout: an all-in player can win chips back.

## Player Order and Removal

The roster stays fixed during a hand. Snapshots preserve its internal order 1:1, so all blind indices refer to valid entries in that same snapshot. Folding does not remove or reorder anyone.

Planned disconnect behavior:

- Connection handling stays outside the engine; the action source supplies fallback actions.
- Folding will be allowed unconditionally by the engine. The client may choose not to offer it when checking is available.
- Departing players will be marked for removal. Busted players will use the same deferred-removal mechanism. `Folded` is hand-local; removal marks must persist until cleanup.
- Finish the hand and payouts before removing marked players. All-in players remain eligible for the hand's payout.
- Apply roster changes inside the engine's between-hand flow, not by concurrently mutating its players from network code.
- After removal, remap indices while preserving dealer/blind progression before emitting events containing new blind indices.
- A disconnected player and a busted player can share removal machinery without sharing the same reason. `PlayerBusted` describes a zero-stack outcome, not every departure.

`Player.Left` is present, but removal handling and unconditional-fold behavior remain planned. Currently `legalActions()` offers `Fold` only when `ToCall > 0`.

A `Leave` action has been considered: fold, mark the player for removal, and emit `PlayerLeft`, distinct from `PlayerBusted`. It is not in `ActionType` yet. Through `ActionSource`, it would only arrive at a player's turn; folded or all-in players will not receive another request. A queued `MarkForRemoval` method or a between-hand mechanism remains an option, not a settled interface. Removal alone must not cancel an all-in player's payout eligibility.

Additional betting-limit options remain deferred.

## Logging and Replay Consumers

The logging consumer chooses formatting and destinations: stdout, a per-run file, or both. It can include extra diagnostic presentation without changing event meanings.

The replay runner supplies recorded decisions through `ActionSource`, checks each request before responding, and compares emitted events against expected outcomes. Comparisons belong in the replay runner, not the engine. Record initial setup and randomness needed for deterministic replay separately from human-readable logs.

Start with the snapshot helper and a few events, then fill in this checklist. No callback registry, background worker, or generic event framework is required.
