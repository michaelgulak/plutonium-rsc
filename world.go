package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
)

var (
	//allowedWallIdTypes = []int{5, 6, 14, 42, 63, 128, 229, 230}
	maxWorldHeight = 4032 // 3776
	maxWorldWidth  = 1008 // 944
)

const (
	sourceEast  = 8
	sourceNorth = 4
	sourceSouth = 1
	sourceWest  = 2
)

var (
	sourceNorthEast = sourceNorth | sourceEast
	sourceNorthWest = sourceNorth | sourceWest
	sourceSouthEast = sourceSouth | sourceEast
	sourceSouthWest = sourceSouth | sourceWest

	southEastBlocked = fullBlock | wallSouthEast
	southWestBlocked = fullBlock | wallSouthWest
	northEastBlocked = fullBlock | wallNorthEast
	northWestBlocked = fullBlock | wallNorthWest
)

func withinWorld(x, y int) bool {
	return x >= 0 && x < maxWorldWidth && y >= 0 && y < maxWorldHeight
}

type world struct {
	regions       map[int]map[int]*region
	pathWalkerMap map[int]struct{}
}

func newWorld() (z *world) {
	z = &world{
		regions:       map[int]map[int]*region{},
		pathWalkerMap: map[int]struct{}{},
	}
	return
}

func fileToBuffer(r io.ReadCloser) *bytes.Buffer {
	defer r.Close()
	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, r)
	if err != nil {
		panic(err)
	}
	return buf
}

func (z *world) getRegion(x int, y int) *region {
	regionX := x / 48
	regionY := y / 48
	return z.getRegionFromSectorCoordinates(regionX, regionY)
}

func (z *world) getRegionFromSectorCoordinates(regionX int, regionY int) *region {
	if _, ok := z.regions[regionX]; !ok {
		z.regions[regionX] = map[int]*region{}
	}
	if _, ok := z.regions[regionX][regionY]; !ok {
		z.regions[regionX][regionY] = newRegion(regionX, regionY)
	}
	return z.regions[regionX][regionY]
}

func (z *world) getTile(x int, y int) int {
	if !withinWorld(x, y) {
		return math.MaxInt
	}
	return z.getRegion(x, y).getTileValue(x%48, y%48)
}

type region struct {
	tiles   [][]int
	tile    int
	regionX int
	regionY int
}

func newRegion(regionX int, regionY int) (z *region) {
	z = &region{}
	z.regionX = regionX
	z.regionY = regionY
	z.tile = math.MaxInt
	z.tiles = make([][]int, 48)
	for i := 0; i < 48; i++ {
		z.tiles[i] = make([]int, 48)
		for j := 0; j < 48; j++ {
			z.tiles[i][j] = 0
		}
	}
	return
}

/*func NewTileValue() (z *tileValue) {
	z = &tileValue{}
	z.traversalMask = FULL_BLOCK
	z.diagWallVal = 0
	z.horizontalWallVal = 0
	z.overlay = 0
	z.verticalWallVal = 0
	z.elevation = 0
	z.projectileAllowed = false
	return
}*/

func (r *region) checkRedundantRegion() {
	allTilesEqual := true
	firstTile := r.tiles[0][0]
	for i := 0; i < 48 && allTilesEqual; i++ {
		for j := 0; j < 48 && allTilesEqual; j++ {
			allTilesEqual = allTilesEqual && firstTile == r.tiles[i][j]
		}
	}

	if allTilesEqual {
		r.tile = r.tiles[0][0]
		r.tiles = nil
	}
}

/*func (t *tileValue) equals(t2 *tileValue) bool {
	return t.traversalMask == t2.traversalMask &&
		t.diagWallVal == t2.diagWallVal &&
		t.horizontalWallVal == t2.horizontalWallVal &&
		t.overlay == t2.overlay &&
		t.verticalWallVal == t2.verticalWallVal &&
		t.elevation == t2.elevation &&
		t.projectileAllowed == t2.projectileAllowed
}*/

func (z *region) getTileValue(regionX int, regionY int) int {
	if z.tile != math.MaxInt {
		return z.tile
	}
	return z.tiles[regionX][regionY]
}

func (z *world) loadSection(reader *zip.ReadCloser, sectionX int, sectionY int, height int, bigX int, bigY int) bool {
	s := (*sector)(nil)

	filename := fmt.Sprintf("%v%v%v%v%v%v", "h", height, "x", sectionX, "y", sectionY)
	e, err := reader.Open(filename)
	if err != nil {
		panic(err)
	}
	if e == nil {
		return false
	}
	data := fileToBuffer(e)
	s = unpackSector(data)

	for y := 0; y < 48; y++ {
		for x := 0; x < 48; x++ {
			bx := bigX + x
			by := bigY + y
			if !withinWorld(bx, by) {
				continue
			}
			sectorTile := s.getTile2(x, y)
			//tile := z.getTile(bx, by)
			if int(sectorTile.groundOverlay) == 250 {
				sectorTile.groundOverlay = 2
			}
			groundOverlay := sectorTile.groundOverlay
			if groundOverlay > 0 && tileDefs[groundOverlay-1].objectType != 0 {
				z.getRegion(bx, by).tiles[bx%48][by%48] |= 0x40
			}
			verticalWall := int(sectorTile.verticalWall)
			if verticalWall > 0 && doorDefs[verticalWall-1].unknown == 0 && doorDefs[verticalWall-1].doorType != 0 {
				z.getRegion(bx, by).tiles[bx%48][by%48] |= 1       // 1
				z.getRegion(bx, by-1).tiles[bx%48][(by-1)%48] |= 4 // 4
				/*if projectileClipAllowed(verticalWall) {
					tile.projectileAllowed = true
					z.getTile(bx, by-1).projectileAllowed = true
				}*/
			}
			horizontalWall := int(sectorTile.horizontalWall)
			if horizontalWall > 0 && doorDefs[horizontalWall-1].unknown == 0 && doorDefs[horizontalWall-1].doorType != 0 {
				z.getRegion(bx, by).tiles[bx%48][by%48] |= 2
				z.getRegion(bx-1, by).tiles[(bx-1)%48][by%48] |= 8 // 8
				/*if projectileClipAllowed(horizontalWall) {
					tile.projectileAllowed = true
					z.getTile(bx-1, by).projectileAllowed = true
				}*/
			}
			diagonalWalls := int(sectorTile.diagonalWalls)
			if diagonalWalls > 0 && diagonalWalls < 12000 && doorDefs[diagonalWalls-1].unknown == 0 && doorDefs[diagonalWalls-1].doorType != 0 {
				z.getRegion(bx, by).tiles[bx%48][by%48] |= 0x20
				/*if projectileClipAllowed(diagonalWalls) {
					tile.projectileAllowed = true
				}*/
			}
			if diagonalWalls > 12000 && diagonalWalls < 24000 && doorDefs[diagonalWalls-12001].unknown == 0 && doorDefs[diagonalWalls-12001].doorType != 0 {
				z.getRegion(bx, by).tiles[bx%48][by%48] |= 0x10
				/*if projectileClipAllowed(diagonalWalls) {
					tile.projectileAllowed = true
				}*/
			}
			/*if tile.overlay == 2 || tile.overlay == 11 {
				tile.projectileAllowed = true
			}*/
		}
	}
	return true
}

type coord struct {
	X int
	Y int
}

type objectLoc struct {
	ID        int
	Pos       *coord
	Direction int
}

type objectLocs struct {
	Sceneries []*objectLoc
}

func (z *world) loadWorld() {
	archiveFile := settings.DataSettings.Directory + string(os.PathSeparator) + settings.DataSettings.Landscape
	tileArchive, err := zip.OpenReader(archiveFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer tileArchive.Close()
	sectors := 0
	FloorCount := 4
	for lvl := 0; lvl < FloorCount; lvl++ {
		wildX := 2304
		wildY := 1776 - lvl*944
		for sx := 0; sx <= 1008-48; sx += 48 {
			for sy := 0; sy < 944; sy += 48 {
				x := (sx + wildX) / 48
				y := (sy + lvl*944 + wildY) / 48
				if z.loadSection(tileArchive, x, y, lvl, sx, sy+944*lvl) {
					sectors++
				}
			}
		}
	}

	for lvl := 0; lvl < 4; lvl++ {
		for sx := 0; sx < 20; sx++ {
			for sy := 0; sy < 20; sy++ {
				region := z.getRegion(sx*48, sy*48+(48*20*lvl))
				region.checkRedundantRegion()
			}
		}
	}
	objectLocStr := settings.DataSettings.Directory + string(os.PathSeparator) + settings.DataSettings.SceneryLocs
	f, err := os.Open(objectLocStr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	var sceneries objectLocs
	err = json.NewDecoder(f).Decode(&sceneries)
	if err != nil {
		fmt.Printf("Error decoding object locs: %s\n", err)
		os.Exit(1)
	}
	for i := 0; i < len(sceneries.Sceneries); i++ {
		obj := sceneries.Sceneries[i]
		z.setUnwalkableTiles(obj)
	}
}

func (w *world) setUnwalkableTiles(o *objectLoc) {
	var (
		objWidth  int
		objHeight int
	)
	def := objectDefs[o.ID]
	if o.Direction != 0 && o.Direction != 4 {
		objWidth = def.height
		objHeight = def.width
	} else {
		objHeight = def.height
		objWidth = def.width
	}
	destX := o.Pos.X
	destZ := o.Pos.Y

	if def.typ != 2 && def.typ != 3 && def.typ != 0 {
		for x := destX; x <= objWidth-1+destX; x++ {
			for y := destZ; y <= destZ+objHeight-1; y++ {
				w.pathWalkerMap[(x<<32)|y] = struct{}{}
			}
		}
	} /* else {
		if o.Direction == 0 {
			objWidth++
			destZ--
		}

		if o.Direction == 2 {
			objHeight++
		}

		if o.Direction == 6 {
			destZ--
			objHeight++
		}

		if o.Direction == 4 {
			objWidth++
		}
		for x := destX; x <= objWidth+(destX-1); x++ {
			for y := destZ; y <= objHeight+destZ-1; y++ {
				w.pathWalkerMap[(x<<32)|y] = struct{}{}
			}
		}
	}*/
}

/*func projectileClipAllowed(wallID int) bool {
	for _, allowedWallIdType := range allowedWallIdTypes {
		if allowedWallIdType == wallID {
			return true
		}
	}
	return false
}*/

type sector struct {
	tiles []*tile
}

func newSector() (z *sector) {
	z = &sector{}
	z.tiles = make([]*tile, 48*48)
	for i := 0; i < len(z.tiles); i++ {
		z.tiles[i] = newTile()
	}
	return
}

func (z *sector) getTile(i int) *tile {
	return z.tiles[i]
}

func (z *sector) getTile2(x int, y int) *tile {
	return z.getTile(x*48 + y)
}

func (z *sector) setTile(i int, t *tile) {
	z.tiles[i] = t
}

func (z *sector) SetTile2(x int, y int, t *tile) {
	z.setTile(x*48+y, t)
}

func unpackSector(in *bytes.Buffer) *sector {
	length := 48 * 48
	sector := newSector()
	for i := 0; i < length; i++ {
		sector.setTile(i, unpackTile(in))
	}
	return sector
}

type tile struct {
	diagonalWalls   int16
	groundElevation int8
	groundOverlay   int8
	groundTexture   int8
	horizontalWall  int8
	roofTexture     int8
	verticalWall    int8
}

func newTile() (z *tile) {
	z = &tile{}
	z.diagonalWalls = 0
	z.groundElevation = 0
	z.groundOverlay = 0
	z.groundTexture = 0
	z.horizontalWall = 0
	z.roofTexture = 0
	z.verticalWall = 0
	return
}

func readByteFromBuffer(buf *bytes.Buffer) int8 {
	b, err := buf.ReadByte()
	if err != nil {
		panic(err)
	}
	return int8(b)
}

func readByte0FromBuffer(buf *bytes.Buffer) byte {
	b, err := buf.ReadByte()
	if err != nil {
		panic(err)
	}
	return b
}

func unpackTile(in *bytes.Buffer) *tile {
	tile := newTile()
	tile.groundElevation = readByteFromBuffer(in)
	tile.groundTexture = readByteFromBuffer(in)
	tile.groundOverlay = readByteFromBuffer(in)
	tile.roofTexture = readByteFromBuffer(in)
	tile.horizontalWall = readByteFromBuffer(in)
	tile.verticalWall = readByteFromBuffer(in)
	tile.diagonalWalls = int16((int(readByte0FromBuffer(in)) << 24) | (int(readByte0FromBuffer(in)) << 16) | (int(readByte0FromBuffer(in)) << 8) | int(readByte0FromBuffer(in)))
	return tile
}

func (c *client) registerGameObject(o *object) {
	dir := o.dir
	if o.id == 1147 {
		return
	}
	def := objectDefs[o.id]
	if def.typ != 1 && def.typ != 2 {
		return
	}
	var width int
	var height int
	if dir == 0 || dir == 4 {
		width = def.width
		height = def.height
	} else {
		height = def.width
		width = def.height
	}
	for x := o.x; x < o.x+width; x++ {
		for y := o.z; y < o.z+height; y++ {
			if def.typ == 1 {
				//z.getTile(x, y).traversalMask |= FULL_BLOCK_C
				c.localWorld.setTileOr(x, y, fullBlockC)
			} else if dir == 0 {
				//z.getTile(x, y).traversalMask |= WALL_EAST
				c.localWorld.setTileOr(x, y, wallEast)
				if c.world.getTile(x-1, y) != math.MaxInt {
					//z.getTile(x-1, y).traversalMask |= WALL_WEST
					c.localWorld.setTileOr(x-1, y, wallWest)
				}
			} else if dir == 2 {
				//z.getTile(x, y).traversalMask |= WALL_SOUTH
				c.localWorld.setTileOr(x, y, wallSouth)
				if c.world.getTile(x, y+1) != math.MaxInt {
					//z.getTile(x, y+1).traversalMask |= WALL_NORTH
					c.localWorld.setTileOr(x, y+1, wallNorth)
				}
			} else if dir == 4 {
				//z.getTile(x, y).traversalMask |= WALL_WEST
				c.localWorld.setTileOr(x, y, wallWest)
				if c.world.getTile(x+1, y) != math.MaxInt {
					//z.getTile(x+1, y).traversalMask |= WALL_EAST
					c.localWorld.setTileOr(x+1, y, wallEast)
				}
			} else if dir == 6 {
				//z.getTile(x, y).traversalMask |= WALL_NORTH
				c.localWorld.setTileOr(x, y, wallNorth)
				if c.world.getTile(x, y-1) != math.MaxInt {
					//z.getTile(x, y-1).traversalMask |= WALL_SOUTH
					c.localWorld.setTileOr(x, y-1, wallSouth)
				}
			}
		}
	}
}

func (c *client) registerWallObject(o *wallObject) {
	def := doorDefs[o.id]
	if def.doorType != 1 {
		return
	}
	dir := o.dir
	x, y := o.x, o.z
	if dir == 0 {
		//w.getTile(x, y).traversalMask |= WALL_NORTH
		c.localWorld.setTileOr(x, y, wallNorth)
		if c.world.getTile(x, y-1) != math.MaxInt {
			//w.getTile(x, y-1).traversalMask |= WALL_SOUTH
			c.localWorld.setTileOr(x, y-1, wallSouth)
		}
	} else if dir == 1 {
		//w.getTile(x, y).traversalMask |= WALL_EAST
		c.localWorld.setTileOr(x, y, wallEast)
		if c.world.getTile(x-1, y) != math.MaxInt {
			//w.getTile(x-1, y).traversalMask |= WALL_WEST
			c.localWorld.setTileOr(x-1, y, wallWest)
		}
	} else if dir == 2 {
		//w.getTile(x, y).traversalMask |= FULL_BLOCK_A
		c.localWorld.setTileOr(x, y, fullBlockA)
	} else if dir == 3 {
		//w.getTile(x, y).traversalMask |= FULL_BLOCK_B
		c.localWorld.setTileOr(x, y, fullBlockB)
	}
}

func (c *client) unregisterGameObject(o *object) {
	dir := o.dir
	def := objectDefs[o.id]
	if def.typ != 1 && def.typ != 2 {
		return
	}
	var width int
	var height int
	if dir == 0 || dir == 4 {
		width = def.width
		height = def.height
	} else {
		height = def.width
		width = def.height
	}
	for x := o.x; x < o.x+width; x++ {
		for y := o.z; y < o.z+height; y++ {
			if def.typ == 1 {
				//z.getTile(x, y).traversalMask &= 0xffbf
				c.localWorld.setTileAnd(x, y, 0xffbf)
			} else if dir == 0 {
				//z.getTile(x, y).traversalMask &= 0xfffd
				//z.getTile(x-1, y).traversalMask &= 65535 - 8
				c.localWorld.setTileAnd(x, y, 0xfffd)
				c.localWorld.setTileAnd(x-1, y, 65535-8)
			} else if dir == 2 {
				//z.getTile(x, y).traversalMask &= 0xfffb
				//z.getTile(x, y+1).traversalMask &= 65535 - 1
				c.localWorld.setTileAnd(x, y, 0xfffb)
				c.localWorld.setTileAnd(x, y+1, 65535-1)
			} else if dir == 4 {
				//z.getTile(x, y).traversalMask &= 0xfff7
				//z.getTile(x+1, y).traversalMask &= 65535 - 2
				c.localWorld.setTileAnd(x, y, 0xfff7)
				c.localWorld.setTileAnd(x+1, y, 65535-2)
			} else if dir == 6 {
				//z.getTile(x, y).traversalMask &= 0xfffe
				//z.getTile(x, y-1).traversalMask &= 65535 - 4
				c.localWorld.setTileAnd(x, y, 0xfffe)
				c.localWorld.setTileAnd(x, y-1, 65535-4)
			}
		}
	}
}

func (c *client) unregisterWallObject(o *wallObject) {
	def := doorDefs[o.id]
	if def.doorType != 1 {
		return
	}
	dir := o.dir
	x, y := o.x, o.z
	if dir == 0 {
		//w.getTile(x, y).traversalMask &= 0xfffe
		c.localWorld.setTileAnd(x, y, 0xfffe)
		//w.getTile(x, y-1).traversalMask &= 65535 - 4
		c.localWorld.setTileAnd(x, y-1, 65535-4)
	} else if dir == 1 {
		//w.getTile(x, y).traversalMask &= 0xfffd
		c.localWorld.setTileAnd(x, y, 0xfffd)
		//w.getTile(x-1, y).traversalMask &= 65535 - 8
		c.localWorld.setTileAnd(x-1, y, 65535-8)
	} else if dir == 2 {
		//w.getTile(x, y).traversalMask &= 0xffef
		c.localWorld.setTileAnd(x, y, 0xffef)
	} else if dir == 3 {
		c.localWorld.setTileAnd(x, y, 0xffdf)
		//w.getTile(x, y).traversalMask &= 0xffdf
	}
}

func (c *client) getTile(x, z int) int {
	return c.world.getTile(x, z)
}

func (c *client) combinedTile(x, z int) int {
	tile1 := c.getTile(x, z)
	if tile1 == math.MaxInt {
		return math.MaxInt
	}
	tile2 := c.localWorld.getTileTraversalMask(x, z)

	return tile1 | tile2
}

func newLocalWorld() *localWorld {
	return &localWorld{
		tiles: make(map[int64]int),
	}
}

type localWorld struct {
	tiles map[int64]int
}

func (c *localWorld) setTileAnd(x, z, traversalMask int) {
	xz := (int64(x) << 32) | int64(z)
	c.tiles[xz] = c.tiles[xz] & traversalMask
}

func (c *localWorld) setTileOr(x, z, traversalMask int) {
	xz := (int64(x) << 32) | int64(z)
	c.tiles[xz] = c.tiles[xz] | traversalMask
}

func (c *localWorld) getTileTraversalMask(x, z int) int {
	xz := (int64(x) << 32) | int64(z)
	if tile, ok := c.tiles[xz]; ok {
		return tile
	}
	c.tiles[xz] = 0
	return c.tiles[xz]
}

func (c *client) findPath(startX int, startZ int, xLow int, xHigh int, zLow int, zHigh int, reachBorder bool) int {
	baseX := c.regionX
	baseZ := c.regionZ

	startX = startX - c.regionX
	startZ = startZ - c.regionZ

	xLow = xLow - c.regionX
	xHigh = xHigh - c.regionX
	zLow = zLow - c.regionZ
	zHigh = zHigh - c.regionZ

	for x := 0; x < 96; x++ {
		for y := 0; y < 96; y++ {
			c.pathFindSource[x][y] = 0
		}
	}
	var20 := 0
	openListRead := 0
	x := startX
	z := startZ
	c.pathFindSource[startX][startZ] = 99
	c.pathX[var20] = startX
	c.pathZ[var20] = startZ
	openListWrite := var20 + 1
	openListSize := len(c.pathX)
	complete := false
	for openListRead != openListWrite {
		x = c.pathX[openListRead]
		z = c.pathZ[openListRead]
		openListRead = (1 + openListRead) % openListSize
		if x >= xLow && x <= xHigh && z >= zLow && z <= zHigh {
			complete = true
			break
		}
		if reachBorder {
			if x > 0 && xLow <= x-1 && xHigh >= x-1 && zLow <= z && zHigh >= z && c.combinedTile(baseX+x-1, baseZ+z)&wallWest == 0 {
				complete = true
				break
			}
			if x < 95 && 1+x >= xLow && x+1 <= xHigh && z >= zLow && zHigh >= z && wallEast&c.combinedTile(baseX+x+1, baseZ+z) == 0 {
				complete = true
				break
			}
			if z > 0 && xLow <= x && xHigh >= x && z-1 >= zLow && zHigh >= z-1 && wallSouth&c.combinedTile(baseX+x, baseZ+z-1) == 0 {
				complete = true
				break
			}
			if z < 95 && xLow <= x && x <= xHigh && zLow <= z+1 && zHigh >= z+1 && wallNorth&c.combinedTile(baseX+x, baseZ+z+1) == 0 {
				complete = true
				break
			}
		}
		if x > 0 && c.pathFindSource[x-1][z] == 0 && c.combinedTile(baseX+x-1, baseZ+z)&westBlocked == 0 {
			c.pathX[openListWrite] = x - 1
			c.pathZ[openListWrite] = z
			c.pathFindSource[x-1][z] = sourceWest
			openListWrite = (openListWrite + 1) % openListSize
		}
		if x < 95 && c.pathFindSource[1+x][z] == 0 && c.combinedTile(baseX+1+x, baseZ+z)&eastBlocked == 0 {
			c.pathX[openListWrite] = 1 + x
			c.pathZ[openListWrite] = z
			c.pathFindSource[x+1][z] = sourceEast
			openListWrite = (1 + openListWrite) % openListSize
		}
		if z > 0 && c.pathFindSource[x][z-1] == 0 && southBlocked&c.combinedTile(baseX+x, baseZ+z-1) == 0 {
			c.pathX[openListWrite] = x
			c.pathZ[openListWrite] = z - 1
			c.pathFindSource[x][z-1] = sourceSouth
			openListWrite = (openListWrite + 1) % openListSize
		}
		if z < 95 && c.pathFindSource[x][1+z] == 0 && northBlocked&c.combinedTile(baseX+x, baseZ+1+z) == 0 {
			c.pathX[openListWrite] = x
			c.pathZ[openListWrite] = z + 1
			c.pathFindSource[x][z+1] = sourceNorth
			openListWrite = (openListWrite + 1) % openListSize
		}
		if x > 0 && z > 0 && southBlocked&c.combinedTile(baseX+x, baseZ+z-1) == 0 && westBlocked&c.combinedTile(baseX+x-1, baseZ+z) == 0 && southWestBlocked&c.combinedTile(baseX+x-1, baseZ+z-1) == 0 && c.pathFindSource[x-1][z-1] == 0 {
			c.pathX[openListWrite] = x - 1
			c.pathZ[openListWrite] = z - 1
			c.pathFindSource[x-1][z-1] = sourceSouthWest
			openListWrite = (1 + openListWrite) % openListSize
		}
		if x < 95 && z > 0 && c.combinedTile(baseX+x, baseZ+z-1)&southBlocked == 0 && c.combinedTile(baseX+1+x, baseZ+z)&eastBlocked == 0 && c.combinedTile(baseX+x+1, baseZ+z-1)&southEastBlocked == 0 && c.pathFindSource[1+x][z-1] == 0 {
			c.pathX[openListWrite] = 1 + x
			c.pathZ[openListWrite] = z - 1
			c.pathFindSource[x+1][z-1] = sourceSouthEast
			openListWrite = (1 + openListWrite) % openListSize
		}
		if x > 0 && z < 95 && c.combinedTile(baseX+x, baseZ+1+z)&northBlocked == 0 && c.combinedTile(baseX+x-1, baseZ+z)&westBlocked == 0 && c.combinedTile(baseX+x-1, baseZ+1+z)&northWestBlocked == 0 && c.pathFindSource[x-1][1+z] == 0 {
			c.pathX[openListWrite] = x - 1
			c.pathZ[openListWrite] = 1 + z
			openListWrite = (1 + openListWrite) % openListSize
			c.pathFindSource[x-1][z+1] = sourceNorthWest
		}
		if x < 95 && z < 95 && northBlocked&c.combinedTile(baseX+x, baseZ+1+z) == 0 && c.combinedTile(baseX+x+1, baseX+z)&eastBlocked == 0 && northEastBlocked&c.combinedTile(baseX+x+1, baseZ+1+z) == 0 && c.pathFindSource[x+1][1+z] == 0 {
			c.pathX[openListWrite] = 1 + x
			c.pathZ[openListWrite] = 1 + z
			c.pathFindSource[1+x][1+z] = sourceNorthEast
			openListWrite = (openListWrite + 1) % openListSize
		}
	}
	if !complete {
		return -1
	} else {
		c.pathX[0] = x
		c.pathZ[0] = z
		openListRead = 1
		var prevSource int
		prevSource = c.pathFindSource[x][z]
		source := prevSource
		for x != startX || z != startZ {
			if prevSource != source {
				prevSource = source
				c.pathX[openListRead] = x
				c.pathZ[openListRead] = z
				openListRead++
			}
			if source&sourceSouth != 0 {
				z++
			} else if sourceNorth&source != 0 {
				z--
			}
			if sourceWest&source != 0 {
				x++
			} else if source&sourceEast != 0 {
				x--
			}
			source = c.pathFindSource[x][z]
		}
		return openListRead
	}
}
