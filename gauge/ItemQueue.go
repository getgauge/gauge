package gauge

type ItemQueue struct {
	Items []Item
}

func (queue *ItemQueue) Next() Item {
	if queue.hasNext() {
		next := queue.Items[0]
		queue.Items = queue.Items[1:]
		return next
	}
	return nil
}

func (queue *ItemQueue) Peek() Item {
	if queue.hasNext() {
		return queue.Items[0]
	}
	return nil
}

func (queue *ItemQueue) hasNext() bool {
	if len(queue.Items) > 0 {
		return true
	}
	return false
}
