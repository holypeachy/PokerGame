## Build the Common Data Once
- type
- hand number
- street
- player states.

## Event Checklist

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
