package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	pokergame "pokergame"
)

type ConsoleActionSource struct {
	engine *pokergame.Game
	reader *bufio.Reader
}

func NewConsoleActionSource() *ConsoleActionSource {
	return &ConsoleActionSource{
		reader: bufio.NewReader(os.Stdin),
	}
}

func (s *ConsoleActionSource) NextAction(gameState pokergame.GameState) (pokergame.Action, error) {
	if s.engine == nil {
		return pokergame.Action{}, fmt.Errorf("%w: ConsoleActionSource: when engine is used it should already be assigned", pokergame.ErrInternal)
	}

	moves := map[int]pokergame.ActionType{}
	// _engine.PrintGameState();
	fmt.Println("IO Request")
	fmt.Printf("Output Type: %s\n", gameState.OutputType)
	currentPlayer := "null"
	if gameState.PlayerToAct != nil {
		currentPlayer = gameState.PlayerToAct.ID
	}
	fmt.Println("Current Player: " + currentPlayer)
	fmt.Printf("AdditionalRaiseCount: %d\n", s.engine.AdditionalRaiseCount)
	fmt.Print("Possible Moves: ")
	count := 1
	if gameState.LegalActions != nil {
		for _, item := range gameState.LegalActions {
			fmt.Printf("%d-%s ", count, item)
			moves[count] = item
			count++
		}
		fmt.Println()
	} else {
		fmt.Println("null")
	}
	fmt.Printf("ToCall: %d\n", gameState.ToCall)
	fmt.Println("Select your move:")
	moveIn, err := s.reader.ReadString('\n')
	if err != nil {
		return pokergame.Action{}, err
	}
	moveIn = strings.TrimSpace(moveIn)
	if moveIn == "" {
		return pokergame.Action{}, fmt.Errorf("%w: empty input", pokergame.ErrGame)
	}
	moveNumber, err := strconv.Atoi(moveIn)
	if err != nil {
		return pokergame.Action{}, err
	}
	selectedMove, ok := moves[moveNumber]
	if !ok {
		return pokergame.Action{}, fmt.Errorf("%w: invalid move selection", pokergame.ErrGame)
	}

	amount := gameState.ToCall
	if selectedMove == pokergame.Raise {
		fmt.Println("ToCall + What Amount:")
		amountIn, err := s.reader.ReadString('\n')
		if err != nil {
			return pokergame.Action{}, err
		}
		amountIn = strings.TrimSpace(amountIn)
		amount, err = strconv.Atoi(amountIn)
		if err != nil {
			return pokergame.Action{}, err
		}
	}

	return pokergame.Action{Type: selectedMove, Amount: gameState.ToCall + amount}, nil
}

func (s *ConsoleActionSource) SetEngine(engine *pokergame.Game) {
	s.engine = engine
}
