package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
)

const (
	maxWorldHeight = 4032 // 3776
	maxWorldWidth  = 1008 // 944

	wallNorth  = 1 << 0
	wallEast   = 1 << 1
	wallSouth  = 1 << 2
	wallWest   = 1 << 3
	fullBlockA = 1 << 4
	fullBlockB = 1 << 5
	fullBlockC = 1 << 6

	wallNorthEast    = wallNorth | wallEast
	wallNorthWest    = wallNorth | wallWest
	wallSouthEast    = wallSouth | wallEast
	wallSouthWest    = wallSouth | wallWest
	fullBlock        = fullBlockA | fullBlockB | fullBlockC
	westBlocked      = fullBlock | wallWest
	southBlocked     = fullBlock | wallSouth
	northBlocked     = fullBlock | wallNorth
	eastBlocked      = fullBlock | wallEast
	southEastBlocked = fullBlock | wallSouthEast
	southWestBlocked = fullBlock | wallSouthWest
	northEastBlocked = fullBlock | wallNorthEast
	northWestBlocked = fullBlock | wallNorthWest

	sourceSouth = 1 << 0
	sourceWest  = 1 << 1
	sourceNorth = 1 << 2
	sourceEast  = 1 << 3

	sourceNorthEast = sourceNorth | sourceEast
	sourceNorthWest = sourceNorth | sourceWest
	sourceSouthEast = sourceSouth | sourceEast
	sourceSouthWest = sourceSouth | sourceWest
)

type world struct {
	regions       map[int16]map[int16]*region
	pathWalkerMap map[int32]struct{}
}

type sector struct {
	tiles []*tile
}

type region struct {
	tiles   [][]int32
	tile    int32
	regionX int16
	regionY int16
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

type coord struct {
	X int16
	Y int16
}

type objectLoc struct {
	ID        int
	Pos       *coord
	Direction int
}

type localWorld struct {
	tiles map[int32]int32
}

func withinWorld(x, y int) bool {
	return x >= 0 && x < maxWorldWidth && y >= 0 && y < maxWorldHeight
}

func newWorld() *world {
	w := &world{
		regions:       map[int16]map[int16]*region{},
		pathWalkerMap: map[int32]struct{}{},
	}
	return w
}

func (w *world) getRegion(x int, y int) *region {
	regionX := int16(x / 48)
	regionY := int16(y / 48)

	var (
		r0 map[int16]*region
		ok bool
	)
	if r0, ok = w.regions[regionX]; !ok {
		r0 = map[int16]*region{}
		w.regions[regionX] = r0
	}

	var r *region
	if r, ok = r0[regionY]; !ok {
		r = newRegion(regionX, regionY)
		w.regions[regionX][regionY] = r
	}

	return r
}

func (w *world) getTile(x int, y int) int32 {
	if !withinWorld(x, y) {
		return math.MaxInt32
	}
	r := w.getRegion(x, y)
	if r.tile != math.MaxInt32 {
		return r.tile
	}
	return r.tiles[x%48][y%48]
}

func newRegion(regionX int16, regionY int16) (z *region) {
	z = &region{}
	z.regionX = regionX
	z.regionY = regionY
	z.tile = math.MaxInt32
	z.tiles = make([][]int32, 48)
	for i := 0; i < 48; i++ {
		z.tiles[i] = make([]int32, 48)
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

func fileToBuffer(r io.ReadCloser) *bytes.Buffer {
	defer r.Close()
	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, r)
	if err != nil {
		fmt.Println("fileToBuffer:", err)
		os.Exit(1)
	}
	return buf
}

func (w *world) loadSection(reader *zip.ReadCloser, sectionX int, sectionY int, height int, bigX int, bigY int) bool {
	filename := fmt.Sprintf("h%dx%dy%d", height, sectionX, sectionY)
	e, err := reader.Open(filename)
	if err != nil {
		fmt.Println("loadSection:", err)
		os.Exit(1)
	}
	if e == nil {
		return false
	}
	data := fileToBuffer(e)
	s := unpackSector(data)

	for y := 0; y < 48; y++ {
		for x := 0; x < 48; x++ {
			bx := bigX + x
			by := bigY + y
			if !withinWorld(bx, by) {
				continue
			}
			sectorTile := s.tiles[x*48+y]

			groundOverlay := byte(sectorTile.groundOverlay)
			if groundOverlay == 250 {
				groundOverlay = 2
			}
			if groundOverlay > 0 && tileDefs[groundOverlay-1].ObjectType != 0 {
				w.getRegion(bx, by).tiles[bx%48][by%48] |= 0x40
			}
			verticalWall := byte(sectorTile.verticalWall)
			if verticalWall > 0 && doorDefs[verticalWall-1].Unknown == 0 && doorDefs[verticalWall-1].DoorType != 0 {
				w.getRegion(bx, by).tiles[bx%48][by%48] |= 1       // 1
				w.getRegion(bx, by-1).tiles[bx%48][(by-1)%48] |= 4 // 4
			}
			horizontalWall := byte(sectorTile.horizontalWall)
			if horizontalWall > 0 && doorDefs[horizontalWall-1].Unknown == 0 && doorDefs[horizontalWall-1].DoorType != 0 {
				w.getRegion(bx, by).tiles[bx%48][by%48] |= 2
				w.getRegion(bx-1, by).tiles[(bx-1)%48][by%48] |= 8 // 8
			}
			diagonalWalls := uint16(sectorTile.diagonalWalls)
			if diagonalWalls > 0 && diagonalWalls < 12000 && doorDefs[diagonalWalls-1].Unknown == 0 && doorDefs[diagonalWalls-1].DoorType != 0 {
				w.getRegion(bx, by).tiles[bx%48][by%48] |= 0x20
			}
			if diagonalWalls > 12000 && diagonalWalls < 24000 && doorDefs[diagonalWalls-12001].Unknown == 0 && doorDefs[diagonalWalls-12001].DoorType != 0 {
				w.getRegion(bx, by).tiles[bx%48][by%48] |= 0x10
			}

		}
	}
	return true
}

func (w *world) loadWorld() {
	fmt.Printf("[BOT] Loading landscape...")
	archiveFile := filepath.Join(settings.executablePath, settings.DataSettings.Directory, settings.DataSettings.Landscape)
	tileArchive, err := zip.OpenReader(archiveFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer tileArchive.Close()
	sectors := 0
	floorCount := 4
	for lvl := 0; lvl < floorCount; lvl++ {
		wildX := 2304
		wildY := 1776 - lvl*944
		for sx := 0; sx <= 1008-48; sx += 48 {
			for sy := 0; sy < 944; sy += 48 {
				x := (sx + wildX) / 48
				y := (sy + lvl*944 + wildY) / 48
				if w.loadSection(tileArchive, x, y, lvl, sx, sy+944*lvl) {
					sectors++
				}
			}
		}
	}

	for lvl := 0; lvl < 4; lvl++ {
		for sx := 0; sx < 20; sx++ {
			for sy := 0; sy < 20; sy++ {
				region := w.getRegion(sx*48, sy*48+(48*20*lvl))
				region.checkRedundantRegion()
			}
		}
	}
	objectLocStr := filepath.Join(settings.executablePath, settings.DataSettings.Directory, settings.DataSettings.SceneryLocs)
	f, err := os.Open(objectLocStr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	type objectLocs struct {
		Sceneries []*objectLoc
	}

	var sceneries objectLocs
	err = json.NewDecoder(f).Decode(&sceneries)
	if err != nil {
		fmt.Println("Error decoding object locs:", err)
		os.Exit(1)
	}

	for _, v := range sceneries.Sceneries {
		w.setUnwalkableTiles(v)
	}
	fmt.Println("complete")
}

func (w *world) setUnwalkableTiles(o *objectLoc) {
	var (
		objWidth  int16
		objHeight int16
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
				w.pathWalkerMap[(int32(x)<<16)|int32(y)] = struct{}{}
			}
		}
	}
}

func unpackSector(in *bytes.Buffer) *sector {
	length := 48 * 48
	sector := &sector{
		tiles: make([]*tile, 48*48),
	}
	for i := 0; i < length; i++ {
		sector.tiles[i] = unpackTile(in)
	}
	return sector
}

func readInt8FromBuffer(buf *bytes.Buffer) int8 {
	b, err := buf.ReadByte()
	if err != nil {
		fmt.Println("readInt8FromBuffer:", err)
		os.Exit(1)
	}
	return int8(b)
}

func readByteFromBuffer(buf *bytes.Buffer) byte {
	b, err := buf.ReadByte()
	if err != nil {
		fmt.Println("readByteFromBuffer:", err)
		os.Exit(1)
	}
	return b
}

func unpackTile(in *bytes.Buffer) *tile {
	tile := &tile{
		groundElevation: readInt8FromBuffer(in),
		groundTexture:   readInt8FromBuffer(in),
		groundOverlay:   readInt8FromBuffer(in),
		roofTexture:     readInt8FromBuffer(in),
		horizontalWall:  readInt8FromBuffer(in),
		verticalWall:    readInt8FromBuffer(in),
		diagonalWalls: int16((int(readByteFromBuffer(in)) << 24) |
			(int(readByteFromBuffer(in)) << 16) |
			(int(readByteFromBuffer(in)) << 8) |
			int(readByteFromBuffer(in))),
	}
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
		width = int(def.Width)
		height = int(def.Height)
	} else {
		height = int(def.Width)
		width = int(def.Height)
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
		width = int(def.Width)
		height = int(def.Height)
	} else {
		height = int(def.Width)
		width = int(def.Height)
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

func (c *client) combinedTile(x, z int) int32 {
	tile1 := c.world.getTile(x, z)
	if tile1 == math.MaxInt32 {
		return math.MaxInt32
	}
	tile2 := c.localWorld.getTile(x, z)

	return tile1 | tile2
}

func newLocalWorld() *localWorld {
	return &localWorld{
		tiles: make(map[int32]int32),
	}
}

func (l *localWorld) setTileAnd(x, z int, traversalMask int32) {
	xz := (int32(x) << 16) | int32(z)
	l.tiles[xz] = l.tiles[xz] & traversalMask
}

func (l *localWorld) setTileOr(x, z int, traversalMask int32) {
	xz := (int32(x) << 16) | int32(z)
	l.tiles[xz] = l.tiles[xz] | traversalMask
}

func (l *localWorld) getTile(x, z int) int32 {
	xz := (int32(x) << 16) | int32(z)
	return l.tiles[xz]
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
	c.pathX[var20] = int16(startX)
	c.pathZ[var20] = int16(startZ)
	openListWrite := var20 + 1
	openListSize := len(c.pathX)
	complete := false

	for openListRead != openListWrite {
		x = int(c.pathX[openListRead])
		z = int(c.pathZ[openListRead])
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
			c.pathX[openListWrite] = int16(x - 1)
			c.pathZ[openListWrite] = int16(z)
			c.pathFindSource[x-1][z] = sourceWest
			openListWrite = (openListWrite + 1) % openListSize
		}
		if x < 95 && c.pathFindSource[1+x][z] == 0 && c.combinedTile(baseX+1+x, baseZ+z)&eastBlocked == 0 {
			c.pathX[openListWrite] = int16(1 + x)
			c.pathZ[openListWrite] = int16(z)
			c.pathFindSource[x+1][z] = sourceEast
			openListWrite = (1 + openListWrite) % openListSize
		}
		if z > 0 && c.pathFindSource[x][z-1] == 0 && southBlocked&c.combinedTile(baseX+x, baseZ+z-1) == 0 {
			c.pathX[openListWrite] = int16(x)
			c.pathZ[openListWrite] = int16(z - 1)
			c.pathFindSource[x][z-1] = sourceSouth
			openListWrite = (openListWrite + 1) % openListSize
		}
		if z < 95 && c.pathFindSource[x][1+z] == 0 && northBlocked&c.combinedTile(baseX+x, baseZ+1+z) == 0 {
			c.pathX[openListWrite] = int16(x)
			c.pathZ[openListWrite] = int16(z + 1)
			c.pathFindSource[x][z+1] = sourceNorth
			openListWrite = (openListWrite + 1) % openListSize
		}
		if x > 0 && z > 0 && southBlocked&c.combinedTile(baseX+x, baseZ+z-1) == 0 && westBlocked&c.combinedTile(baseX+x-1, baseZ+z) == 0 && southWestBlocked&c.combinedTile(baseX+x-1, baseZ+z-1) == 0 && c.pathFindSource[x-1][z-1] == 0 {
			c.pathX[openListWrite] = int16(x - 1)
			c.pathZ[openListWrite] = int16(z - 1)
			c.pathFindSource[x-1][z-1] = int16(sourceSouthWest)
			openListWrite = (1 + openListWrite) % openListSize
		}
		if x < 95 && z > 0 && c.combinedTile(baseX+x, baseZ+z-1)&southBlocked == 0 && c.combinedTile(baseX+1+x, baseZ+z)&eastBlocked == 0 && c.combinedTile(baseX+x+1, baseZ+z-1)&southEastBlocked == 0 && c.pathFindSource[1+x][z-1] == 0 {
			c.pathX[openListWrite] = int16(1 + x)
			c.pathZ[openListWrite] = int16(z - 1)
			c.pathFindSource[x+1][z-1] = int16(sourceSouthEast)
			openListWrite = (1 + openListWrite) % openListSize
		}
		if x > 0 && z < 95 && c.combinedTile(baseX+x, baseZ+1+z)&northBlocked == 0 && c.combinedTile(baseX+x-1, baseZ+z)&westBlocked == 0 && c.combinedTile(baseX+x-1, baseZ+1+z)&northWestBlocked == 0 && c.pathFindSource[x-1][1+z] == 0 {
			c.pathX[openListWrite] = int16(x - 1)
			c.pathZ[openListWrite] = int16(1 + z)
			openListWrite = (1 + openListWrite) % openListSize
			c.pathFindSource[x-1][z+1] = int16(sourceNorthWest)
		}
		if x < 95 && z < 95 && northBlocked&c.combinedTile(baseX+x, baseZ+1+z) == 0 && c.combinedTile(baseX+x+1, baseZ+z)&eastBlocked == 0 && northEastBlocked&c.combinedTile(baseX+x+1, baseZ+1+z) == 0 && c.pathFindSource[x+1][1+z] == 0 {
			c.pathX[openListWrite] = int16(1 + x)
			c.pathZ[openListWrite] = int16(1 + z)
			c.pathFindSource[1+x][1+z] = int16(sourceNorthEast)
			openListWrite = (openListWrite + 1) % openListSize
		}
	}
	if !complete {
		return -1
	} else {
		c.pathX[0] = int16(x)
		c.pathZ[0] = int16(z)
		openListRead = 1
		prevSource := c.pathFindSource[x][z]
		source := prevSource
		for x != startX || z != startZ {
			if prevSource != source {
				prevSource = source
				c.pathX[openListRead] = int16(x)
				c.pathZ[openListRead] = int16(z)
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
