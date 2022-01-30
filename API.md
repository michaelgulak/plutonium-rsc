# API Documentation

## Globals

`BANKERS` list of ids of bankers

`SLEEPING_BAG` id of sleeping bag.

## Events

###### on_progress_report()

Return a string dict from this function to generate a progress report.

###### on_kill_signal()

After this is called, the script has 15 seconds to clean up (ie. get out of combat) and call stop_account(), otherwise the bot will kill the account loop.

###### on_server_message(msg)
###### on_trade_request(name)
###### on_chat_message(msg, from_name)
###### on_private_message(msg, from_name)
###### on_player_damaged(damage, player)
###### on_npc_damaged(damage, npc)
###### on_npc_spawned(npc)
###### on_npc_despawned(npc)
###### on_npc_message(msg, npc, player)
###### on_npc_projectile(projectile_type, npc, player)
###### on_death()
###### on_ground_item_spawned(ground_item)
###### on_ground_item_despawned(ground_item)
###### on_object_spawned(object)
###### on_object_despawned(object)
###### on_wall_object_spawned(wall_object)
###### on_wall_object_despawned(wall_object)


## Base

###### stop_account()

Tells the bot to close the account loop, terminating the running account. Usually used with on_kill_signal() after the account is out of combat and the script prepares everything to be able to kill the account loop.

###### in_rect(x, z, width, height)
###### set_fatigue_tricking(True)
###### stop_script()
###### walk_to(x, z)
###### walk_path_to(x, z)

Capped at 200 distance to prevent burning out cpus. If you find this isn't enough for a long distance journey, then break up the calls to walk_path_to into shorter distances in conjunction with in_radius_of. Using a distance of 50 is preferred over the max of 200. FIXME: need a more efficient algorithm for walking long distances.

Can walk long distances across the map similar to the APOS PathWalker by Storm but with a simplified API and shouldn't have as many issues finding paths. Returns true or false depending on whether a path was found or not. Note that this function is expensive. Don't use it for too far of distances and return a high wait time before calling it again.

###### walk_path_depth_to(x, z, depth)

Capped at 800 depth to prevent burning out cpus. If you find this isn't enough for a long distance journey, then break up the calls to walk_path_to into shorter distances in conjunction with in_radius_of. Using a depth of 200 is preferred over the max of 800.FIXME: need a more efficient algorithm for walking long distances.

The same as walk_path_to except it takes a depth for the internal A Star algorithm. The depth is set to euclidean distance \* 4 for walk_path_to.

###### open_bank()

Code was taken from the AA series of APOS scripts, aka Chomp's scripts. Talks to banker and answers. Returns a wait time.

Proper use is like this:

```python
    if not is_bank_open():
        return open_bank()
```

###### get_pid()
###### logout()
###### get_fatigue()
###### get_x()
###### get_z()
###### in_combat()
###### is_skilling()
###### get_combat_style()
###### set_combat_style(idx)
###### get_max_stat(idx)
###### get_current_level(idx)
###### get_experience(idx)
###### get_hp_percent()
###### skip_tutorial()
###### is_reachable(x, z)
###### is_ground_item_reachable(ground_item)
###### is_object_reachable(object)
###### is_wall_object_reachable(wallobject)
###### distance_to(x, z)
###### distance(x1, z1, x2, z2)
###### in_radius_of(x, z, radius)
###### set_autologin(True)
###### is_appearance_screen()
###### send_appearance_update(head_gender, head_type, body_type, hair_colour, top_colour, pants_colour, skin_colour)
###### random(min, max)

Returns a random integer between min and max, inclusive.

## Ground items

#### Ground Item Object

Contains the following fields: `id`, `x`, `z`.

###### get_ground_items()
###### get_ground_item_count()
###### get_nearest_ground_item_by_id(id OR ids=[....], reachable=True, x=216, z=636, radius=10)

`x` and `z` are used as the start point for the radius. Otherwise they're unused.

###### get_nearest_ground_item_by_id_in_rect(id OR ids=[...], x=218, z=636, width=10, height=10, reachable=True)

Reachable is optional.

###### is_ground_item_at(id, x, z)
###### pickup_item(item)
###### use_item_on_ground_item(item, ground_item)
###### cast_on_ground_item(spellid, ground_item)

## Inventory items

#### Inventory Item Object

Contains the following fields: `id`, `amount`, `equipped`, `slot`. 

###### get_inventory_items()
###### get_inventory_item_by_id(id OR ids=[....])
###### get_inventory_count_by_id(id OR ids=[...])
###### has_inventory_item(id)
###### get_total_inventory_count()
###### get_empty_slots()
###### is_inventory_item_equipped(id)
###### use_item(item)
###### drop_item(item)
###### equip_item(item)
###### unequip_item(item)
###### use_item_with_item(item, item)
###### cast_on_item(spellid, item)
###### use_item_on_object(item, object)
###### use_item_on_wall_object(item, wall_object)
###### use_item_on_player(item, player)
###### use_sleeping_bag()

## NPCs

#### NPC Object

Contains the following fields: `x`, `z`, `sid`, `sprite`, `current_hp`, `max_hp`. 
Contains the following methods: `in_combat()`, `is_talking()`.

###### get_npcs()
###### get_nearest_npc_by_id(id OR ids=[...], in_combat=False, talking=False, reachable=True, x=218, z=636, radius=10)

`x` and `z` are used as the start point for the radius. Otherwise they're unused.

###### get_nearest_npc_by_id_in_rect(id OR ids=[...], in_combat=False, talking=False, reachable=True, x=218, z=636, width=10, height=10)
###### attack_npc(npc)
###### talk_to_npc(npc)
###### thieve_npc(npc)
###### cast_on_npc(spell_id, npc)
###### use_item_on_npc(item, npc)

## Quest menu

###### is_option_menu()
###### get_option_menu()
###### get_option_menu_count()
###### get_option_menu_option(idx)
###### get_option_menu_index(str)
###### answer(idx)

## Players

#### Player Object

Contains the following fields: `x`, `z`, `pid`, `sprite`, `current_hp`, `max_hp`, `username`, `combat_level`. 
Contains the following methods: `in_combat()`, `is_talking()`.

###### get_my_player()
###### get_players()
###### get_player_count()
###### get_player_by_name(str)
###### attack_player(player)
###### cast_on_player(spellid, player)
###### trade_player(player)
###### follow_player(player)
###### cast_on_self(spellid)

## Chat

##### Friend Object

Contains the following fields: `username`, `online`.

#### Ignored Object

Contains the following fields: `username`.

###### send_chat_message(str)
###### send_private_message(to_str, message_str)
###### add_friend(str)
###### remove_friend(str)
###### add_ignore(str)
###### remove_ignore(str)
###### get_friend_count()
###### get_friends()
###### is_friend(name_str)
###### is_ignored(name_str)
###### get_ignored()

## Trading

#### Trade Item Object

Contains the following fields: `id`, `amount`.

###### get_my_trade_items()
###### get_recipient_trade_items()
###### get_my_confirm_items()
###### get_recipient_confirm_items()
###### is_trade_offer_screen()
###### is_trade_confirm_screen()
###### trade_offer_item(amount, item)
###### has_my_offer(id, amount)
###### has_my_confirm(id, amount)
###### has_recipient_offer(id, amount)
###### has_recipient_confirm(id, amount)
###### is_trade_accepted()
###### is_recipient_trade_accepted()
###### is_trade_confirm_accepted()
###### accept_trade_offer()
###### confirm_trade()
###### decline_trade()

## Objects

#### Game Object Object

Contains the following fields: `id`, `x`, `z`, `dir`.

###### get_objects()
###### get_nearest_object_by_id(id OR ids=[...], reachable=True, x=218, z=636, radius=10)

`x` and `z` are used as the start point for the radius. Otherwise they're unused.

###### get_nearest_object_by_id_in_rect(id OR ids=[...], x=218, z=636, width=10, height=10, reachable=True)

Reachable is optional.

###### get_object_count()
###### get_object_from_coords(x, z)
###### is_object_at(x, z)
###### at_object(object)
###### at_object2(object)
###### cast_on_object(spellid, object)

## Wall Objects

#### Wall Object Object

Contains the following fields: `id`, `x`, `z`, `dir`.

###### get_wall_objects()
###### get_nearest_wall_object_by_id(id OR ids=[...], reachable=True, x=218, z=636, radius=10)

`x` and `z` are used as the start point for the radius. Otherwise they're unused.

###### get_nearest_wall_object_by_id_in_rect(id OR ids=[...], x=218, z=636, width=10, height=10, reachable=True)

Reachable is optional.

###### get_wall_object_count()
###### get_wall_object_from_coords(x, z)
###### at_wall_object(wall_object)
###### at_wall_object2(wall_object)
###### cast_on_wall_object(spellid, wall_object)

## Bank Items

#### Bank Item Object

Contains the following fields: `id`, `amount`.

###### get_bank_items()
###### deposit(id, amount)
###### deposit_all()
###### withdraw(id, amount)
###### get_bank_count(id OR ids=[...])
###### get_bank_size()
###### has_bank_item(id)
###### is_bank_open()
###### close_bank()

## Prayers

###### enable_prayer(idx)
###### disable_prayer(idx)
###### is_prayer_enabled(idx)

## Shop

#### Shop Item Object

Contains the following fields: `id`, `amount`, `price`.

###### get_shop_items()
###### is_shop_open()
###### get_shop_item_by_id(id)
###### buy_shop_item(id, amount)
###### sell_shop_item(id, amount)
###### close_shop()

## Quests

#### Quest Object

Contains the following fields: `id`, `name`, `stage`.

###### get_quests()
###### is_quest_complete(id)
###### get_quest_points()

## Raw Packet Operations

###### create_packet(opcode)
###### write_byte(int)
###### write_short(int)
###### write_int(int)
###### write_bytes(bytearray)
###### send_packet