package event_bus

import (
	"fmt"
	"sync"
)

type Event struct {
	EventName string
	Data      any
}

type innerEvent struct {
	ch      chan Event
	handler func(Event)
}

func (s *innerEvent) publish(event Event) {
	s.ch <- event
}

func (s *innerEvent) close() {
	close(s.ch)
}

func (s *innerEvent) apply() {
	go func() {
		for {
			select {
			case event, ok := <-s.ch:
				if !ok {
					return
				}
				s.handler(event)
			}
		}
	}()
}

type EventBus struct {
	subscribers map[string][]innerEvent
	chanSize    int
	sync.RWMutex
}

func NewEventBus(size int) *EventBus {
	return &EventBus{
		subscribers: make(map[string][]innerEvent),
		chanSize:    size,
	}
}

func (eb *EventBus) Subscribe(eventName string, handler func(Event)) {
	ch := make(chan Event, eb.chanSize)
	sub := innerEvent{ch: ch, handler: handler}
	eb.Lock()
	eb.subscribers[eventName] = append(eb.subscribers[eventName], sub)
	eb.Unlock()
	sub.apply()
}

func (eb *EventBus) Publish(eventName string, data any) error {
	eb.RLock()
	defer eb.RUnlock()
	events, ok := eb.subscribers[eventName]
	if !ok {
		return fmt.Errorf("event %s not found", eventName)
	}
	for _, e := range events {
		e.publish(Event{EventName: eventName, Data: data})
	}
	return nil
}

func (eb *EventBus) PublishSync(eventName string, data any) error {
	eb.RLock()
	defer eb.RUnlock()
	events, ok := eb.subscribers[eventName]
	if !ok {
		return fmt.Errorf("event %s not found", eventName)
	}
	for _, e := range events {
		e.handler(Event{EventName: eventName, Data: data})
	}
	return nil
}

func (eb *EventBus) CloseEvent(eventName string) {
	eb.Lock()
	defer eb.Unlock()
	for _, sub := range eb.subscribers[eventName] {
		sub.close()
	}
	delete(eb.subscribers, eventName)
}

func (eb *EventBus) Close() {
	eb.Lock()
	defer eb.Unlock()
	for _, subs := range eb.subscribers {
		for _, sub := range subs {
			sub.close()
		}
	}
	eb.subscribers = make(map[string][]innerEvent)
}
