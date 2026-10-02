## Betting Round Logic

- Before a player acts
  - If only 1 non-folded player remains
    - True -> end hand immediately; that player wins the pot
    - False -> continue
  - If fewer than 2 players can act && bets are settled
    - (i.e., among non-folded players, fewer than 2 are not all-in)
    - True -> stop requesting input; run remaining board cards to 5; showdown
    - False -> continue
  - If player is folded
    - True -> skip to next player
    - False -> continue
  - If player is all-in
    - True -> skip to next player
    - False -> continue
  - Betting-round closure check (your "closing reference" logic)
    - Condition: "everyone has had a chance to respond since the last price change"
      - False -> round is not closed; continue to generate moves
      - True -> then check CurrentPlayerBet == CurrentBet
      - True -> betting round ends; advance street (or showdown if river)
      - False -> round is not closed; continue to generate moves
  - Generate legal moves
    - Compute ToCall = CurrentBet - PlayerBet
    - If ToCall > 0 -> allow Call / Fold (+ Raise if allowed)
    - If ToCall == 0 -> allow Check (+ Bet if allowed)
- Player acts
- After a player acts
  - If action is Fold
    - Apply fold
    - Then re-check: only 1 non-folded remains?
      - True -> end hand; remaining player wins pot
      - False -> continue
  - If action changes the price (Bet/Raise)
    - Update CurrentBet
  - If you enforce a raise cap:
    - If cap reached -> future players cannot raise again (only call/fold/check as applicable)
  - If after the action fewer than 2 players can act AND if bets are settled
    - True -> stop betting; run board to 5; showdown
    - False -> continue to next player

## After a betting round ends

- Advance street
  - Preflop -> Flop -> Turn -> River
  - River round end -> showdown
- Reset per-street flags

## After a hand ends

- Award pots
  - main pot + side pots
- Mark busted players
  - Stack == 0 -> busted

## After a hand (between hands)

- Remove busted players from active table list
  - busted -> removed from _players (but can remain in your "all players" list/history)
- If fewer than 2 non-busted players remain
  - True -> game ends
  - False -> next hand
