package calculator

import (
	"calc/internal/domain"
	"sync"
)

type node struct {
	value  *domain.Arbitrage
	next   *node
	before *node
}

func (n *node) isTail() bool {
	return n.next == nil
}

func (n *node) isHead() bool {
	return n.before == nil
}

func (n *node) insertBefore(node *node) {
	node.next = n
	node.before = n.before
	if n.before != nil {
		n.before.next = node
	}
	n.before = node
}

func (n *node) insertAfter(node *node) {
	node.next = n.next
	node.before = n
	if n.next != nil {
		n.next.before = node
	}
	n.next = node
}

// Top это двусвязный список для хранения в отсортированном виде результатов расчета по Box strategy
type Top struct {
	mu     sync.RWMutex
	names  map[string]*node
	head   *node
	tail   *node
	length int
}

func NewTop() *Top {
	return &Top{
		names: make(map[string]*node),
	}
}

func (o *Top) search(head *node, tail *node, percent float64) *node {
	for head.value.Profit > percent && tail.value.Profit < percent {
		head = head.next
		tail = tail.before
	}
	if head.value.Profit <= percent {
		return head
	}
	return tail
}

func (o *Top) remove(node *node) {
	if !node.isHead() {
		node.before.next = node.next
	} else {
		if !node.isTail() {
			node.next.before = nil
		}
		o.head = node.next
	}
	if !node.isTail() {
		node.next.before = node.before
	} else {
		if !node.isHead() {
			node.before.next = nil
		}
		o.tail = node.before
	}
}

func (o *Top) Put(v *domain.Arbitrage) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.head == nil {
		n := &node{
			value: v,
		}
		o.head = n
		o.tail = n
		o.length++

		o.names[v.Pair] = n

		return
	}

	if foundNode, ok := o.names[v.Pair]; ok {
		head := o.head
		if foundNode.value.Profit > v.Profit {
			head = foundNode
		}
		currentNode := o.search(head, o.tail, v.Profit)
		foundNode.value = v

		if currentNode != foundNode {
			o.remove(foundNode)
			if currentNode.value.Profit > v.Profit {
				currentNode.insertAfter(foundNode)
				if currentNode == o.tail {
					o.tail = foundNode
				}
			} else {
				currentNode.insertBefore(foundNode)
				if currentNode == o.head {
					o.head = foundNode
				}
			}
		}
	} else {
		if v.Profit < o.tail.value.Profit {
			n := &node{
				value:  v,
				before: o.tail,
			}
			o.tail.next = n
			o.tail = n

			o.names[v.Pair] = n
		} else {
			n := &node{
				value: v,
			}

			currentNode := o.search(o.head, o.tail, v.Profit)
			if currentNode.value.Profit > v.Profit {
				currentNode.insertAfter(n)
				if currentNode == o.tail {
					o.tail = n
				}
			} else {
				currentNode.insertBefore(n)
				if currentNode == o.head {
					o.head = n
				}
			}

			o.names[v.Pair] = n
		}
		o.length++
	}
}

func (o *Top) Get() []*domain.Arbitrage {
	resp := make([]*domain.Arbitrage, 0)

	currentNode := o.head
	for currentNode != nil {
		resp = append(resp, currentNode.value)
		currentNode = currentNode.next
	}

	return resp
}
