package main

import "github.com/qeedquan/go-media/sdl"

type Node struct {
	parent  sdl.Point
	typ     int
	f, g, h int
}

func newNode() *Node {
	return &Node{
		parent: sdl.Point{-1, -1},
		typ:    -1,
		f:      -1,
		g:      -1,
		h:      -1,
	}
}

type Path struct {
	grid      []*Node
	size      sdl.Point
	start     sdl.Point
	end       sdl.Point
	current   sdl.Point
	neighbors []sdl.Point

	open   []sdl.Point
	closed []sdl.Point

	chain    string
	chainRev string
}

func newPath() *Path {
	return &Path{
		start:   sdl.Point{-1, -1},
		end:     sdl.Point{-1, -1},
		current: sdl.Point{-1, -1},
		neighbors: []sdl.Point{
			{-1, 0},
			{1, 0},
			{0, -1},
			{0, 1},
		},
	}
}

func (p *Path) Resize(size sdl.Point) {
	p.grid = make([]*Node, size.X*size.Y)
	p.size = size

	for y := int32(0); y < p.size.Y; y++ {
		for x := int32(0); x < p.size.X; x++ {
			i := sdl.Point{x, y}
			p.Set(i, newNode())
			p.SetType(i, 0)
		}
	}
}

func (p *Path) Set(pos sdl.Point, node *Node)  { p.grid[p.unfold(pos)] = node }
func (p *Path) SetType(pos sdl.Point, typ int) { p.node(pos).typ = typ }

func (p *Path) GetType(pos sdl.Point) int { return p.node(pos).typ }

func (p *Path) addOpen(pos sdl.Point)   { p.open = append(p.open, pos) }
func (p *Path) addClosed(pos sdl.Point) { p.closed = append(p.closed, pos) }

func (p *Path) inOpen(pos sdl.Point) bool   { return inList(p.open, pos) }
func (p *Path) inClosed(pos sdl.Point) bool { return inList(p.closed, pos) }

func (p *Path) unfold(pos sdl.Point) int { return int(pos.Y*p.size.X + pos.X) }
func (p *Path) node(pos sdl.Point) *Node { return p.grid[p.unfold(pos)] }

func (p *Path) setF(pos sdl.Point, v int) { p.node(pos).f = v }
func (p *Path) setG(pos sdl.Point, v int) { p.node(pos).g = v }
func (p *Path) setH(pos sdl.Point, v int) { p.node(pos).h = v }

func (p *Path) getF(pos sdl.Point) int { return p.node(pos).f }
func (p *Path) getG(pos sdl.Point) int { return p.node(pos).g }

func (p *Path) lowestF() sdl.Point {
	lv := 1000
	lp := sdl.Point{-1, -1}

	for _, op := range p.open {
		if v := p.getF(op); v < lv {
			lv = v
			lp = op
		}
	}
	return lp
}

func (p *Path) removeOpen(pos sdl.Point) {
	for i := range p.open {
		if p.open[i] == pos {
			copy(p.open[i:], p.open[i+1:])
			p.open = p.open[:len(p.open)-1]
			break
		}
	}
}

func (p *Path) calcH(pos sdl.Point) {
	p.grid[p.unfold(pos)].h = abs(int(pos.X-p.end.X)) + abs(int(pos.Y-p.end.Y))
}

func (p *Path) calcF(pos sdl.Point) {
	n := p.node(pos)
	n.f = n.g + n.h
}

func (p *Path) setParent(pos, parent sdl.Point)   { p.node(pos).parent = parent }
func (p *Path) getParent(pos sdl.Point) sdl.Point { return p.node(pos).parent }

func (p *Path) reset() {
	p.chainRev = ""
	p.chain = ""
	p.current = sdl.Point{-1, -1}
	p.open = p.open[:0]
	p.closed = p.closed[:0]
}

func (p *Path) Find(start, end sdl.Point) (chain string, found bool) {
	p.reset()

	p.start = start
	p.end = end

	p.addOpen(p.start)
	p.setG(p.start, 0)
	p.setH(p.start, 0)
	p.setF(p.start, 0)

	var f sdl.Point
	for {
		f = p.lowestF()
		if f == p.end || (f == sdl.Point{-1, -1}) {
			break
		}

		p.current = f
		p.removeOpen(p.current)
		p.addClosed(p.current)

		for _, o := range p.neighbors {
			n := p.current.Add(o)
			if n.X < 0 || n.Y < 0 || n.X >= p.size.X || n.Y >= p.size.Y {
				continue
			}
			if p.GetType(n) == 1 {
				continue
			}

			cost := p.getG(p.current) + 10
			if p.inOpen(n) && cost < p.getG(n) {
				p.removeOpen(n)
			}

			if !p.inOpen(n) && !p.inClosed(n) {
				p.addOpen(n)
				p.setG(n, cost)
				p.calcH(n)
				p.calcF(n)
				p.setParent(n, p.current)
			}
		}
	}

	if (f == sdl.Point{-1, -1}) {
		return "", false
	}

	p.current = p.end
	for p.current != p.start {
		switch {
		case p.current.X > p.getParent(p.current).X:
			p.chainRev += "R"
		case p.current.X < p.getParent(p.current).X:
			p.chainRev += "L"
		case p.current.Y > p.getParent(p.current).Y:
			p.chainRev += "D"
		case p.current.Y < p.getParent(p.current).Y:
			p.chainRev += "U"
		}
		p.current = p.getParent(p.current)
		p.SetType(p.current, 4)
	}

	for i := len(p.chainRev) - 1; i >= 0; i-- {
		p.chain += string(p.chainRev[i])
	}

	p.SetType(p.start, 2)
	p.SetType(p.end, 3)

	return p.chain, true
}

func inList(list []sdl.Point, pos sdl.Point) bool {
	for _, l := range list {
		if l == pos {
			return true
		}
	}
	return false
}
