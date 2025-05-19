package cotlib

import (
	"bytes"
	"sync"
)

var eventPool = sync.Pool{
	New: func() any { return new(Event) },
}

func getEvent() *Event { return eventPool.Get().(*Event) }

func ReleaseEvent(e *Event) {
	if e == nil {
		return
	}
	*e = Event{}
	eventPool.Put(e)
}

var bufPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

func getBuffer() *bytes.Buffer {
	b := bufPool.Get().(*bytes.Buffer)
	b.Reset()
	return b
}

func putBuffer(b *bytes.Buffer) {
	if b == nil {
		return
	}
	b.Reset()
	bufPool.Put(b)
}
