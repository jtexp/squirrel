package main

import (
	"github.com/squirrel-land/squirrel"
	"math"
	"sync"
)

type PositionManager struct {
	pos []*squirrel.Position
	mu  []*sync.RWMutex

	isEnabled      []bool
	enabledChanged []chan<- []int
	muEnabled      *sync.RWMutex // mutex for isEnabled, enabled and enabledChanged
}

func NewPositionManager(size int) squirrel.PositionManager {
	ret := new(PositionManager)
	ret.pos = make([]*squirrel.Position, size)
	ret.mu = make([]*sync.RWMutex, size)
	ret.isEnabled = make([]bool, size)
	ret.enabledChanged = make([]chan<- []int, 0)
	ret.muEnabled = new(sync.RWMutex)
	for i := 0; i < size; i++ {
		ret.pos[i] = &squirrel.Position{0, 0, 0}
		ret.mu[i] = new(sync.RWMutex)
	}
	return ret
}

func (p *PositionManager) Capacity() int {
	return len(p.pos)
}

// Get returns a copy of Position at given index. Avoid this if possible. It
// causes copying Position struct.
func (p *PositionManager) Get(index int) squirrel.Position {
	p.mu[index].RLock()
	defer p.mu[index].RUnlock()
	return *(p.pos[index])
}

// Distance calculates Euclidean distance between positions at index1 and
// index2.
func (p *PositionManager) Distance(index1, index2 int) float64 {
	p.mu[index1].RLock()
	defer p.mu[index1].RUnlock()
	p.mu[index2].RLock()
	defer p.mu[index2].RUnlock()
	dist := math.Sqrt(math.Pow(p.pos[index1].X-p.pos[index2].X, 2) + math.Pow(p.pos[index1].Y-p.pos[index2].Y, 2) + math.Pow(p.pos[index1].Height-p.pos[index2].Height, 2))
	return dist
}

func (p *PositionManager) Set(index int, x, y, height float64) {
	p.mu[index].Lock()
	defer p.mu[index].Unlock()
	p.pos[index].X = x
	p.pos[index].Y = y
	p.pos[index].Height = height
}

func (p *PositionManager) SetPosition(index int, pos *squirrel.Position) {
	p.mu[index].Lock()
	defer p.mu[index].Unlock()
	p.pos[index].X = pos.X
	p.pos[index].Y = pos.Y
	p.pos[index].Height = pos.Height
}

// Enable marks a node enabled.
func (p *PositionManager) Enable(index int) {
	p.muEnabled.Lock()
	defer p.muEnabled.Unlock()
	p.isEnabled[index] = true
	p.notifyEnabledChanged()
}

// Disable marks a node disabled.
func (p *PositionManager) Disable(index int) {
	p.muEnabled.Lock()
	defer p.muEnabled.Unlock()
	p.isEnabled[index] = false
	p.notifyEnabledChanged()
}

func (p *PositionManager) IsEnabled(index int) bool {
	p.muEnabled.RLock()
	defer p.muEnabled.RUnlock()
	return p.isEnabled[index]
}

func (p *PositionManager) calculateEnabled() []int {
	e := make([]int, 0)
	for i, v := range p.isEnabled {
		if v {
			e = append(e, i)
		}
	}
	return e
}

func (p *PositionManager) Enabled() []int {
	p.muEnabled.RLock()
	defer p.muEnabled.RUnlock()
	return p.calculateEnabled()
}

// RegisterEnabledChanged registers a channel used to receive a slice of
// indices of all enabled nodes.  Slice is sent into channel each time any node
// is enabled/disabled.
func (p *PositionManager) RegisterEnabledChanged(channel chan<- []int) {
	p.muEnabled.Lock()
	defer p.muEnabled.Unlock()
	p.enabledChanged = append(p.enabledChanged, channel)
}

func (p *PositionManager) notifyEnabledChanged() {
	for _, c := range p.enabledChanged {
		c <- p.calculateEnabled()
	}
}