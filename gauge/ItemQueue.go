/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package gauge

type ItemQueue struct {
	Items []Item
}

func (queue *ItemQueue) Next() Item {
	if len(queue.Items) > 0 {
		next := queue.Items[0]
		queue.Items = queue.Items[1:]
		return next
	}
	return nil
}

func (queue *ItemQueue) Peek() Item {
	if len(queue.Items) > 0 {
		return queue.Items[0]
	}
	return nil
}
