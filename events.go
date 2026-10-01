package pokergame

type Event struct {
}

type EventSink interface {
	OnEvent(event Event)
}
