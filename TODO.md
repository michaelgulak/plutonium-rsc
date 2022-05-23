# TODO 

### Bot

- [x] *HIGH PRIORITY*: Optimize or change A Star algorithm such that it can path longer distances
- [ ] Make a frontend for the bot
- [x] Make scripts retain state between logouts
- [ ] Add log function for scripts (so it prints username before argument)

### Functions

##### Events

- [x] onServerMessage
- [x] onTradeRequest
- [x] onChatMessage
- [x] onPrivateMessage
- [x] onPlayerDamaged
- [x] onNpcDamaged
- [x] onNpcSpawned
- [x] onNpcDespawned
- [x] onNpcProjectileDamaged
- [x] onDeath
- [x] onGroundItemSpawned
- [x] onGroundItemDespawned
- [x] onObjectSpawned
- [x] onObjectDespawned
- [x] onWallObjectSpawned
- [x] onWallObjectDespawned

##### Accessors

###### Base

- [x] inRect
- [x] setFatigueTricking
- [x] stopScript
- [x] walkPathTo
- [x] getPID
- [x] logout
- [x] getFatigue
- [x] getX
- [x] getZ
- [x] isTalking (npc, myplayer)
- [x] inCombat
- [x] isSkilling
- [x] getFightMode
- [x] setFightMode
- [x] getBaseLevel
- [x] getCurrentLevel
- [x] getExperience
- [x] getHPPercent
- [x] castOnSelf
- [x] getTotalInventoryCount
- [x] getEmptySlots
- [x] getInventoryCount(...ids)
- [x] hasInventoryItem
- [x] useSleepingBag
- [x] isItemEquippedByID
- [x] getQuestPoints
- [x] isQuestComplete
- [x] walkTo
- [x] isReachable
- [x] distanceTo
- [x] inRadiusOf
- [x] setAutoLogin
- [x] isAppearanceScreen
- [x] sendAppearanceUpdate

###### Ground items

- [x] getGroundItems
- [x] getGroundItemCount
- [x] getNearestGroundIemByID
- [x] isGroundItemAt(id, x, z)
- [x] pickupItem
- [x] useItemOnGroundItem
- [x] castOnGroundItem

###### Inventory items

- [x] getInventoryItems
- [x] getInventoryItemByID
- [x] useItem
- [x] dropItem
- [x] wearItem
- [x] removeItem
- [x] useItemWithItem
- [x] castOnItem
- [x] useItemOnObject
- [x] useItemOnWallObject
- [x] useItemWithPlayer

###### NPCs

- [x] getNPCs
- [x] getNPCInRect
- [x] attackNPC
- [x] talkToNPC
- [x] thieveNPC
- [x] mageNPC (cast_on_npc)
- [x] useOnNPC

###### Quest menu

- [x] isQuestMenu
- [x] questMenuOptions (get_option_menu)
- [x] questMenuCount
- [x] getQuestMenuOption
- [x] getMenuIndex
- [x] answer

###### Players

- [x] getMyPlayer
- [x] getPlayers
- [x] getPlayerCount
- [x] getPlayerByName
- [x] isPlayerInCombat (player.inCombat())
- [x] isPlayerTalking
- [x] getPlayerCount
- [x] attackPlayer
- [x] magePlayer (cast_on_player)
- [x] sendTradeRequest (trade_player)
- [x] followPlayer

###### Chat

- [x] sendChatMessage
- [x] sendPrivateMessage
- [x] addFriend
- [x] removeFriend
- [x] addIgnore
- [x] removeIgnore
- [x] getFriendCount
- [x] getFriendNames (get_friends)
- [x] isFriend(name)
- [x] isIgnored(name)
- [x] getIgnored

###### Trading

- [x] getMyTradeItems
- [x] getRecipientTradeItems
- [x] getMyConfirmTradeItems
- [x] getRecipientConfirmTradeItems
- [x] isInTradeOffer
- [x] isInTradeConfirm
- [x] offerItemTrade
- [x] hasMyOffer
- [x] hasMyConfirm
- [x] hasOtherOffer(id, amount)
- [x] hasRecipientConfirm(id, amount)
- [x] isTradeAccepted
- [x] isRecipientTradeAccepted
- [x] isTradeConfirmAccepted
- [x] acceptOffer
- [x] confirmTrade
- [x] declineTrade

###### Objects

- [x] getObjects
- [x] getObjectByID(...ids)
- [x] getObjectCount
- [x] getObjectIDFromCoords
- [x] isObjectAt(x, z)
- [x] atObject
- [x] atObject2

###### Wall Objects

- [x] getWallObjects
- [x] getWallObjectByID(...ids)
- [x] getWallObjectIDFromCoords(x, z)
- [x] getWallObjectCount
- [x] atWallObject
- [x] atWallObject2

###### Bank Items

- [x] getBankItems
- [x] deposit
- [x] depositAll
- [x] withdraw
- [x] bankCount(...ids)
- [x] getBankSize
- [x] hasBankItem
- [x] isBanking
- [x] closeBank

###### Prayers

- [x] enablePrayer
- [x] disablePrayer
- [x] isPrayerEnabled

###### Shop

- [x] getShopItems
- [x] isShopOpen
- [x] getShopItemByID
- [x] buyShopItem(item, amount)
- [x] sellShopItem(item, amount)
- [x] closeShop