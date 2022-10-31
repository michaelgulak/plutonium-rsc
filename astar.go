package main

import (
	"container/list"
	"math"
)

var (
	wallNorth     int32 = 1 << 0
	wallEast      int32 = 1 << 1
	wallSouth     int32 = 1 << 2
	wallWest      int32 = 1 << 3
	fullBlockA    int32 = 1 << 4
	fullBlockB    int32 = 1 << 5
	fullBlockC    int32 = 1 << 6
	wallNorthEast       = wallNorth | wallEast
	wallNorthWest       = wallNorth | wallWest
	wallSouthEast       = wallSouth | wallEast
	wallSouthWest       = wallSouth | wallWest
	fullBlock           = fullBlockA | fullBlockB | fullBlockC
	westBlocked         = fullBlock | wallWest
	southBlocked        = fullBlock | wallSouth
	northBlocked        = fullBlock | wallNorth
	eastBlocked         = fullBlock | wallEast
)

type adjacentDirection int

const (
	_south adjacentDirection = iota
	_southWest
	_west
	_northWest
	_north
	_northEast
	_east
	_southEast
)

type nodeState int

const (
	initState nodeState = iota
	openState
	closedState
)

const basicCost = 10
const diagCost = 14

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

func newAstarPathFinder(c *client, start *point, end *point, depth int) (z *astarPathFinder) {
	z = &astarPathFinder{}
	z.costBoard = nil
	z.openNodes = []*node{}
	z.closedNodes = []*node{}
	z.worldStart = start
	z.pointStart = &point{depth, depth}
	z.pointEnd = &point{start.x + depth - end.x, end.y - (start.y - depth)}
	z.depth = depth
	z.path = newPath(start.x, start.y)
	z.generateTraversalInfo(c, start, depth)

	return z
}

type node struct {
	fCost        int
	gCost        int
	hCost        int
	state        nodeState
	southBlocked bool
	northBlocked bool
	westBlocked  bool
	eastBlocked  bool
	position     *point
	parent       *point
}

func newNode(x int, y int) (z *node) {
	z = &node{}
	z.southBlocked = false
	z.northBlocked = false
	z.westBlocked = false
	z.eastBlocked = false
	z.position = &point{x, y}
	z.fCost = 0
	z.gCost = 0
	z.hCost = 0
	z.state = initState
	return
}

func (z *astarPathFinder) buildPath() *path {
	parent := z.closedNodes[len(z.closedNodes)-1].parent
	endNode := z.costBoard[parent.x][parent.y]
	for endNode != nil {
		worldX := z.worldStart.x + z.depth - endNode.position.x
		worldY := z.worldStart.y - z.depth + endNode.position.y
		if endNode.parent == nil {
			endNode = nil
		} else {
			z.path.addDirect(worldX, worldY)
			endNode = z.costBoard[endNode.parent.x][endNode.parent.y]
		}
	}
	return z.path
}

func (z *astarPathFinder) generateTraversalInfo(c *client, center *point, depth int) {
	if depth < 1 {
		return
	}
	z.costBoard = make([][]*node, 2*depth+1)
	z.initBoard(z.costBoard, 2*depth+1)
	var curPosX int
	var curPosY int
	for x := -depth; x <= depth; x++ {
		for y := -depth; y <= depth; y++ {
			cdx, cdy := center.x-x, center.y+y
			if !withinWorld(cdx, cdy) {
				continue
			}
			_, ok := c.world.pathWalkerMap[int32(cdx<<16)|int32(cdy)]
			curPosX = x + depth
			curPosY = y + depth
			if ok {
				if y < depth {
					z.costBoard[curPosX][curPosY+1].northBlocked = true
				}
				if x > -depth {
					z.costBoard[curPosX-1][curPosY].eastBlocked = true
				}
				if y > -depth {
					z.costBoard[curPosX][curPosY-1].southBlocked = true
				}
				if x < depth {
					z.costBoard[curPosX+1][curPosY].westBlocked = true
				}
			}
			tile := c.combinedTile(cdx, cdy)
			if tile == math.MaxInt32 {
				continue
			}
			if tile&(fullBlockA|fullBlockB|fullBlockC) != 0 {
				if y < depth {
					z.costBoard[curPosX][curPosY+1].northBlocked = true
				}
				if x > -depth {
					z.costBoard[curPosX-1][curPosY].eastBlocked = true
				}
				if y > -depth {
					z.costBoard[curPosX][curPosY-1].southBlocked = true
				}
				if x < depth {
					z.costBoard[curPosX+1][curPosY].westBlocked = true
				}
			} else {
				if !z.costBoard[curPosX][curPosY].southBlocked {
					z.costBoard[curPosX][curPosY].southBlocked = (tile & southBlocked) != 0
				}
				if !z.costBoard[curPosX][curPosY].westBlocked {
					z.costBoard[curPosX][curPosY].westBlocked = (tile & westBlocked) != 0
				}
				if !z.costBoard[curPosX][curPosY].northBlocked {
					z.costBoard[curPosX][curPosY].northBlocked = (tile & northBlocked) != 0
				}
				if !z.costBoard[curPosX][curPosY].eastBlocked {
					z.costBoard[curPosX][curPosY].eastBlocked = (tile & eastBlocked) != 0
				}
			}
		}
	}
}

func (z *astarPathFinder) findNextNode() *node {
	minimum := math.MaxInt
	minNode := (*node)(nil)
	for _, node := range z.openNodes {
		if node.hCost < minimum {
			minimum = node.hCost
			minNode = node
		} else if node.hCost == minimum && node.fCost < minNode.fCost {
			minNode = node
		}
	}
	return minNode
}

func (z *astarPathFinder) findPath() *path {
	if z.depth < 1 {
		return nil
	}
	if z.pointStart.x == z.pointEnd.x && z.pointStart.y == z.pointEnd.y {
		return nil
	}
	z.costBoard[z.depth][z.depth].selectNode(z)
	for {
		next := z.findNextNode()
		if next == nil {
			return nil
		}
		if next.position.x == z.pointEnd.x && next.position.y == z.pointEnd.y {
			z.closedNodes = append(z.closedNodes, next)
			return z.buildPath()
		}
		next.selectNode(z)
	}
}

func (z *astarPathFinder) initBoard(board [][]*node, depth int) {
	if board == nil {
		return
	}
	for i := 0; i < len(z.costBoard); i++ {
		board[i] = make([]*node, depth)
		for j := 0; j < len(z.costBoard[i]); j++ {
			board[i][j] = newNode(i, j)
		}
	}
}

func (z *node) calcGCost(a *astarPathFinder) {
	z.gCost = calcDistance(z.position, a.pointEnd)
}

func (z *node) calcHCost() {
	z.hCost = z.fCost + z.gCost
}

func (z *node) getNeighbor(a *astarPathFinder, dir adjacentDirection) (n *node) {
	switch dir {
	case _south:
		if z.position.y < 2*a.depth {
			return a.costBoard[z.position.x][z.position.y+1]
		} else {
			return nil
		}
	case _southWest:
		if z.position.y < 2*a.depth && z.position.x > 0 {
			n := a.costBoard[z.position.x-1][z.position.y+1]
			return n
		} else {
			return nil
		}
	case _west:
		if z.position.x > 0 {
			return a.costBoard[z.position.x-1][z.position.y]
		} else {
			return nil
		}
	case _northWest:
		if z.position.y > 0 && z.position.x > 0 {
			n := a.costBoard[z.position.x-1][z.position.y-1]
			return n
		} else {
			return nil
		}
	case _north:
		if z.position.y > 0 {
			return a.costBoard[z.position.x][z.position.y-1]
		} else {
			return nil
		}
	case _northEast:
		if z.position.y > 0 && z.position.x < 2*a.depth {
			n := a.costBoard[z.position.x+1][z.position.y-1]
			return n
		} else {
			return nil
		}
	case _east:
		if z.position.x < 2*a.depth {
			return a.costBoard[z.position.x+1][z.position.y]
		} else {
			return nil
		}
	case _southEast:
		if z.position.y < 2*a.depth && z.position.x < 2*a.depth {
			n := a.costBoard[z.position.x+1][z.position.y+1]
			return n
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

func (z *node) selectNode(a *astarPathFinder) {
	if z.state == openState {
		idx := findIndex(z, a.openNodes)
		a.openNodes = append(a.openNodes[:idx], a.openNodes[idx+1:]...)
	}
	z.setState(closedState)
	a.closedNodes = append(a.closedNodes, z)
	neighbor := (*node)(nil)
	neighbor = z.getNeighbor(a, _south)
	if !z.southBlocked && neighbor != nil {
		neighbor.update(a, z, basicCost)
	}
	neighbor = z.getNeighbor(a, _west)
	if !z.westBlocked && neighbor != nil {
		neighbor.update(a, z, basicCost)
	}
	neighbor = z.getNeighbor(a, _north)
	if !z.northBlocked && neighbor != nil {
		neighbor.update(a, z, basicCost)
	}
	neighbor = z.getNeighbor(a, _east)
	if !z.eastBlocked && neighbor != nil {
		neighbor.update(a, z, basicCost)
	}
	neighbor = z.getNeighbor(a, _southWest)
	if !(z.southBlocked || z.westBlocked) && !diagBlocked(a, z, _southWest) && neighbor != nil {
		neighbor.update(a, z, diagCost)
	}
	neighbor = z.getNeighbor(a, _northWest)
	if !(z.northBlocked || z.westBlocked) && !diagBlocked(a, z, _northWest) && neighbor != nil {
		neighbor.update(a, z, diagCost)
	}
	neighbor = z.getNeighbor(a, _northEast)
	if !(z.northBlocked || z.eastBlocked) && !diagBlocked(a, z, _northEast) && neighbor != nil {
		neighbor.update(a, z, diagCost)
	}
	neighbor = z.getNeighbor(a, _southEast)
	if !(z.southBlocked || z.eastBlocked) && !diagBlocked(a, z, _southEast) && neighbor != nil {
		neighbor.update(a, z, diagCost)
	}
}

func diagBlocked(a *astarPathFinder, n *node, dir adjacentDirection) bool {
	neighbor1 := (*node)(nil)
	neighbor2 := (*node)(nil)
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

func (z *node) setState(state nodeState) {
	z.state = state
}

func (z *node) update(a *astarPathFinder, node *node, cost int) {
	if z.state == initState {
		z.setState(openState)
		z.fCost = node.fCost + cost
		z.calcGCost(a)
		a.openNodes = append(a.openNodes, z)
	} else if z.state == closedState {
		return
	} else {
		newFcost := node.fCost + cost
		if newFcost > z.fCost {
			return
		}
		z.fCost = newFcost
	}
	z.calcHCost()
	z.parent = node.position
}

type path struct {
	waypoints *list.List
	startX    int
	startY    int
}

func newPath(startX, startY int) (z *path) {
	z = &path{
		startX: startX,
		startY: startY,
	}
	z.waypoints = list.New()
	return
}

func (z *path) addDirect(x int, y int) {
	z.waypoints.PushFront(&point{x, y})
}
