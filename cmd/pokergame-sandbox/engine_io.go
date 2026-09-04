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
	engine *pokergame.PokerGame
	reader *bufio.Reader
}

func NewConsoleActionSource() *ConsoleActionSource {
	return &ConsoleActionSource{
		reader: bufio.NewReader(os.Stdin),
	}
}

func (s *ConsoleActionSource) NextAction(gameState pokergame.GameState) (pokergame.PlayerAction, error) {
	if s.engine == nil {
		return pokergame.PlayerAction{}, fmt.Errorf("ConsoleActionSource: when engine is used it should already be assigned")
	}

	moves := map[int]pokergame.PlayerMove{}
	// _engine.PrintGameState();
	fmt.Println("IO Request")
	fmt.Printf("Output Type: %s\n", gameState.OutputType)
	currentPlayer := "null"
	if gameState.PlayerToAct != nil {
		currentPlayer = gameState.PlayerToAct.Id
	}
	fmt.Println("Current Player: " + currentPlayer)
	fmt.Printf("AdditionalRaiseCount: %d\n", s.engine.AdditionalRaiseCount)
	fmt.Print("Possible Moves: ")
	count := 1
	if gameState.PossibleMoves != nil {
		for _, item := range gameState.PossibleMoves {
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
		return pokergame.PlayerAction{}, err
	}
	moveIn = strings.TrimSpace(moveIn)
	if moveIn == "" {
		return pokergame.PlayerAction{}, fmt.Errorf("empty input")
	}
	moveNumber, err := strconv.Atoi(moveIn)
	if err != nil {
		return pokergame.PlayerAction{}, err
	}
	selectedMove, ok := moves[moveNumber]
	if !ok {
		return pokergame.PlayerAction{}, fmt.Errorf("invalid move selection")
	}

	amount := gameState.ToCall
	if selectedMove == pokergame.Raise {
		fmt.Println("ToCall + What Amount:")
		amountIn, err := s.reader.ReadString('\n')
		if err != nil {
			return pokergame.PlayerAction{}, err
		}
		amountIn = strings.TrimSpace(amountIn)
		amount, err = strconv.Atoi(amountIn)
		if err != nil {
			return pokergame.PlayerAction{}, err
		}
	}

	return pokergame.PlayerAction{Move: selectedMove, Amount: gameState.ToCall + amount}, nil
}

func (s *ConsoleActionSource) SetEngine(engine *pokergame.PokerGame) {
	s.engine = engine
}
