package hnsw

import (
	"fmt"
	"math"
	"math/rand"
	"sync"

	"github.com/AnishAdkikar/og/distqueue"
	"github.com/AnishAdkikar/og/f32"
)

type Point []float32

func (a Point) Size() int {
	return len(a) * 4
}

type node struct {
	sync.RWMutex
	Text    string
	P       Point
	Level   int
	Friends [][]uint32
}

type Hnsw struct {
	sync.RWMutex
	M              int
	M0             int
	efConstruction int
	Nodes          []node
	LevelMult      float64
	maxLayer       int
	enterpoint     uint32
	Size           int
}

var DistFunc = f32.L2Squared8AVX

func New(M int, efConstruction int) *Hnsw {
	h := Hnsw{}
	h.M = M
	h.LevelMult = 1 / math.Log(float64(M))
	h.efConstruction = efConstruction
	h.M0 = 2 * M
	h.Nodes = []node{}
	h.Size = 0
	return &h
}

func (h *Hnsw) Grow(size int) {
	if size+1 <= len(h.Nodes) {
		return
	}
	newNodes := make([]node, len(h.Nodes), size+1)
	copy(newNodes, h.Nodes)
	h.Nodes = newNodes
}

func (h *Hnsw) Add(q Point, id uint32, textdata string) {
	// if id == 0 {
	if len(h.Nodes) == 0 {
		h.Nodes = append(h.Nodes, node{Level: 0, P: q, Text: ""})
		h.enterpoint = 0
		h.Size++
	}
	// return
	// }
	if int(id) < 0 {
		fmt.Printf("Warning: Attempt to add a node with negative ID (%d). Skipping.", id)
		return
	}
	if int(id)+1 > len(h.Nodes) {
		newSize := max(int(id)+1, len(h.Nodes)*2)
		h.Grow(newSize)
	}
	curlevel := int(math.Floor(-math.Log(rand.Float64() * h.LevelMult)))

	epID := h.enterpoint
	currentMaxLayer := h.Nodes[epID].Level
	ep := &distqueue.Item{ID: h.enterpoint, D: DistFunc(h.Nodes[h.enterpoint].P, q)}

	newID := id
	newNode := node{P: q, Level: curlevel, Friends: make([][]uint32, min(curlevel, currentMaxLayer)+1), Text: textdata}

	for level := currentMaxLayer; level > curlevel; level-- {
		changed := true
		for changed {
			changed = false
			for _, i := range h.getFriends(ep.ID, level) {
				d := DistFunc(h.Nodes[i].P, q)
				if d < ep.D {
					ep = &distqueue.Item{ID: i, D: d}
					changed = true
				}
			}
		}
	}

	for level := min(curlevel, currentMaxLayer); level >= 0; level-- {

		resultSet := &distqueue.DistQueueClosestLast{}
		h.searchAtLayer(q, resultSet, h.efConstruction, ep, level)

		h.getNeighborsByHeuristicClosestLast(resultSet, h.M)
		newNode.Friends[level] = make([]uint32, resultSet.Len())
		for i := resultSet.Len() - 1; i >= 0; i-- {
			item := resultSet.Pop()
			newNode.Friends[level][i] = item.ID
		}
	}

	h.Lock()
	if len(h.Nodes) < int(newID)+1 {
		h.Nodes = h.Nodes[0 : newID+1]
	}
	h.Nodes[newID] = newNode
	h.Unlock()
	for level := min(curlevel, currentMaxLayer); level >= 0; level-- {
		for _, n := range newNode.Friends[level] {
			h.Link(n, newID, level)
		}
	}

	h.Lock()
	if curlevel > h.maxLayer {
		h.maxLayer = curlevel
		h.enterpoint = newID
	}
	h.Unlock()
}

func (h *Hnsw) getFriends(n uint32, level int) []uint32 {
	if len(h.Nodes[n].Friends) < level+1 {
		return make([]uint32, 0)
	}
	return h.Nodes[n].Friends[level]
}

func (h *Hnsw) Link(first, second uint32, level int) {

	maxL := h.M
	if level == 0 {
		maxL = h.M0
	}

	h.RLock()
	node := &h.Nodes[first]
	h.RUnlock()

	node.Lock()

	if len(node.Friends) < level+1 {
		for j := len(node.Friends); j <= level; j++ {

			node.Friends = append(node.Friends, make([]uint32, 0, maxL))
		}
		node.Friends[level] = node.Friends[level][0:1]
		node.Friends[level][0] = second

	} else {
		node.Friends[level] = append(node.Friends[level], second)
	}

	l := len(node.Friends[level])

	if l > maxL {
		resultSet := &distqueue.DistQueueClosestFirst{Size: len(node.Friends[level])}

		for _, n := range node.Friends[level] {
			resultSet.Push(n, DistFunc(node.P, h.Nodes[n].P))
		}
		h.getNeighborsByHeuristicClosestFirst(resultSet, maxL)

		node.Friends[level] = node.Friends[level][0:maxL]
		for i := 0; i < maxL; i++ {
			item := resultSet.Pop()
			node.Friends[level][i] = item.ID
		}
	}
	node.Unlock()
}

func (h *Hnsw) getNeighborsByHeuristicClosestLast(resultSet1 *distqueue.DistQueueClosestLast, M int) {
	if resultSet1.Len() <= M {
		return
	}
	resultSet := &distqueue.DistQueueClosestFirst{Size: resultSet1.Len()}
	tempList := &distqueue.DistQueueClosestFirst{Size: resultSet1.Len()}
	result := make([]*distqueue.Item, 0, M)
	for resultSet1.Len() > 0 {
		resultSet.PushItem(resultSet1.Pop())
	}
	for resultSet.Len() > 0 {
		if len(result) >= M {
			break
		}
		e := resultSet.Pop()
		good := true
		for _, r := range result {
			if DistFunc(h.Nodes[r.ID].P, h.Nodes[e.ID].P) < e.D {
				good = false
				break
			}
		}
		if good {
			result = append(result, e)
		} else {
			tempList.PushItem(e)
		}
	}
	for len(result) < M && tempList.Len() > 0 {
		result = append(result, tempList.Pop())
	}
	for _, item := range result {
		resultSet1.PushItem(item)
	}
}

func (h *Hnsw) getNeighborsByHeuristicClosestFirst(resultSet *distqueue.DistQueueClosestFirst, M int) {
	if resultSet.Len() <= M {
		return
	}
	tempList := &distqueue.DistQueueClosestFirst{Size: resultSet.Len()}
	result := make([]*distqueue.Item, 0, M)
	for resultSet.Len() > 0 {
		if len(result) >= M {
			break
		}
		e := resultSet.Pop()
		good := true
		for _, r := range result {
			if DistFunc(h.Nodes[r.ID].P, h.Nodes[e.ID].P) < e.D {
				good = false
				break
			}
		}
		if good {
			result = append(result, e)
		} else {
			tempList.PushItem(e)
		}
	}
	for len(result) < M && tempList.Len() > 0 {
		result = append(result, tempList.Pop())
	}
	resultSet.Reset()

	for _, item := range result {
		resultSet.PushItem(item)
	}
}

func (h *Hnsw) searchAtLayer(q Point, resultSet *distqueue.DistQueueClosestLast, efConstruction int, ep *distqueue.Item, level int) {

	visited := make(map[uint32]bool)

	candidates := &distqueue.DistQueueClosestFirst{Size: efConstruction * 3}

	visited[ep.ID] = true
	candidates.Push(ep.ID, ep.D)

	resultSet.Push(ep.ID, ep.D)

	for candidates.Len() > 0 {
		_, lowerBound := resultSet.Top()
		c := candidates.Pop()

		if c.D > lowerBound {
			break
		}

		if len(h.Nodes[c.ID].Friends) >= level+1 {
			Friends := h.Nodes[c.ID].Friends[level]
			for _, n := range Friends {
				if !visited[n] {
					visited[n] = true
					d := DistFunc(q, h.Nodes[n].P)
					_, topD := resultSet.Top()
					if resultSet.Len() < efConstruction {
						item := resultSet.Push(n, d)
						candidates.PushItem(item)
					} else if topD > d {
						item := resultSet.PopAndPush(n, d)
						candidates.PushItem(item)
					}
				}
			}
		}
	}
}

func (h *Hnsw) Search(q Point, ef int, K int) []string {

	h.RLock()
	currentMaxLayer := h.maxLayer
	ep := &distqueue.Item{ID: h.enterpoint, D: DistFunc(h.Nodes[h.enterpoint].P, q)}
	h.RUnlock()

	resultSet := &distqueue.DistQueueClosestLast{Size: ef + 1}
	for level := currentMaxLayer; level > 0; level-- {
		changed := true
		for changed {
			changed = false
			for _, i := range h.getFriends(ep.ID, level) {
				d := DistFunc(h.Nodes[i].P, q)
				if d < ep.D {
					ep.ID, ep.D = i, d
					changed = true
				}
			}
		}
	}
	h.searchAtLayer(q, resultSet, ef, ep, 0)

	for resultSet.Len() > K {
		resultSet.Pop()
	}
	var textData []string
	// var textData []uint32
	results := resultSet.Items()
	for _, item := range results {
		nodeID := item.ID
		nodeText := h.Nodes[nodeID].Text
		textData = append(textData, nodeText)
		// textData = append(textData, nodeID)
		// fmt.Printf("Node ID: %d, Text: %s\n", nodeID, nodeText)
	}
	return textData
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
