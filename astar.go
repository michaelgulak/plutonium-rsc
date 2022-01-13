package main

import (
	"container/list"
	"math"
)

var (
	wallNorth  = 1 << 0
	wallEast   = 1 << 1
	wallSouth  = 1 << 2
	wallWest   = 1 << 3
	fullBlockA = 1 << 4
	fullBlockB = 1 << 5
	fullBlockC = 1 << 6
	//_object       = 1 << 7
	wallNorthEast = wallNorth | wallEast
	wallNorthWest = wallNorth | wallWest
	wallSouthEast = wallSouth | wallEast
	wallSouthWest = wallSouth | wallWest
	fullBlock     = fullBlockA | fullBlockB | fullBlockC
	westBlocked   = fullBlock | wallWest
	southBlocked  = fullBlock | wallSouth
	northBlocked  = fullBlock | wallNorth
	eastBlocked   = fullBlock | wallEast
)

type adjacent_direction int

const (
	_south adjacent_direction = iota
	_southWest
	_west
	_northWest
	_north
	_northEast
	_east
	_southEast
)

type node_state int

const (
	initState node_state = iota
	openState
	closedState
)

const basicCost = 10
const diagCost = 14

type aStarpathfinder struct {
	depth       int
	costBoard   [][]*node
	bestpath    *path
	worldStart  *point
	pointStart  *point
	pointEnd    *point
	path        *path
	opennodes   []*node
	closednodes []*node
	entity      bool
	longPath    bool
}

type point struct {
	x int
	y int
}

func (p *point) getX() int {
	return p.x
}

func (p *point) getY() int {
	return p.y
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func calcDistance(one, two *point) int {
	xdiff := abs(one.getX() - two.getX())
	ydiff := abs(one.getY() - two.getY())

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

func newAStarpathfinder(c *client, start *point, end *point, longPath bool, entity bool, depth int) (z *aStarpathfinder) {
	z = &aStarpathfinder{}
	z.entity = entity
	z.longPath = longPath
	z.costBoard = nil
	z.bestpath = nil
	z.opennodes = []*node{}
	z.closednodes = []*node{}
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
	state        node_state
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

func (z *aStarpathfinder) buildPath() *path {
	parent := z.closednodes[len(z.closednodes)-1].parent
	endNode := z.costBoard[parent.getX()][parent.getY()]
	for endNode != nil {
		worldX := z.worldStart.getX() + z.depth - endNode.position.getX()
		worldY := z.worldStart.getY() - z.depth + endNode.position.getY()
		if endNode.parent == nil {
			endNode = nil
		} else {
			z.path.addDirect(worldX, worldY)
			endNode = z.costBoard[endNode.parent.getX()][endNode.parent.getY()]
		}
	}
	return z.path
}

func (z *aStarpathfinder) generateTraversalInfo(c *client, center *point, depth int) {
	if depth < 1 {
		return
	}
	z.costBoard = make([][]*node, 2*depth+1)
	z.initBoard(z.costBoard, 2*depth+1)
	var curposx int
	var curposy int
	for x := -depth; x <= depth; x++ {
		for y := -depth; y <= depth; y++ {
			if z.longPath {
				cdx, cdy := center.getX()-x, center.getY()+y
				if !withinWorld(cdx, cdy) {
					continue
				}
				_, ok := c.world.pathWalkerMap[(cdx<<32)|cdy]
				curposx = x + depth
				curposy = y + depth
				if ok {
					if y < depth {
						z.costBoard[curposx][curposy+1].northBlocked = true
					}
					if x > -depth {
						z.costBoard[curposx-1][curposy].eastBlocked = true
					}
					if y > -depth {
						z.costBoard[curposx][curposy-1].southBlocked = true
					}
					if x < depth {
						z.costBoard[curposx+1][curposy].westBlocked = true
					}
				}
				tile := c.combinedTile(cdx, cdy)
				if tile == math.MaxInt {
					continue
				}
				if tile&(fullBlockA|fullBlockB|fullBlockC) != 0 {
					if y < depth {
						z.costBoard[curposx][curposy+1].northBlocked = true
					}
					if x > -depth {
						z.costBoard[curposx-1][curposy].eastBlocked = true
					}
					if y > -depth {
						z.costBoard[curposx][curposy-1].southBlocked = true
					}
					if x < depth {
						z.costBoard[curposx+1][curposy].westBlocked = true
					}
				} else {
					if !z.costBoard[curposx][curposy].southBlocked {
						z.costBoard[curposx][curposy].southBlocked = (tile & southBlocked) != 0
					}
					if !z.costBoard[curposx][curposy].westBlocked {
						z.costBoard[curposx][curposy].westBlocked = (tile & westBlocked) != 0
					}
					if !z.costBoard[curposx][curposy].northBlocked {
						z.costBoard[curposx][curposy].northBlocked = (tile & northBlocked) != 0
					}
					if !z.costBoard[curposx][curposy].eastBlocked {
						z.costBoard[curposx][curposy].eastBlocked = (tile & eastBlocked) != 0
					}
				}
			} else {
				tile := c.combinedTile(center.getX()-x, center.getY()+y)
				if tile == math.MaxInt {
					continue
				}
				curposx = x + depth
				curposy = y + depth
				if curposx == z.pointEnd.x && curposy == z.pointEnd.y && z.entity {
					continue
				}
				if tile&(fullBlockA|fullBlockB|fullBlockC) != 0 {
					if y < depth {
						z.costBoard[curposx][curposy+1].northBlocked = true
					}
					if x > -depth {
						z.costBoard[curposx-1][curposy].eastBlocked = true
					}
					if y > -depth {
						z.costBoard[curposx][curposy-1].southBlocked = true
					}
					if x < depth {
						z.costBoard[curposx+1][curposy].westBlocked = true
					}
				} else {
					if !z.costBoard[curposx][curposy].southBlocked {
						z.costBoard[curposx][curposy].southBlocked = (tile & southBlocked) != 0
					}
					if !z.costBoard[curposx][curposy].westBlocked {
						z.costBoard[curposx][curposy].westBlocked = (tile & westBlocked) != 0
					}
					if !z.costBoard[curposx][curposy].northBlocked {
						z.costBoard[curposx][curposy].northBlocked = (tile & northBlocked) != 0
					}
					if !z.costBoard[curposx][curposy].eastBlocked {
						z.costBoard[curposx][curposy].eastBlocked = (tile & eastBlocked) != 0
					}
				}
			}
		}
	}
}

func (z *aStarpathfinder) findNextnode() *node {
	minimum := 0xffffffff
	minnode := (*node)(nil)
	for _, node := range z.opennodes {
		if node.hCost < minimum {
			minimum = node.hCost
			minnode = node
		} else if node.hCost == minimum && node.fCost < minnode.fCost {
			minnode = node
		}
	}
	return minnode
}

func (z *aStarpathfinder) findPath() *path {
	if z.depth < 1 {
		return nil
	}
	if z.pointStart.getX() == z.pointEnd.getX() && z.pointStart.getY() == z.pointEnd.getY() {
		return nil
	}
	z.costBoard[z.depth][z.depth].selectNode(z)
	for {
		next := z.findNextnode()
		if next == nil {
			return nil
		}
		if next.position.getX() == z.pointEnd.getX() && next.position.getY() == z.pointEnd.getY() {
			z.closednodes = append(z.closednodes, next)
			return z.buildPath()
		}
		next.selectNode(z)
	}
}

func (z *aStarpathfinder) initBoard(board [][]*node, depth int) {
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

func (z *node) calcGCost(a *aStarpathfinder) {
	z.gCost = calcDistance(z.position, a.pointEnd)
}

func (z *node) calcHCost() {
	z.hCost = z.fCost + z.gCost
}

func (z *node) getNeighbor(a *aStarpathfinder, dir adjacent_direction) (n *node) {
	switch dir {
	case _south:
		if z.position.getY() < 2*a.depth {
			return a.costBoard[z.position.getX()][z.position.getY()+1]
		} else {
			return nil
		}
	case _southWest:
		if z.position.getY() < 2*a.depth && z.position.getX() > 0 {
			n := a.costBoard[z.position.getX()-1][z.position.getY()+1]
			if a.entity && n.position.x == a.pointEnd.x && n.position.y == a.pointEnd.y {
				return nil
			}
			return n
		} else {
			return nil
		}
	case _west:
		if z.position.getX() > 0 {
			return a.costBoard[z.position.getX()-1][z.position.getY()]
		} else {
			return nil
		}
	case _northWest:
		if z.position.getY() > 0 && z.position.getX() > 0 {
			n := a.costBoard[z.position.getX()-1][z.position.getY()-1]
			if a.entity && n.position.x == a.pointEnd.x && n.position.y == a.pointEnd.y {
				return nil
			}
			return n
		} else {
			return nil
		}
	case _north:
		if z.position.getY() > 0 {
			return a.costBoard[z.position.getX()][z.position.getY()-1]
		} else {
			return nil
		}
	case _northEast:
		if z.position.getY() > 0 && z.position.getX() < 2*a.depth {
			n := a.costBoard[z.position.getX()+1][z.position.getY()-1]
			if a.entity && n.position.x == a.pointEnd.x && n.position.y == a.pointEnd.y {
				return nil
			}
			return n
		} else {
			return nil
		}
	case _east:
		if z.position.getX() < 2*a.depth {
			return a.costBoard[z.position.getX()+1][z.position.getY()]
		} else {
			return nil
		}
	case _southEast:
		if z.position.getY() < 2*a.depth && z.position.getX() < 2*a.depth {
			n := a.costBoard[z.position.getX()+1][z.position.getY()+1]
			if a.entity && n.position.x == a.pointEnd.x && n.position.y == a.pointEnd.y {
				return nil
			}
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

func (z *node) selectNode(a *aStarpathfinder) {
	if z.state == openState {
		idx := findIndex(z, a.opennodes)
		a.opennodes = append(a.opennodes[:idx], a.opennodes[idx+1:]...)
	}
	z.setState(closedState)
	a.closednodes = append(a.closednodes, z)
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

func diagBlocked(a *aStarpathfinder, n *node, dir adjacent_direction) bool {
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

func (z *node) setState(state node_state) {
	z.state = state
}

func (z *node) update(a *aStarpathfinder, node *node, cost int) {
	if z.state == initState {
		z.setState(openState)
		z.fCost = node.fCost + cost
		z.calcGCost(a)
		a.opennodes = append(a.opennodes, z)
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

//const MAXIMUM_SIZE = 50

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
	/*if z.waypoints.Len() > MAXIMUM_SIZE {
		return
	}*/
	z.waypoints.PushFront(&point{x, y})
}
