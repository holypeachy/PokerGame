package eventlog

import (
	"fmt"
	"io"
	"log"
	"os"
	"pokeralgo"
	"strings"
	"text/tabwriter"

	"pokergame"
)

type EventLogger struct {
	logger *log.Logger
}

func New(w io.Writer) *EventLogger {
	l := log.New(w, "", 0)

	return &EventLogger{
		logger: l,
	}
}

func (l *EventLogger) OnEvent(e pokergame.Event) {
	message := l.formatEvent(&e)
	if err := l.logger.Output(2, message); err != nil {
		fmt.Fprintf(os.Stderr, "event log failed: %v\n%s", err, message)
	}
}

func (l *EventLogger) formatEvent(e *pokergame.Event) string {
	var b strings.Builder
	formatHead(&b, e)
	formatPlayers(&b, e.Players)

	switch e.Type {
	case pokergame.HandStarted, pokergame.StreetStarted, pokergame.StreetEnded, pokergame.HandEnded:
		formatBoard(&b, e.Board)
	case pokergame.BlindsAdvanced, pokergame.BlindsPosted:
		fmt.Fprintf(&b, "    D:%d SB:%d BB:%d\n", e.BlindIndices.Dealer, e.BlindIndices.SmallBlind, e.BlindIndices.BigBlind)
	case pokergame.ActionRequested:
		fmt.Fprintf(&b, "    Player:%s Legal:%v ToCall:%d\n", *e.PlayerID, e.LegalActions, *e.Amount)
		formatBoard(&b, e.Board)
	case pokergame.ActionValid:
		fmt.Fprintf(&b, "    Player:%s Action:%s Amount:%d\n", *e.PlayerID, *e.ActionType, *e.Amount)
		formatBoard(&b, e.Board)
	case pokergame.ActionInvalid:
		fmt.Fprintf(&b, "    Player:%s Action:%s Legal:%v ToCall:%d\n", *e.PlayerID, *e.ActionType, e.LegalActions, *e.Amount)
		formatBoard(&b, e.Board)
	case pokergame.OnePlayerLeft:
		fmt.Fprintf(&b, "    Winner:%s\n", *e.PlayerID)
		formatBoard(&b, e.Board)
		for _, p := range *e.Pots {
			fmt.Fprintf(&b, "    %s\n", formatPot(&p))
		}
	case pokergame.PotsCreated:
		formatBoard(&b, e.Board)
		for _, p := range *e.Pots {
			fmt.Fprintf(&b, "    Pot Eligible:[%s] Amount:%d\n", strings.Join(p.EligiblePlayerIDs, ", "), p.Amount)
		}
	case pokergame.WinnersDetermined, pokergame.ChipsAwarded:
		formatBoard(&b, e.Board)
		for _, p := range *e.Pots {
			fmt.Fprintf(&b, "    %s\n", formatPot(&p))
		}
	case pokergame.PlayerBusted, pokergame.PlayerLeft:
		fmt.Fprintf(&b, "    Player:%s\n", *e.PlayerID)
	case pokergame.GameEnded:
		fmt.Fprintf(&b, "    Winner:%s\n", *e.PlayerID)
	case pokergame.ErrorState:
		fmt.Fprintf(&b, "    Error:%v\n", e.Err)
	}

	b.WriteByte('\n')
	return b.String()
}

func formatHead(b *strings.Builder, e *pokergame.Event) {
	fmt.Fprintf(b, "%d-%s: ", e.HandNumber, e.Street)

	switch e.Type {
	case pokergame.GameStarted:
		fmt.Fprintf(b, "%s 🎲\n", e.Type)
	case pokergame.BlindsPosted:
		fmt.Fprintf(b, "%s 🪙\n", e.Type)
	case pokergame.StreetStarted:
		fmt.Fprintf(b, "%s 🃏\n", e.Type)
	case pokergame.ActionRequested:
		fmt.Fprintf(b, "%s 💭\n", e.Type)
	case pokergame.ActionValid:
		fmt.Fprintf(b, "%s ✅\n", e.Type)
	case pokergame.ActionInvalid:
		fmt.Fprintf(b, "%s ❌\n", e.Type)
	case pokergame.ChipsAwarded:
		fmt.Fprintf(b, "%s 💰\n", e.Type)
	case pokergame.PlayerBusted:
		fmt.Fprintf(b, "%s 💸\n", e.Type)
	case pokergame.PlayerLeft:
		fmt.Fprintf(b, "%s 🚪\n", e.Type)
	case pokergame.GameEnded:
		fmt.Fprintf(b, "%s 🏁\n", e.Type)
	case pokergame.ErrorState:
		fmt.Fprintf(b, "%s ⚠️\n", e.Type)
	default:
		fmt.Fprintf(b, "%s\n", e.Type)
	}
}

func formatPlayers(b *strings.Builder, players []pokergame.PlayerState) {
	w := tabwriter.NewWriter(b, 0, 4, 2, ' ', 0)
	for _, p := range players {
		fmt.Fprintf(w, "    %s\tFolded:%t\tLeft:%t\tStack:%d\t%s%s\tBet:%d\n", p.ID, p.Folded, p.Left, p.Stack, formatCard(&p.HoleCards.First), formatCard(&p.HoleCards.Second), p.Bet)
	}
	w.Flush()
}

func formatCard(c *pokeralgo.Card) string {
	return fmt.Sprintf("[%s%s]", pokeralgo.RankToSymbol[c.Rank], pokeralgo.SuitToSymbol[c.Suit])
}

func formatBoard(b *strings.Builder, cards []pokeralgo.Card) {
	b.WriteString("    Board:")
	if len(cards) == 0 {
		b.WriteString("[]")
	}
	for _, c := range cards {
		b.WriteString(formatCard(&c))
	}
	b.WriteByte('\n')
}

func formatPot(p *pokergame.PotState) string {
	return fmt.Sprintf("Pot Eligible:[%s] Amount:%d Winners:[%s]", strings.Join(p.EligiblePlayerIDs, ", "), p.Amount, strings.Join(p.WinnerIDs, ", "))
}
