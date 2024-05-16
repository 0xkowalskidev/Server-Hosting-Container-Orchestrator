package statemanager

type EventType string

const (
	NodeAdded        EventType = "NodeAdded"
	NodeRemoved      EventType = "NodeRemoved"
	NamespaceAdded   EventType = "NamespaceAdded"
	NamespaceRemoved EventType = "NamespaceRemoved"
	ContainerAdded   EventType = "ContainerAdded"
	ContainerRemoved EventType = "ContainerRemoved"
)

type Event struct {
	Type EventType
	Data interface{}
}

type Listener func(Event)

func (sm *StateManager) Subscribe(listener Listener) {
	sm.listeners = append(sm.listeners, listener)
}

// emit broadcasts an event to all subscribers.
func (sm *StateManager) emit(event Event) {
	for _, listener := range sm.listeners {
		listener(event)
	}
}
