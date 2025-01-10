package main

// This file was ported and slightly modified from the OpenRSC server
// The code was originally written by Aenge

import (
	"math"
)

const (
	initState = iota
	openState
	closedState

	_south = iota
	_southWest
	_west
	_northWest
	_north
	_northEast
	_east
	_southEast

	basicCost = 10
	diagCost  = 14
)

type astarPathFinder struct {
	depth       int
	costBoard   [][]*node
	worldStart  *point
	pointStart  *point
	pointEnd    *point
	path        *path
	openNodes   []*node
	closedNodes []*node
}

type point struct {
	x int
	y int
}

type node struct {
	fCost        int
	gCost        int
	hCost        int
	state        int
	southBlocked bool
	northBlocked bool
	westBlocked  bool
	eastBlocked  bool
	position     *point
	parent       *point
}

type path struct {
	waypoints Deque
	startX    int
	startY    int
}

type longPath struct {
	points    []*point
	currentPt int
	length    int
	script    *script
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func calcDistance(one, two *point) int {
	xdiff := abs(one.x - two.x)
	ydiff := abs(one.y - two.y)

	var (
		shortL int
		longL  int
	)
	if xdiff > ydiff {
		shortL = ydiff
	} else {
		shortL = xdiff
	}
	if xdiff > ydiff {
		longL = xdiff
	} else {
		longL = ydiff
	}

	return shortL*diagCost + (longL-shortL)*basicCost
}

func newAstarPathFinder(c *client, start *point, end *point, depth int, skipLocal bool) *astarPathFinder {
	p := &astarPathFinder{
		costBoard:   nil,
		openNodes:   []*node{},
		closedNodes: []*node{},
		worldStart:  start,
		pointStart:  &point{depth, depth},
		pointEnd:    &point{start.x + depth - end.x, end.y - (start.y - depth)},
		depth:       depth,
		path:        newPath(start.x, start.y),
	}
	p.generateTraversalInfo(c, start, depth, skipLocal)

	return p
}

func newNode(x int, y int) *node {
	return &node{
		position: &point{x, y},
		state:    initState,
	}
}

func (p *astarPathFinder) buildPath() *path {
	parent := p.closedNodes[len(p.closedNodes)-1].parent
	endNode := p.costBoard[parent.x][parent.y]
	for endNode != nil {
		worldX := p.worldStart.x + p.depth - endNode.position.x
		worldY := p.worldStart.y - p.depth + endNode.position.y
		if endNode.parent == nil {
			endNode = nil
		} else {
			p.path.addWaypoint(worldX, worldY)
			endNode = p.costBoard[endNode.parent.x][endNode.parent.y]
		}
	}
	return p.path
}

func (p *astarPathFinder) generateTraversalInfo(c *client, center *point, depth int, skipLocal bool) {
	if depth < 1 {
		return
	}
	p.costBoard = make([][]*node, 2*depth+1)
	p.initBoard(p.costBoard, 2*depth+1)
	var curPosX int
	var curPosY int

	for x := -depth; x <= depth; x++ {
		for y := -depth; y <= depth; y++ {
			cdx, cdy := center.x-x, center.y+y
			_, ok := c.world.pathWalkerMap[int32(cdx<<16)|int32(cdy)]
			curPosX = x + depth
			curPosY = y + depth
			if !withinWorld(cdx, cdy) || ok {
				if y < depth {
					p.costBoard[curPosX][curPosY+1].northBlocked = true
				}
				if x > -depth {
					p.costBoard[curPosX-1][curPosY].eastBlocked = true
				}
				if y > -depth {
					p.costBoard[curPosX][curPosY-1].southBlocked = true
				}
				if x < depth {
					p.costBoard[curPosX+1][curPosY].westBlocked = true
				}
			}
			var tile int32
			if skipLocal {
				tile = c.world.getTile(cdx, cdy)
			} else {
				tile = c.combinedTile(cdx, cdy)
			}
			if tile == math.MaxInt32 {
				continue
			}
			if tile&(fullBlockA|fullBlockB|fullBlockC) != 0 {
				if y < depth {
					p.costBoard[curPosX][curPosY+1].northBlocked = true
				}
				if x > -depth {
					p.costBoard[curPosX-1][curPosY].eastBlocked = true
				}
				if y > -depth {
					p.costBoard[curPosX][curPosY-1].southBlocked = true
				}
				if x < depth {
					p.costBoard[curPosX+1][curPosY].westBlocked = true
				}
			} else {
				if !p.costBoard[curPosX][curPosY].southBlocked {
					p.costBoard[curPosX][curPosY].southBlocked = (tile & southBlocked) != 0
				}
				if !p.costBoard[curPosX][curPosY].westBlocked {
					p.costBoard[curPosX][curPosY].westBlocked = (tile & westBlocked) != 0
				}
				if !p.costBoard[curPosX][curPosY].northBlocked {
					p.costBoard[curPosX][curPosY].northBlocked = (tile & northBlocked) != 0
				}
				if !p.costBoard[curPosX][curPosY].eastBlocked {
					p.costBoard[curPosX][curPosY].eastBlocked = (tile & eastBlocked) != 0
				}
			}
		}
	}
}

func (p *astarPathFinder) findNextNode() *node {
	var (
		minimum = math.MaxInt
		minNode *node
	)
	for _, node := range p.openNodes {
		if node.hCost < minimum {
			minimum = node.hCost
			minNode = node
		} else if node.hCost == minimum && node.fCost < minNode.fCost {
			minNode = node
		}
	}
	return minNode
}

func (p *astarPathFinder) findPath() *path {
	if p.depth < 1 {
		return nil
	}
	if p.pointStart.x == p.pointEnd.x && p.pointStart.y == p.pointEnd.y {
		return nil
	}
	p.costBoard[p.depth][p.depth].selectNode(p)
	for {
		next := p.findNextNode()
		if next == nil {
			return nil
		}
		if next.position.x == p.pointEnd.x && next.position.y == p.pointEnd.y {
			p.closedNodes = append(p.closedNodes, next)
			return p.buildPath()
		}
		next.selectNode(p)
	}
}

func (p *astarPathFinder) initBoard(board [][]*node, depth int) {
	if board == nil {
		return
	}
	for i := 0; i < len(p.costBoard); i++ {
		board[i] = make([]*node, depth)
		for j := 0; j < len(p.costBoard[i]); j++ {
			board[i][j] = newNode(i, j)
		}
	}
}

func (n *node) calcGCost(p *astarPathFinder) {
	n.gCost = calcDistance(n.position, p.pointEnd)
}

func (n *node) calcHCost() {
	n.hCost = n.fCost + n.gCost
}

func (n *node) getNeighbor(p *astarPathFinder, dir int) *node {
	switch dir {
	case _south:
		if n.position.y < 2*p.depth {
			return p.costBoard[n.position.x][n.position.y+1]
		} else {
			return nil
		}
	case _southWest:
		if n.position.y < 2*p.depth && n.position.x > 0 {
			return p.costBoard[n.position.x-1][n.position.y+1]
		} else {
			return nil
		}
	case _west:
		if n.position.x > 0 {
			return p.costBoard[n.position.x-1][n.position.y]
		} else {
			return nil
		}
	case _northWest:
		if n.position.y > 0 && n.position.x > 0 {
			return p.costBoard[n.position.x-1][n.position.y-1]
		} else {
			return nil
		}
	case _north:
		if n.position.y > 0 {
			return p.costBoard[n.position.x][n.position.y-1]
		} else {
			return nil
		}
	case _northEast:
		if n.position.y > 0 && n.position.x < 2*p.depth {
			return p.costBoard[n.position.x+1][n.position.y-1]
		} else {
			return nil
		}
	case _east:
		if n.position.x < 2*p.depth {
			return p.costBoard[n.position.x+1][n.position.y]
		} else {
			return nil
		}
	case _southEast:
		if n.position.y < 2*p.depth && n.position.x < 2*p.depth {
			return p.costBoard[n.position.x+1][n.position.y+1]
		} else {
			return nil
		}
	}
	return nil
}

func findIndex(node *node, ns []*node) int {
	for i, n := range ns {
		if node == n {
			return i
		}
	}
	return -1
}

func (n *node) selectNode(p *astarPathFinder) {
	if n.state == openState {
		idx := findIndex(n, p.openNodes)
		p.openNodes = append(p.openNodes[:idx], p.openNodes[idx+1:]...)
	}
	n.state = closedState
	p.closedNodes = append(p.closedNodes, n)
	neighbor := (*node)(nil)
	neighbor = n.getNeighbor(p, _south)
	if !n.southBlocked && neighbor != nil {
		neighbor.update(p, n, basicCost)
	}
	neighbor = n.getNeighbor(p, _west)
	if !n.westBlocked && neighbor != nil {
		neighbor.update(p, n, basicCost)
	}
	neighbor = n.getNeighbor(p, _north)
	if !n.northBlocked && neighbor != nil {
		neighbor.update(p, n, basicCost)
	}
	neighbor = n.getNeighbor(p, _east)
	if !n.eastBlocked && neighbor != nil {
		neighbor.update(p, n, basicCost)
	}
	neighbor = n.getNeighbor(p, _southWest)
	if !(n.southBlocked || n.westBlocked) && !diagBlocked(p, n, _southWest) && neighbor != nil {
		neighbor.update(p, n, diagCost)
	}
	neighbor = n.getNeighbor(p, _northWest)
	if !(n.northBlocked || n.westBlocked) && !diagBlocked(p, n, _northWest) && neighbor != nil {
		neighbor.update(p, n, diagCost)
	}
	neighbor = n.getNeighbor(p, _northEast)
	if !(n.northBlocked || n.eastBlocked) && !diagBlocked(p, n, _northEast) && neighbor != nil {
		neighbor.update(p, n, diagCost)
	}
	neighbor = n.getNeighbor(p, _southEast)
	if !(n.southBlocked || n.eastBlocked) && !diagBlocked(p, n, _southEast) && neighbor != nil {
		neighbor.update(p, n, diagCost)
	}
}

func diagBlocked(a *astarPathFinder, n *node, dir int) bool {
	var (
		neighbor1 *node
		neighbor2 *node
	)
	if dir == _southWest {
		neighbor1 = n.getNeighbor(a, _west)
		if neighbor1 != nil {
			neighbor2 = n.getNeighbor(a, _south)
			if neighbor2 != nil {
				if !neighbor1.southBlocked && !neighbor2.westBlocked {
					return false
				}
			}
		}
	} else if dir == _northWest {
		neighbor1 = n.getNeighbor(a, _west)
		if neighbor1 != nil {
			neighbor2 = n.getNeighbor(a, _north)
			if neighbor2 != nil {
				if !neighbor1.northBlocked && !neighbor2.westBlocked {
					return false
				}
			}
		}
	} else if dir == _northEast {
		neighbor1 = n.getNeighbor(a, _east)
		if neighbor1 != nil {
			neighbor2 = n.getNeighbor(a, _north)
			if neighbor2 != nil {
				if !neighbor1.northBlocked && !neighbor2.eastBlocked {
					return false
				}
			}
		}
	} else if dir == _southEast {
		neighbor1 = n.getNeighbor(a, _east)
		if neighbor1 != nil {
			neighbor2 = n.getNeighbor(a, _south)
			if neighbor2 != nil {
				if !neighbor1.southBlocked && !neighbor2.eastBlocked {
					return false
				}
			}
		}
	}
	return true
}

func (n *node) update(a *astarPathFinder, node *node, cost int) {
	if n.state == initState {
		n.state = openState
		n.fCost = node.fCost + cost
		n.calcGCost(a)
		a.openNodes = append(a.openNodes, n)
	} else if n.state == closedState {
		return
	} else {
		newFcost := node.fCost + cost
		if newFcost > n.fCost {
			return
		}
		n.fCost = newFcost
	}
	n.calcHCost()
	n.parent = node.position
}

func newPath(startX, startY int) *path {
	return &path{
		startX:    startX,
		startY:    startY,
		waypoints: NewDeque(),
	}
}

func (p *path) addWaypoint(x int, y int) {
	p.waypoints.PushFront(&point{x, y})
}
