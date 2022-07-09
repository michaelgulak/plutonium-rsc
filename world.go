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
	regions       map[int32]map[int32]*region
	pathWalkerMap map[int32]struct{}
}

func newWorld() (z *world) {
	z = &world{
		regions:       map[int32]map[int32]*region{},
		pathWalkerMap: map[int32]struct{}{},
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
	if _, ok := z.regions[int32(regionX)]; !ok {
		z.regions[int32(regionX)] = map[int32]*region{}
	}
	if _, ok := z.regions[int32(regionX)][int32(regionY)]; !ok {
		z.regions[int32(regionX)][int32(regionY)] = newRegion(regionX, regionY)
	}
	return z.regions[int32(regionX)][int32(regionY)]
}

func (z *world) getTile(x int, y int) int32 {
	if !withinWorld(x, y) {
		return math.MaxInt32
	}
	return z.getRegion(x, y).getTileValue(x%48, y%48)
}

type region struct {
	tiles   [][]int32
	tile    int32
	regionX int
	regionY int
}

func newRegion(regionX int, regionY int) (z *region) {
	z = &region{}
	z.regionX = regionX
	z.regionY = regionY
	z.tile = math.MaxInt32
	z.tiles = make([][]int32, 48)
	for i := 0; i < 48; i++ {
		z.tiles[i] = make([]int32, 48)
		for j := 0; j < 48; j++ {
			z.tiles[i][j] = 0
		}
	}
	return
}

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

func (z *region) getTileValue(regionX int, regionY int) int32 {
	if z.tile != math.MaxInt32 {
		return z.tile
	}
	return z.tiles[regionX][regionY]
}

func (z *world) loadSection(reader *zip.ReadCloser, sectionX int, sectionY int, height int, bigX int, bigY int) bool {
	s := (*sector)(nil)

	filename := fmt.Sprintf("h%vx%vy%v", height, sectionX, sectionY)
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
			if int(sectorTile.groundOverlay) == 250 {
				sectorTile.groundOverlay = 2
			}
			groundOverlay := sectorTile.groundOverlay
			if groundOverlay > 0 && tileDefs[groundOverlay-1].ObjectType != 0 {
				z.getRegion(bx, by).tiles[bx%48][by%48] |= 0x40
			}
			verticalWall := int(sectorTile.verticalWall)
			if verticalWall > 0 && doorDefs[verticalWall-1].Unknown == 0 && doorDefs[verticalWall-1].DoorType != 0 {
				z.getRegion(bx, by).tiles[bx%48][by%48] |= 1       // 1
				z.getRegion(bx, by-1).tiles[bx%48][(by-1)%48] |= 4 // 4
			}
			horizontalWall := int(sectorTile.horizontalWall)
			if horizontalWall > 0 && doorDefs[horizontalWall-1].Unknown == 0 && doorDefs[horizontalWall-1].DoorType != 0 {
				z.getRegion(bx, by).tiles[bx%48][by%48] |= 2
				z.getRegion(bx-1, by).tiles[(bx-1)%48][by%48] |= 8 // 8
			}
			diagonalWalls := int(sectorTile.diagonalWalls)
			if diagonalWalls > 0 && diagonalWalls < 12000 && doorDefs[diagonalWalls-1].Unknown == 0 && doorDefs[diagonalWalls-1].DoorType != 0 {
				z.getRegion(bx, by).tiles[bx%48][by%48] |= 0x20
			}
			if diagonalWalls > 12000 && diagonalWalls < 24000 && doorDefs[diagonalWalls-12001].Unknown == 0 && doorDefs[diagonalWalls-12001].DoorType != 0 {
				z.getRegion(bx, by).tiles[bx%48][by%48] |= 0x10
			}
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
		objWidth = def.Height
		objHeight = def.Width
	} else {
		objHeight = def.Height
		objWidth = def.Width
	}
	destX := o.Pos.X
	destZ := o.Pos.Y

	if def.Typ != 2 && def.Typ != 3 && def.Typ != 0 {
		for x := destX; x <= objWidth-1+destX; x++ {
			for y := destZ; y <= destZ+objHeight-1; y++ {
				w.pathWalkerMap[int32(x<<16)|int32(y)] = struct{}{}
			}
		}
	}
}

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
	if def.Typ != 1 && def.Typ != 2 {
		return
	}
	var width int
	var height int
	if dir == 0 || dir == 4 {
		width = def.Width
		height = def.Height
	} else {
		height = def.Width
		width = def.Height
	}
	for x := o.x; x < o.x+width; x++ {
		for y := o.z; y < o.z+height; y++ {
			if def.Typ == 1 {
				c.localWorld.setTileOr(x, y, fullBlockC)
			} else if dir == 0 {
				c.localWorld.setTileOr(x, y, wallEast)
				if c.world.getTile(x-1, y) != math.MaxInt32 {
					c.localWorld.setTileOr(x-1, y, wallWest)
				}
			} else if dir == 2 {
				c.localWorld.setTileOr(x, y, wallSouth)
				if c.world.getTile(x, y+1) != math.MaxInt32 {
					c.localWorld.setTileOr(x, y+1, wallNorth)
				}
			} else if dir == 4 {
				c.localWorld.setTileOr(x, y, wallWest)
				if c.world.getTile(x+1, y) != math.MaxInt32 {
					c.localWorld.setTileOr(x+1, y, wallEast)
				}
			} else if dir == 6 {
				c.localWorld.setTileOr(x, y, wallNorth)
				if c.world.getTile(x, y-1) != math.MaxInt32 {
					c.localWorld.setTileOr(x, y-1, wallSouth)
				}
			}
		}
	}
}

func (c *client) registerWallObject(o *wallObject) {
	def := doorDefs[o.id]
	if def.DoorType != 1 {
		return
	}
	dir := o.dir
	x, y := o.x, o.z
	if dir == 0 {
		c.localWorld.setTileOr(x, y, wallNorth)
		if c.world.getTile(x, y-1) != math.MaxInt32 {
			c.localWorld.setTileOr(x, y-1, wallSouth)
		}
	} else if dir == 1 {
		c.localWorld.setTileOr(x, y, wallEast)
		if c.world.getTile(x-1, y) != math.MaxInt32 {
			c.localWorld.setTileOr(x-1, y, wallWest)
		}
	} else if dir == 2 {
		c.localWorld.setTileOr(x, y, fullBlockA)
	} else if dir == 3 {
		c.localWorld.setTileOr(x, y, fullBlockB)
	}
}

func (c *client) unregisterGameObject(o *object) {
	dir := o.dir
	def := objectDefs[o.id]
	if def.Typ != 1 && def.Typ != 2 {
		return
	}
	var width int
	var height int
	if dir == 0 || dir == 4 {
		width = def.Width
		height = def.Height
	} else {
		height = def.Width
		width = def.Height
	}
	for x := o.x; x < o.x+width; x++ {
		for y := o.z; y < o.z+height; y++ {
			if def.Typ == 1 {
				c.localWorld.setTileAnd(x, y, 0xffbf)
			} else if dir == 0 {
				c.localWorld.setTileAnd(x, y, 0xfffd)
				c.localWorld.setTileAnd(x-1, y, 65535-8)
			} else if dir == 2 {
				c.localWorld.setTileAnd(x, y, 0xfffb)
				c.localWorld.setTileAnd(x, y+1, 65535-1)
			} else if dir == 4 {
				c.localWorld.setTileAnd(x, y, 0xfff7)
				c.localWorld.setTileAnd(x+1, y, 65535-2)
			} else if dir == 6 {
				c.localWorld.setTileAnd(x, y, 0xfffe)
				c.localWorld.setTileAnd(x, y-1, 65535-4)
			}
		}
	}
}

func (c *client) unregisterWallObject(o *wallObject) {
	def := doorDefs[o.id]
	if def.DoorType != 1 {
		return
	}
	dir := o.dir
	x, y := o.x, o.z
	if dir == 0 {
		c.localWorld.setTileAnd(x, y, 0xfffe)
		c.localWorld.setTileAnd(x, y-1, 65535-4)
	} else if dir == 1 {
		c.localWorld.setTileAnd(x, y, 0xfffd)
		c.localWorld.setTileAnd(x-1, y, 65535-8)
	} else if dir == 2 {
		c.localWorld.setTileAnd(x, y, 0xffef)
	} else if dir == 3 {
		c.localWorld.setTileAnd(x, y, 0xffdf)
	}
}

func (c *client) getTile(x, z int) int32 {
	return c.world.getTile(x, z)
}

func (c *client) combinedTile(x, z int) int32 {
	tile1 := c.getTile(x, z)
	if tile1 == math.MaxInt32 {
		return math.MaxInt32
	}
	tile2 := c.localWorld.getTileTraversalMask(x, z)

	return tile1 | tile2
}

func newLocalWorld() *localWorld {
	return &localWorld{
		tiles: make(map[int32]int32),
	}
}

type localWorld struct {
	tiles map[int32]int32
}

func (c *localWorld) setTileAnd(x, z int, traversalMask int32) {
	xz := (int32(x) << 16) | int32(z)
	c.tiles[xz] = c.tiles[xz] & traversalMask
}

func (c *localWorld) setTileOr(x, z int, traversalMask int32) {
	xz := (int32(x) << 16) | int32(z)
	c.tiles[xz] = c.tiles[xz] | traversalMask
}

func (c *localWorld) getTileTraversalMask(x, z int) int32 {
	xz := (int32(x) << 16) | int32(z)
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
