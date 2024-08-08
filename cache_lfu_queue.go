package cache

type LFUItem struct {
	key  uint64
	freq uint64
}

type PriorityQueue []*LFUItem

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].freq < pq[j].freq
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(item interface{}) {
	*pq = append(*pq, item.(*LFUItem))
}

func (pq *PriorityQueue) Pop() interface{} {
	prev := *pq
	if len(prev) == 0 {
		return nil
	}
	item := prev[len(prev)-1]
	prev[len(prev)-1] = nil
	*pq = prev[0 : len(prev)-1]
	return item
}
