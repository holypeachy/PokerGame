package pokergame

type EngineEvent struct {
}

type EventSink interface {
	OnEvent(event EngineEvent)
}
