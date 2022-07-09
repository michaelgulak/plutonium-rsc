package main

import (
	"math"
	"strings"
	"time"
)

var (
	bankersID     = []int{95, 224, 268, 485, 540, 617}
	sleepingBagID = 1263
)

func (c *client) getInventoryCountByID(ids ...int) int {
	count := 0
	for i := 0; i < c.inventoryCount; i++ {
		if inIntArray(c.inventory[i].id, ids) {
			count += c.inventory[i].amount
		}
	}
	return count
}

func (c *client) getInventoryItemByID(ids ...int) *item {
	for i := 0; i < c.inventoryCount; i++ {
		if inIntArray(c.inventory[i].id, ids) {
			return c.inventory[i]
		}
	}
	return nil
}

func (c *client) isItemEquippedByID(id int) bool {
	for i := 0; i < c.inventoryCount; i++ {
		if c.inventory[i].id == id && c.inventory[i].equipped {
			return true
		}
	}
	return false
}

func (c *client) useSleepingBag() {
	bag := c.getInventoryItemByID(sleepingBagID)
	if bag != nil {
		c.useInventoryItem(bag)
	}
}

func (c *player) isSkilling() bool {
	return time.Since(c.skillTime) < c.skillingTimeout
}

func (c *quest) isComplete() bool {
	return c.stage < 0
}

func (c *client) getHPPercent() int {
	return int((float64(c.currentStats[3]) / float64(c.baseStats[3])) * 100)
}

func distance(x1, z1, x2, z2 int) float64 {
	dx := float64(x1 - x2)
	dy := float64(z1 - z2)
	return math.Sqrt(dx*dx + dy*dy)
}

func (c *client) distanceTo(x, z int) float64 {
	return distance(x, z, c.x, c.z)
}

func (c *client) inRadiusOf(x, z int, radius float64) bool {
	return c.distanceTo(x, z) <= radius
}

func (c *client) getObjectFromCoords(x, z int) *object {
	for i := 0; i < c.objectCount; i++ {
		if c.objects[i].x == x && c.objects[i].z == z {
			return c.objects[i]
		}
	}

	return nil
}

func (c *client) getWallObjectFromCoords(x, z int) *wallObject {
	for i := 0; i < c.wallObjectCount; i++ {
		if c.wallObjects[i].x == x && c.wallObjects[i].z == z {
			return c.wallObjects[i]
		}
	}

	return nil
}

func (c *client) isReachable(x, z int) bool {
	if c.x == x && c.z == z {
		return true
	}
	l := c.findPath(c.x, c.z, x, x, z, z, false)
	return l != -1
}

func (c *client) isGroundItemReachable(g *groundItem) bool {
	l := c.findPath(c.x, c.z, g.x, g.x, g.z, g.z, false)
	if l == -1 {
		l = c.findPath(c.x, c.z, g.x, g.x, g.z, g.z, true)
		return l != -1
	}
	return true
}

func (c *client) isWallObjectReachable(w *wallObject) bool {
	if w.dir == 0 {
		l := c.findPath(c.x, c.z, w.x, w.x, w.z-1, w.z, false)
		return l != -1
	} else if w.dir != 1 {
		l := c.findPath(c.x, c.z, w.x, w.x, w.z, w.z, true)
		return l != -1
	} else {
		l := c.findPath(c.x, c.z, w.x-1, w.x, w.z, w.z, false)
		return l != -1
	}
}

func (c *client) isObjectReachable(o *object) bool {
	var (
		objWidth  int
		objHeight int
	)
	def := objectDefs[o.id]
	if o.dir != 0 && o.dir != 4 {
		objWidth = def.Height
		objHeight = def.Width
	} else {
		objHeight = def.Height
		objWidth = def.Width
	}
	destX := o.x
	destZ := o.z

	if def.Typ != 2 && def.Typ != 3 {
		l := c.findPath(c.x, c.z, o.x, objWidth-1+destX, o.z, destZ+objHeight-1, true)
		return l != -1
	} else {
		if o.dir == 0 {
			objWidth++
			destZ--
		}

		if o.dir == 2 {
			objHeight++
		}

		if o.dir == 6 {
			destZ--
			objHeight++
		}

		if o.dir == 4 {
			objWidth++
		}

		l := c.findPath(c.x, c.z, destX, objWidth+(destX-1), destZ, objHeight+destZ-1, false)
		return l != -1
	}
}

func (n *npc) inCombat() bool {
	return n.sprite == 8 || n.sprite == 9
}

func (c *client) inCombat() bool {
	return c.sprite == 8 || c.sprite == 9
}

func (n *npc) isTalking() bool {
	return time.Since(n.messageTime) < n.lastMessageTimeout
}

func (n *player) isTalking() bool {
	return time.Since(n.messageTime) < n.lastMessageTimeout
}

func (p *player) inCombat() bool {
	return p.sprite == 8 || p.sprite == 9
}

func (c *client) inRect(x, z, width, height int) bool {
	return c.x <= x && c.z >= z && c.x >= x-width && c.z <= z+height
}

func inRect(x1, z1, x2, z2, width, height int) bool {
	return x1 <= x2 && z1 >= z2 && x1 >= x2-width && z1 <= z2+height
}

func inIntArray(needle int, haystack []int) bool {
	for _, n := range haystack {
		if needle == n {
			return true
		}
	}
	return false
}

func (c *client) getNearestNPCByID(inCombat, talking, reachable bool, x, z int, radius float64, ids ...int) *npc {
	var (
		min    = float64(0xfffffff)
		minNPC *npc
	)
	for i := 0; i < c.npcCount; i++ {
		if !inIntArray(c.npcs[i].id, ids) {
			continue
		}
		if !inCombat && c.npcs[i].inCombat() {
			continue
		}
		if !talking && c.npcs[i].isTalking() {
			continue
		}
		if radius != -1 && distance(c.npcs[i].x, c.npcs[i].z, x, z) > radius {
			continue
		}
		if reachable && !c.isReachable(c.npcs[i].x, c.npcs[i].z) {
			continue
		}
		if dist := distance(c.x, c.z, c.npcs[i].x, c.npcs[i].z); dist < min {
			minNPC = c.npcs[i]
			min = dist
		}
	}
	return minNPC
}

func (c *client) getNearestNPCByIDInRect(inCombat bool, talking bool, reachable bool, x, z, width, height int, ids ...int) *npc {
	var (
		min    = float64(0xfffffff)
		minNPC *npc
	)
	for i := 0; i < c.npcCount; i++ {
		if !inIntArray(c.npcs[i].id, ids) {
			continue
		}
		if !inCombat && c.npcs[i].inCombat() {
			continue
		}
		if !talking && c.npcs[i].isTalking() {
			continue
		}
		if !inRect(c.npcs[i].x, c.npcs[i].z, x, z, width, height) {
			continue
		}
		if reachable && !c.isReachable(c.npcs[i].x, c.npcs[i].z) {
			continue
		}
		if dist := distance(c.x, c.z, c.npcs[i].x, c.npcs[i].z); dist < min {
			minNPC = c.npcs[i]
			min = dist
		}
	}
	return minNPC
}

func (c *client) getNearestObjectByIDInRect(reachable bool, x, z, width, height int, ids ...int) *object {
	var (
		min    = float64(0xfffffff)
		minObj *object
	)
	for i := 0; i < c.objectCount; i++ {
		if !inIntArray(c.objects[i].id, ids) {
			continue
		}
		if !inRect(c.objects[i].x, c.objects[i].z, x, z, width, height) {
			continue
		}
		if reachable && !c.isObjectReachable(c.objects[i]) {
			continue
		}
		if dist := distance(c.x, c.z, c.objects[i].x, c.objects[i].z); dist < min {
			minObj = c.objects[i]
			min = dist
		}
	}
	return minObj
}

func (c *client) getNearestWallObjectByIDInRect(reachable bool, x, z, width, height int, ids ...int) *wallObject {
	var (
		min    = float64(0xfffffff)
		minObj *wallObject
	)
	for i := 0; i < c.wallObjectCount; i++ {
		if !inIntArray(c.wallObjects[i].id, ids) {
			continue
		}
		if !inRect(c.wallObjects[i].x, c.wallObjects[i].z, x, z, width, height) {
			continue
		}
		if reachable && !c.isWallObjectReachable(c.wallObjects[i]) {
			continue
		}
		if dist := distance(c.x, c.z, c.wallObjects[i].x, c.wallObjects[i].z); dist < min {
			minObj = c.wallObjects[i]
			min = dist
		}
	}
	return minObj
}

func (c *client) getNearestGroundItemByIDInRect(reachable bool, x, z, width, height int, ids ...int) *groundItem {
	var (
		min    = float64(0xfffffff)
		minObj *groundItem
	)
	for i := 0; i < c.groundItemCount; i++ {
		if !inIntArray(c.groundItems[i].id, ids) {
			continue
		}
		if !inRect(c.groundItems[i].x, c.groundItems[i].z, x, z, width, height) {
			continue
		}
		if reachable && !c.isGroundItemReachable(c.groundItems[i]) {
			continue
		}
		if dist := distance(c.x, c.z, c.groundItems[i].x, c.groundItems[i].z); dist < min {
			minObj = c.groundItems[i]
			min = dist
		}
	}
	return minObj
}

func (c *client) getNearestGroundItemByID(x, z int, radius float64, reachable bool, ids ...int) *groundItem {
	var (
		min     = float64(0xfffffff)
		minItem *groundItem
	)
	for i := 0; i < c.groundItemCount; i++ {
		if !inIntArray(c.groundItems[i].id, ids) {
			continue
		}
		if radius != -1 && distance(c.groundItems[i].x, c.groundItems[i].z, x, z) > radius {
			continue
		}
		if reachable && !c.isGroundItemReachable(c.groundItems[i]) {
			continue
		}
		if dist := distance(c.x, c.z, c.groundItems[i].x, c.groundItems[i].z); dist < min {
			minItem = c.groundItems[i]
			min = dist
		}
	}
	return minItem
}

func (c *client) getNearestObjectByID(x, z int, radius float64, reachable bool, ids ...int) *object {
	var (
		min    = float64(0xfffffff)
		minObj *object
	)
	for i := 0; i < c.objectCount; i++ {
		if !inIntArray(c.objects[i].id, ids) {
			continue
		}
		if radius != -1 && distance(c.objects[i].x, c.objects[i].z, x, z) > radius {
			continue
		}
		if reachable && !c.isObjectReachable(c.objects[i]) {
			continue
		}
		if dist := distance(c.x, c.z, c.objects[i].x, c.objects[i].z); dist < min {
			minObj = c.objects[i]
			min = dist
		}
	}
	return minObj
}

func (c *client) getNearestWallObjectByID(x, z int, radius float64, reachable bool, ids ...int) *wallObject {
	var (
		min    = float64(0xfffffff)
		minObj *wallObject
	)
	for i := 0; i < c.wallObjectCount; i++ {
		if !inIntArray(c.wallObjects[i].id, ids) {
			continue
		}
		if radius != -1 && distance(c.wallObjects[i].x, c.wallObjects[i].z, x, z) > radius {
			continue
		}
		if reachable && !c.isWallObjectReachable(c.wallObjects[i]) {
			continue
		}
		if dist := distance(c.x, c.z, c.wallObjects[i].x, c.wallObjects[i].z); dist < min {
			minObj = c.wallObjects[i]
			min = dist
		}
	}
	return minObj
}

func (c *client) bankCount(ids ...int) int {
	if c.banking {
		count := 0
		for i := 0; i < c.bankItemCount; i++ {
			if inIntArray(c.bankItems[i].id, ids) {
				count += c.bankItems[i].amount
			}
		}
		return count
	}
	return 0
}

func (c *client) hasBankItem(id int) bool {
	return c.bankCount(id) > 0
}

func (c *client) getPlayerByName(name string) *player {
	name = strings.ToLower(name)
	for i := 0; i < c.playerCount; i++ {
		if strings.ToLower(c.players[i].username) == name {
			return c.players[i]
		}
	}
	return nil
}

func (c *client) isGroundItemAt(id, x, z int) bool {
	for i := 0; i < c.groundItemCount; i++ {
		if c.groundItems[i].id == id && c.groundItems[i].x == x && c.groundItems[i].z == z {
			return true
		}
	}
	return false
}

func (c *client) hasRecipientOffer(id, amount int) bool {
	for i := 0; i < c.recipientTradeCount; i++ {
		if c.recipientTradeItems[i].id == id && c.recipientTradeItems[i].amount == amount {
			return true
		}
	}
	return false
}

func (c *client) hasRecipientConfirm(id, amount int) bool {
	for i := 0; i < c.recipientConfirmItemCount; i++ {
		if c.recipientConfirmItems[i].id == id && c.recipientConfirmItems[i].amount == amount {
			return true
		}
	}
	return false
}

func (c *client) hasTradeOffer(id, amount int) bool {
	for i := 0; i < c.myTradeCount; i++ {
		if c.myTradeItems[i].id == id && c.myTradeItems[i].amount == amount {
			return true
		}
	}
	return false
}
func (c *client) hasTradeConfirm(id, amount int) bool {
	for i := 0; i < c.myTradeConfirmItemCount; i++ {
		if c.myTradeConfirmItems[i].id == id && c.myTradeConfirmItems[i].amount == amount {
			return true
		}
	}
	return false
}

func (c *client) isObjectAt(x, z int) bool {
	for i := 0; i < c.objectCount; i++ {
		if c.objects[i].x == x && c.objects[i].z == z {
			return true
		}
	}
	return false
}

func (c *client) isWallObjectAt(x, z int) bool {
	for i := 0; i < c.wallObjectCount; i++ {
		if c.wallObjects[i].x == x && c.wallObjects[i].z == z {
			return true
		}
	}
	return false
}

func (c *client) getShopItemByID(id int) *shopItem {
	for i := 0; i < c.shopItemCount; i++ {
		if c.shopItems[i].id == id {
			return c.shopItems[i]
		}
	}
	return nil
}

func (c *client) isFriend(name string) bool {
	name = strings.ToLower(name)
	for i := 0; i < c.friendListCount; i++ {
		if strings.ToLower(c.friendList[i].username) == name {
			return true
		}
	}
	return false
}

func (c *client) isIgnored(name string) bool {
	name = strings.ToLower(name)
	for i := 0; i < c.ignoreListCount; i++ {
		if strings.ToLower(c.ignoreList[i].username) == name {
			return true
		}
	}
	return false
}

func (c *client) getShopAmount(id int) int {
	for i := 0; i < c.shopItemCount; i++ {
		if c.shopItems[i].id == id {
			return c.shopItems[i].amount
		}
	}
	return 0
}

func (c *client) generateLongPathPoint(x, z, depth int) *point {
	if depth == -1 {
		depth = int(distance(c.x, c.z, x, z) * 4)
	}

	if depth > 800 {
		return nil
	}

	p := newAstarPathFinder(c, &point{c.x, c.z}, &point{x, z}, depth)

	path := p.findPath()

	if path == nil {
		return nil
	}

	path.waypoints.PushBack(&point{x, z})

	count := 0
	coordsAwayThreshold := 10
	var pt *point
	if path.waypoints.Len() < coordsAwayThreshold {
		pt = path.waypoints.Back().Value.(*point)
	} else {
		for e := path.waypoints.Front(); e != nil; e = e.Next() {
			if count == coordsAwayThreshold-1 {
				pt = e.Value.(*point)
				break
			}
			count++
		}
	}

	return pt
}

func (c *client) generateLongPath(startX, startZ, x, z, depth int) *longPath {
	var (
		p    *astarPathFinder
		path *path
	)
	if depth == -1 {
		depth = int(math.Pow(2, math.Ceil(math.Log2(distance(startX, startZ, x, z)))))
		for {
			if depth > 512 {
				return nil
			}
			p = newAstarPathFinder(c, &point{startX, startZ}, &point{x, z}, depth)
			path = p.findPath()

			if path != nil {
				break
			}
			depth <<= 1
		}
	} else {
		p = newAstarPathFinder(c, &point{startX, startZ}, &point{x, z}, depth)
		path = p.findPath()

		if path == nil {
			return nil
		}
	}

	path.waypoints.PushBack(&point{x, z})

	count := 0
	coordsAwayThreshold := 10
	lp := &longPath{points: []*point{}, c: c}

	if path.waypoints.Len() < coordsAwayThreshold {
		lp.points = append(lp.points, path.waypoints.Back().Value.(*point))
	} else {
		for e := path.waypoints.Front(); e != nil; e = e.Next() {
			if count%coordsAwayThreshold-1 == 0 {
				lp.points = append(lp.points, e.Value.(*point))
			}
			count++
		}
		if count-1%coordsAwayThreshold-1 != 0 {
			lp.points = append(lp.points, path.waypoints.Back().Value.(*point))
		}
	}

	return lp
}
