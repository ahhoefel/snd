package strum

import "math/rand"

const (
	springConstant = 0.3
	heatConstant   = 0.9999
)

type Strum struct {
	Length float64
	Nodes  []Node
}

type Node struct {
	X, Y, vx, vy float64
}

func New(n int, length, pull float64) *Strum {
	step := length / float64(n+1)
	var x float64
	var nodes []Node
	for i := 0; i < n; i++ {
		x += step
		nodes = append(nodes, Node{x, rand.Float64() * pull, 0, 0})
	}
	return &Strum{length, nodes}
}

func (s *Strum) Step() {
	for i, n := range s.Nodes {
		s.Nodes[i] = Node{n.X + n.vx, n.Y + n.vy, n.vx, n.vy}
	}
	for i, curr := range s.Nodes {
		var prev, next Node
		if i == 0 {
			prev = Node{0, 0, 0, 0}
		} else {
			prev = s.Nodes[i-1]
		}
		if i == len(s.Nodes)-1 {
			next = Node{s.Length, 0, 0, 0}
		} else {
			next = s.Nodes[i+1]
		}
		ax := springConstant * (prev.X + next.X - 2*curr.X)
		ay := springConstant * (prev.Y + next.Y - 2*curr.Y)
		s.Nodes[i] = Node{curr.X, curr.Y, heatConstant * (curr.vx + ax), heatConstant * (curr.vy + ay)}
	}
}

func (s *Strum) Pull() float64 {
	return s.Nodes[200].Y
}
