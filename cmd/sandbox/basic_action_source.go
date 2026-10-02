package main

import (
	"bufio"
	"fmt"
	"os"
	"pokergame"
	"strconv"
	"strings"
)

type BasicActionSource struct {
	reader *bufio.Reader
}

func NewBasicActionSource() *BasicActionSource {
	return &BasicActionSource{
		reader: bufio.NewReader(os.Stdin),
	}
}

func (s *BasicActionSource) NextAction(actionRequest pokergame.ActionRequest) (pokergame.Action, error) {
	moves := map[int]pokergame.ActionType{}
	currentPlayer := "null"
	if actionRequest.PlayerToAct != nil {
		currentPlayer = actionRequest.PlayerToAct.ID
	}
	fmt.Println("Current Player: " + currentPlayer)
	fmt.Print("Possible Moves: ")
	count := 1
	if actionRequest.LegalActions != nil {
		for _, item := range actionRequest.LegalActions {
			fmt.Printf("%d-%s ", count, item)
			moves[count] = item
			count++
		}
		fmt.Println()
	} else {
		fmt.Println("null")
	}
	fmt.Printf("ToCall: %d\n", actionRequest.ToCall)
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

	amount := 0
	if selectedMove == pokergame.Raise {
		fmt.Println("Raise Amount:")
		amountIn, err := s.reader.ReadString('\n')
		if err != nil {
			return pokergame.Action{}, err
		}
		amountIn = strings.TrimSpace(amountIn)
		amount, err = strconv.Atoi(amountIn)
		if err != nil {
			return pokergame.Action{}, err
		}
		if amount >= actionRequest.ToCall {
			amount = amount - actionRequest.ToCall
		}
	}

	return pokergame.Action{Type: selectedMove, Amount: amount + actionRequest.ToCall}, nil
}
