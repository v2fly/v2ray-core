package cache

import "sync"

// Lru simple, fast lru cache implementation
type Lru interface {
	Get(key interface{}) (value interface{}, ok bool)
	GetKeyFromValue(value interface{}) (key interface{}, ok bool)
	Put(key, value interface{})
}

type lru struct {
	capacity       int
	count          int
	head           *lruElement
	tail           *lruElement
	keyToElement   *sync.Map
	valueToElement *sync.Map
	mu             *sync.Mutex
}

type lruElement struct {
	key   interface{}
	value interface{}
	prev  *lruElement
	next  *lruElement
}

// NewLru init a lru cache
func NewLru(cap int) Lru {
	return &lru{
		capacity:       cap,
		keyToElement:   new(sync.Map),
		valueToElement: new(sync.Map),
		mu:             new(sync.Mutex),
	}
}

func (l *lru) Get(key interface{}) (value interface{}, ok bool) {
	if v, ok := l.keyToElement.Load(key); ok {
		element := v.(*lruElement)
		l.mu.Lock()
		l.unlink(element)
		l.link(element)
		l.mu.Unlock()
		return element.value, true
	}
	return nil, false
}

func (l *lru) GetKeyFromValue(value interface{}) (key interface{}, ok bool) {
	if k, ok := l.valueToElement.Load(value); ok {
		element := k.(*lruElement)
		l.mu.Lock()
		l.unlink(element)
		l.link(element)
		l.mu.Unlock()
		return element.key, true
	}
	return nil, false
}

func (l *lru) Put(key, value interface{}) {
	l.mu.Lock()
	if v, ok := l.keyToElement.Load(key); ok {
		element := v.(*lruElement)
		element.value = value
		l.unlink(element)
		l.link(element)
	} else {
		element := &lruElement{key: key, value: value}
		l.keyToElement.Store(key, element)
		l.valueToElement.Store(value, element)
		l.link(element)
		l.count++
	}
	if l.count > l.capacity {
		l.count--
		l.keyToElement.Delete(l.tail.key)
		l.valueToElement.Delete(l.tail.value)
		l.unlink(l.tail)
	}
	l.mu.Unlock()
}

func (l *lru) link(element *lruElement) {
	element.prev = nil
	element.next = l.head
	if l.head != nil {
		l.head.prev = element
	}
	l.head = element

	if l.tail == nil {
		l.tail = element
		l.tail.next = nil
	}
}

func (l *lru) unlink(element *lruElement) {
	if element == l.head {
		l.head = element.next
		if element.next != nil {
			element.next.prev = nil
		}
		element.next = nil
		return
	}
	if element == l.tail {
		l.tail = element.prev
		if element.prev != nil {
			element.prev.next = nil
		}
		element.prev = nil
		return
	}
	element.prev.next = element.next
	element.next.prev = element.prev
}
