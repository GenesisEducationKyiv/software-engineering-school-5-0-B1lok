package event

type Name string

type Event interface {
	EventName() Name
}
