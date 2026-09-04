package pokergame

import "fmt"

type PokerGameError struct {
	Message string
}

func (e *PokerGameError) Error() string {
	return e.Message
}

type InternalPokerGameError struct {
	Message string
}

func (e *InternalPokerGameError) Error() string {
	return e.Message
}

func newPokerGameError(format string, args ...any) error {
	return &PokerGameError{Message: fmt.Sprintf(format, args...)}
}

func newInternalPokerGameError(format string, args ...any) error {
	return &InternalPokerGameError{Message: fmt.Sprintf(format, args...)}
}
