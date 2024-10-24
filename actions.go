package main

import (
	"fmt"
)

func (c *client) walkToEntity(x1, z1, x2, z2 int, border bool) bool {
	return c.walkTo(x1, z1, x2, z2, true, border)
}

func (c *client) walkToWallObject(w *wallObject) bool {
	if w.dir == 0 {
		return c.walkToEntity(w.x, w.z-1, w.x, w.z, false)
	} else if w.dir != 1 {
		return c.walkToEntity(w.x, w.z, w.x, w.z, true)
	} else {
		return c.walkToEntity(w.x-1, w.z, w.x, w.z, false)
	}
}

func (c *client) walkToObject(o *object) bool {
	var (
		objWidth  int
		objHeight int
	)
	def := objectDefs[o.id]
	if o.dir != 0 && o.dir != 4 {
		objWidth = int(def.Height)
		objHeight = int(def.Width)
	} else {
		objHeight = int(def.Height)
		objWidth = int(def.Width)
	}
	destX := o.x
	destZ := o.z

	if def.Typ != 2 && def.Typ != 3 {
		return c.walkTo(o.x, o.z, objWidth-1+destX, destZ+objHeight-1, true, true)
	} else {
		if o.dir == 0 {
			objWidth++
			destX--
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

		return c.walkTo(destX, destZ, objWidth+(destX-1), objHeight+destZ-1, true, false)
	}
}

func (c *client) walkToGroundItem(xLow, zLow, xHigh, zHigh int) bool {
	if !c.walkToGroundItem1(xLow, zLow, xHigh, zHigh, false) {
		return c.walkToEntity(xLow, zLow, xHigh, zHigh, true)
	}
	return true
}

func (c *client) walkTo(xLow, zLow, xHigh, zHigh int, entity, border bool) bool {
	if c.account.debug {
		fmt.Printf("[%s] WALKTO (%d,%d,%d,%d,%v,%v)\n", c.user, xLow, zLow, xHigh, zHigh, entity, border)
	}

	l := c.findPath(c.x, c.z, xLow, xHigh, zLow, zHigh, border)
	px := c.pathX
	pz := c.pathZ

	if l == -1 {
		if !entity {
			if c.account.debug {
				fmt.Printf("[%s] No local path found to (%d,%d)\n", c.user, xLow, zLow)
			}
			return false
		}
		l = 1
		px[0] = int16(xLow)
		pz[0] = int16(zLow)
		if c.account.debug {
			fmt.Printf("[%s] Path length was -1 in walk to entity (%d,%d)\n", c.user, xLow, zLow)
		}
	}

	count := l - 1
	startX := int(px[count])
	startZ := int(pz[count])
	count--

	if entity {
		c.createPacket(16)
	} else {
		c.createPacket(187)
	}

	c.writeShort(c.regionX + startX)
	c.writeShort(c.regionZ + startZ)

	if entity && count == -1 && (startX+c.regionX)%5 == 0 {
		count = 0
	}

	for i := count; i >= 0 && i > count-25; i-- {
		c.writeByte(byte(int(px[i]) - startX))
		c.writeByte(byte(int(pz[i]) - startZ))
	}

	c.sendPacket()
	return true
}

func (c *client) walkToGroundItem1(xLow, zLow, xHigh, zHigh int, border bool) bool {
	l := c.findPath(c.x, c.z, xLow, xHigh, zLow, zHigh, border)

	if l == -1 {
		return false
	}

	if c.account.debug {
		fmt.Printf("[%s] WALKTOGROUNDITEM (%d,%d,%d,%d)\n", c.user, xLow, zLow, xHigh, zHigh)
	}

	px := c.pathX
	pz := c.pathZ

	count := l - 1
	startX := int(px[count])
	startZ := int(pz[count])
	count--

	c.createPacket(16)
	c.writeShort(c.regionX + startX)
	c.writeShort(c.regionZ + startZ)

	if count == -1 && (startX+c.regionX)%5 == 0 {
		count = 0
	}

	for i := count; i >= 0 && i > count-25; i-- {
		c.writeByte(byte(int(px[i]) - startX))
		c.writeByte(byte(int(pz[i]) - startZ))
	}

	c.sendPacket()

	return true
}

func (c *client) atObject(obj *object) {
	c.walkToObject(obj)
	if c.account.debug {
		fmt.Printf("[%s] ATOBJECT (%d,%d,%d,%d)\n", c.user, obj.id, obj.x, obj.z, obj.dir)
	}
	c.createPacket(136)
	c.writeShort(obj.x)
	c.writeShort(obj.z)
	c.sendPacket()
}

func (c *client) atObject2(obj *object) {
	c.walkToObject(obj)
	if c.account.debug {
		fmt.Printf("[%s] ATOBJECT2 (%d,%d,%d,%d)\n", c.user, obj.id, obj.x, obj.z, obj.dir)
	}
	c.createPacket(79)
	c.writeShort(obj.x)
	c.writeShort(obj.z)
	c.sendPacket()
}

func (c *client) atWallObject(obj *wallObject) {
	c.walkToWallObject(obj)
	if c.account.debug {
		fmt.Printf("[%s] ATWALLOBJECT (%d,%d,%d,%d)\n", c.user, obj.id, obj.x, obj.z, obj.dir)
	}
	c.createPacket(14)
	c.writeShort(obj.x)
	c.writeShort(obj.z)
	c.writeByte(byte(obj.dir))
	c.sendPacket()
}

func (c *client) atWallObject2(obj *wallObject) {
	c.walkToWallObject(obj)
	if c.account.debug {
		fmt.Printf("[%s] ATWALLOBJECT2 (%d,%d,%d,%d)\n", c.user, obj.id, obj.x, obj.z, obj.dir)
	}
	c.createPacket(127)
	c.writeShort(obj.x)
	c.writeShort(obj.z)
	c.writeByte(byte(obj.dir))
	c.sendPacket()
}

func (c *client) answer(index int) {
	if c.account.debug {
		fmt.Printf("[%s] ANSWER (%d)\n", c.user, index)
	}
	c.createPacket(116)
	c.writeByte(byte(index))
	c.sendPacket()
	c.optionMenuVisible = false
}

func (c *client) attackNPC(n *npc) {
	c.walkToEntity(n.x, n.z, n.x, n.z, false)
	if c.account.debug {
		fmt.Printf("[%s] ATTACKNPC (%d,%d,%d)\n", c.user, n.id, n.x, n.z)
	}
	c.createPacket(190)
	c.writeShort(n.serverIndex)
	c.sendPacket()
}

func (c *client) talkToNPC(n *npc) {
	c.walkToEntity(n.x, n.z, n.x, n.z, false)
	if c.account.debug {
		fmt.Printf("[%s] TALKTONPC (%d,%d,%d)\n", c.user, n.id, n.x, n.z)
	}
	c.createPacket(153)
	c.writeShort(n.serverIndex)
	c.sendPacket()
}

func (c *client) castOnNPC(spell int, n *npc) {
	if c.account.debug {
		fmt.Printf("[%s] CASTONNPC (%d,%d,%d,%d)\n", c.user, spell, n.id, n.x, n.z)
	}
	c.createPacket(50)
	c.writeShort(spell)
	c.writeShort(n.serverIndex)
	c.sendPacket()
}

func (c *client) castOnInventoryItem(spell int, item *item) {
	if c.account.debug {
		fmt.Printf("[%s] CASTONINVITEM (%d,%d,%d)\n", c.user, spell, item.id, item.slot)
	}
	c.createPacket(4)
	c.writeShort(spell)
	c.writeShort(item.slot)
	c.sendPacket()
}

func (c *client) castOnGroundItem(spell int, g *groundItem) {
	c.walkToGroundItem(g.x, g.z, g.x, g.z)
	if c.account.debug {
		fmt.Printf("[%s] CASTONGROUNDITEM (%d,%d,%d,%d)\n", c.user, spell, g.id, g.x, g.z)
	}
	c.createPacket(249)
	c.writeShort(spell)
	c.writeShort(g.x)
	c.writeShort(g.z)
	c.writeShort(g.id)
	c.sendPacket()
}

func (c *client) castOnPlayer(spell int, p *player) {
	c.walkToEntity(p.x, p.z, p.x, p.z, false)
	if c.account.debug {
		fmt.Printf("[%s] CASTONPLAYER (%d,%d,%d,%d)\n", c.user, spell, p.serverIndex, p.x, p.z)
	}
	c.createPacket(229)
	c.writeShort(spell)
	c.writeShort(p.serverIndex)
	c.sendPacket()
}

func (c *client) castOnSelf(spell int) {
	if c.account.debug {
		fmt.Printf("[%s] CASTONSELF (%d)\n", c.user, spell)
	}
	c.createPacket(137)
	c.writeShort(spell)
	c.sendPacket()
}

func (c *client) sendSleepWord(word string) {
	if c.account.debug {
		fmt.Printf("[%s] SENDSLEEPWORD (%s)\n", c.user, word)
	}
	c.createPacket(45)
	c.writeBytes([]byte(word))
	c.writeByte(10)
	c.sendPacket()
}

func (c *client) pickupGroundItem(item *groundItem) {
	c.walkToGroundItem(item.x, item.z, item.x, item.z)
	if c.account.debug {
		fmt.Printf("[%s] PICKUPGROUNDITEM (%d,%d,%d)\n", c.user, item.id, item.x, item.z)
	}
	c.createPacket(247)
	c.writeShort(item.x)
	c.writeShort(item.z)
	c.writeShort(item.id)
	c.sendPacket()
}

func (c *client) useInventoryItemOnNPC(item *item, n *npc) {
	c.walkToEntity(n.x, n.z, n.x, n.z, false)
	if c.account.debug {
		fmt.Printf("[%s] USEITEMONNPC (%d,%d,%d,%d,%d)\n", c.user, item.id, item.slot, n.id, n.x, n.z)
	}
	c.createPacket(135)
	c.writeShort(n.serverIndex)
	c.writeShort(item.slot)
	c.sendPacket()
}

func (c *client) useInventoryItemOnPlayer(item *item, p *player) {
	c.walkToEntity(p.x, p.z, p.x, p.z, false)
	if c.account.debug {
		fmt.Printf("[%s] USEITEMONPLAYER (%d,%d,%d,%d,%d)\n", c.user, item.id, item.slot, p.serverIndex, p.x, p.z)
	}
	c.createPacket(113)
	c.writeShort(p.serverIndex)
	c.writeShort(item.slot)
	c.sendPacket()
}

func (c *client) useInventoryItemOnInventoryItem(item1, item2 *item) {
	if c.account.debug {
		fmt.Printf("[%s] USEITEMONITEM (%d,%d,%d,%d)\n", c.user, item1.id, item1.slot, item2.id, item2.slot)
	}
	c.createPacket(91)
	c.writeShort(item1.slot)
	c.writeShort(item2.slot)
	c.sendPacket()
}

func (c *client) useInventoryItemOnObject(item *item, obj *object) {
	c.walkToObject(obj)
	if c.account.debug {
		fmt.Printf("[%s] USEITEMONOBJECT (%d,%d,%d,%d,%d,%d)\n", c.user, item.id, item.slot, obj.id, obj.x, obj.z, obj.dir)
	}
	c.createPacket(115)
	c.writeShort(obj.x)
	c.writeShort(obj.z)
	c.writeShort(item.slot)
	c.sendPacket()
}

func (c *client) useInventoryItemOnWallObject(item *item, obj *wallObject) {
	c.walkToWallObject(obj)
	if c.account.debug {
		fmt.Printf("[%s] USEITEMONWALLOBJECT (%d,%d,%d,%d,%d,%d)\n", c.user, item.id, item.slot, obj.id, obj.x, obj.z, obj.dir)
	}
	c.createPacket(161)
	c.writeShort(obj.x)
	c.writeShort(obj.z)
	c.writeByte(byte(obj.dir))
	c.writeShort(item.slot)
	c.sendPacket()
}

func (c *client) castOnObject(spell int, obj *object) {
	c.walkToObject(obj)
	if c.account.debug {
		fmt.Printf("[%s] CASTONOBJECT (%d,%d,%d,%d,%d)\n", c.user, spell, obj.id, obj.x, obj.z, obj.dir)
	}
	c.createPacket(99)
	c.writeShort(spell)
	c.writeShort(obj.x)
	c.writeShort(obj.z)
	c.sendPacket()
}

func (c *client) castOnWallObject(spell int, obj *wallObject) {
	c.walkToWallObject(obj)
	if c.account.debug {
		fmt.Printf("[%s] CASTONOWALLBJECT (%d,%d,%d,%d,%d)\n", c.user, spell, obj.id, obj.x, obj.z, obj.dir)
	}
	c.createPacket(180)
	c.writeShort(spell)
	c.writeShort(obj.x)
	c.writeShort(obj.z)
	c.writeByte(byte(obj.dir))
	c.sendPacket()
}

func (c *client) equipInventoryItem(item *item) {
	if c.account.debug {
		fmt.Printf("[%s] EQUIPITEM (%d,%d)\n", c.user, item.id, item.slot)
	}
	c.createPacket(169)
	c.writeShort(item.slot)
	c.sendPacket()
}

func (c *client) unequipInventoryItem(item *item) {
	if c.account.debug {
		fmt.Printf("[%s] UNEQUIPITEM (%d,%d)\n", c.user, item.id, item.slot)
	}
	c.createPacket(170)
	c.writeShort(item.slot)
	c.sendPacket()
}

func (c *client) dropInventoryItem(item *item) {
	if c.account.debug {
		fmt.Printf("[%s] DROPITEM (%d,%d)\n", c.user, item.id, item.slot)
	}
	c.createPacket(246)
	c.writeShort(item.slot)
	c.writeShort(item.amount)
	c.sendPacket()
}

func (c *client) enablePrayer(id int) {
	if c.account.debug {
		fmt.Printf("[%s] ENABLEPRAYER (%d)\n", c.user, id)
	}
	c.createPacket(60)
	c.writeByte(byte(id))
	c.sendPacket()
}

func (c *client) disablePrayer(id int) {
	if c.account.debug {
		fmt.Printf("[%s] DISABLEPRAYER (%d)\n", c.user, id)
	}
	c.createPacket(254)
	c.writeByte(byte(id))
	c.sendPacket()
}

func (c *client) sendChatMessage(text string) {
	if c.account.debug {
		fmt.Printf("[%s] SENDCHATMESSAGE (%s)\n", c.user, text)
	}
	c.createPacket(216)
	c.writeSmart08_16(len(text))
	c.writeHuffman(text)
	c.sendPacket()
}

func (c *client) sendPrivateMessage(playerName, text string) {
	if c.account.debug {
		fmt.Printf("[%s] SENDPRIVMESSAGE (%s,%s)\n", c.user, playerName, text)
	}
	c.createPacket(218)
	c.writeBytes([]byte(playerName))
	c.writeByte(10)
	c.writeSmart08_16(len(text))
	c.writeHuffman(text)
	c.sendPacket()
}

func (c *client) addFriend(playerName string) {
	if c.account.debug {
		fmt.Printf("[%s] ADDFRIEND (%s)\n", c.user, playerName)
	}
	c.createPacket(195)
	c.writeBytes([]byte(playerName))
	c.writeByte(10)
	c.sendPacket()
}

func (c *client) removeFriend(playerName string) {
	if c.account.debug {
		fmt.Printf("[%s] REMOVEFRIEND (%s)\n", c.user, playerName)
	}
	c.createPacket(167)
	c.writeBytes([]byte(playerName))
	c.writeByte(10)
	c.sendPacket()
}

func (c *client) addIgnore(playerName string) {
	if c.account.debug {
		fmt.Printf("[%s] ADDIGNORE (%s)\n", c.user, playerName)
	}
	c.createPacket(132)
	c.writeBytes([]byte(playerName))
	c.writeByte(10)
	c.sendPacket()
}

func (c *client) removeIgnore(playerName string) {
	if c.account.debug {
		fmt.Printf("[%s] REMOVE IGNORE (%s)\n", c.user, playerName)
	}
	c.createPacket(241)
	c.writeBytes([]byte(playerName))
	c.writeByte(10)
	c.sendPacket()
}

func (c *client) closeBank() {
	if c.account.debug {
		fmt.Printf("[%s] CLOSEBANK\n", c.user)
	}
	c.createPacket(212)
	c.sendPacket()
}

func (c *client) withdraw(id, amount int) {
	if c.account.debug {
		fmt.Printf("[%s] WITHDRAW (%d,%d)\n", c.user, id, amount)
	}
	c.createPacket(22)
	c.writeShort(id)
	c.writeInt(amount)
	c.sendPacket()
}

func (c *client) deposit(id, amount int) {
	if c.account.debug {
		fmt.Printf("[%s] DEPOSIT (%d,%d)\n", c.user, id, amount)
	}
	c.createPacket(23)
	c.writeShort(id)
	c.writeInt(amount)
	c.sendPacket()
}

func (c *client) setCombatStyle(id int) {
	if c.account.debug {
		fmt.Printf("[%s] SETCOMBATSTYLE (%d)\n", c.user, id)
	}
	c.createPacket(29)
	c.writeByte(byte(id))
	c.sendPacket()
}

func (c *client) logout() {
	if c.account.debug {
		fmt.Printf("[%s] LOGOUT\n", c.user)
	}
	c.createPacket(102)
	c.sendPacket()
}

func (c *client) sendAppearanceUpdate(headRestrictions, headType, bodyType, hairColour, topColour, trouserColour, skinColour int) {
	if c.account.debug {
		fmt.Printf("[%s] SENDAPPEARANCEUPDATE (%d,%d,%d,%d,%d,%d,%d)\n", c.user, headRestrictions, headType, bodyType, hairColour, topColour, trouserColour, skinColour)
	}
	c.createPacket(235)
	c.writeByte(byte(headRestrictions))
	c.writeByte(byte(headType))
	c.writeByte(byte(bodyType))
	c.writeByte(2)
	c.writeByte(byte(hairColour))
	c.writeByte(byte(topColour))
	c.writeByte(byte(trouserColour))
	c.writeByte(byte(skinColour))
	c.writeByte(0)
	c.writeByte(0)
	c.sendPacket()
	c.appearanceChange = false
}

func (c *client) tradePlayer(p *player) {
	if c.account.debug {
		fmt.Printf("[%s] TRADEPLAYER (%d,%d,%d)\n", c.user, p.serverIndex, p.x, p.z)
	}
	c.createPacket(142)
	c.writeShort(p.serverIndex)
	c.sendPacket()
}

func (c *client) followPlayer(p *player) {
	if c.account.debug {
		fmt.Printf("[%s] FOLLOWPLAYER (%d,%d,%d)\n", c.user, p.serverIndex, p.x, p.z)
	}
	c.createPacket(165)
	c.writeShort(p.serverIndex)
	c.sendPacket()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (c *client) offerItemToTrade(amount int, it *item) {
	if c.account.debug {
		fmt.Printf("[%s] OFFERITEM (%d,%d,%d)\n", c.user, amount, it.id, it.slot)
	}
	offerSuccess := false
	offered := 0
	def := itemDefs[it.id]

	for i := 0; i < c.myTradeCount; i++ {
		if it.id == c.myTradeItems[i].id {
			if def.IsStackable > 0 {
				c.myTradeItems[i].amount = c.myTradeItems[i].amount + amount
				if c.myTradeItems[i].amount > it.amount {
					c.myTradeItems[i].amount = it.amount
				}
				offerSuccess = true
			} else {
				offered++
			}
		}
	}

	invAvailable := c.getInventoryCountByID(it.id)
	if invAvailable <= offered {
		offerSuccess = true
	}

	if !offerSuccess {
		for i := 0; i < amount && c.myTradeCount < 12 && invAvailable > offered; i++ {
			c.myTradeItems[c.myTradeCount].id = it.id
			c.myTradeItems[c.myTradeCount].amount = 1
			offerSuccess = true
			offered++
			c.myTradeCount++
			if i == 0 && def.IsStackable > 0 {
				c.myTradeItems[c.myTradeCount-1].amount = min(amount, it.amount)
				break
			}
		}
	}

	if offerSuccess {
		c.createPacket(46)
		c.writeByte(byte(c.myTradeCount))

		for i := 0; i < c.myTradeCount; i++ {
			c.writeShort(c.myTradeItems[i].id)
			c.writeInt(c.myTradeItems[i].amount)
			c.writeShort(0)
		}
		c.sendPacket()
	}
}

func (c *client) acceptTradeOffer() {
	if c.account.debug {
		fmt.Printf("[%s] ACCEPTTRADE\n", c.user)
	}
	c.createPacket(55)
	c.sendPacket()
}

func (c *client) acceptTradeConfirm() {
	if c.account.debug {
		fmt.Printf("[%s] ACCEPTCONFIRM\n", c.user)
	}
	c.tradeConfirmAccepted = true
	c.createPacket(104)
	c.sendPacket()
}

func (c *client) declineTrade() {
	if c.account.debug {
		fmt.Printf("[%s] DECLINETRADE\n", c.user)
	}
	c.createPacket(230)
	c.sendPacket()
}

func (c *client) useInventoryItemOnGroundItem(item1 *item, item2 *groundItem) {
	c.walkToGroundItem(item2.x, item2.z, item2.x, item2.z)
	if c.account.debug {
		fmt.Printf("[%s] USEITEMONGROUNDITEM (%d,%d,%d,%d,%d)\n", c.user, item1.id, item1.slot, item2.id, item2.x, item2.z)
	}
	c.createPacket(53)
	c.writeShort(item2.x)
	c.writeShort(item2.z)
	c.writeShort(item1.slot)
	c.writeShort(item2.id)
	c.sendPacket()
}

func (c *client) useInventoryItem(item *item) {
	if c.account.debug {
		fmt.Printf("[%s] USEITEM (%d,%d)\n", c.user, item.id, item.slot)
	}
	c.createPacket(90)
	c.writeShort(item.slot)
	c.writeInt(1)
	c.writeByte(0)
	c.sendPacket()
}

func (c *client) thieveNPC(n *npc) {
	c.walkToEntity(n.x, n.z, n.x, n.z, false)
	if c.account.debug {
		fmt.Printf("[%s] THIEVENPC (%d,%d,%d)\n", c.user, n.serverIndex, n.x, n.z)
	}
	c.createPacket(202)
	c.writeShort(n.serverIndex)
	c.sendPacket()
}

func (c *client) attackPlayer(n *player) {
	c.walkToEntity(n.x, n.z, n.x, n.z, false)
	if c.account.debug {
		fmt.Printf("[%s] ATTACKPLAYER (%d,%d,%d)\n", c.user, n.serverIndex, n.x, n.z)
	}
	c.createPacket(171)
	c.writeShort(n.serverIndex)
	c.sendPacket()
}

func (c *client) closeShop() {
	if c.account.debug {
		fmt.Printf("[%s] CLOSESHOP\n", c.user)
	}
	c.createPacket(166)
	c.sendPacket()
}

func (c *client) buyItem(id, amount int) {
	if c.account.debug {
		fmt.Printf("[%s] BUYITEM (%d,%d)\n", c.user, id, amount)
	}
	c.createPacket(236)
	c.writeShort(id)
	c.writeShort(c.getShopAmount(id))
	c.writeShort(amount)
	c.sendPacket()
}

func (c *client) sellItem(id, amount int) {
	if c.account.debug {
		fmt.Printf("[%s] SELLITEM (%d,%d)\n", c.user, id, amount)
	}
	c.createPacket(221)
	c.writeShort(id)
	c.writeShort(c.getShopAmount(id))
	c.writeShort(amount)
	c.sendPacket()
}

func (c *client) skipTutorial() {
	if c.account.debug {
		fmt.Printf("[%s] SKIPTUTORIAL\n", c.user)
	}
	c.createPacket(84)
	c.sendPacket()
}
