# API Documentation

## Globals

`BANKERS` list of ids of bankers

`SLEEPING_BAG` id of sleeping bag

`settings` dict with script settings

## Events

###### on_progress_report()

Return a string dict from this function to generate a progress report.

###### on_kill_signal()

After this is called, the script has 15 seconds to clean up (ie. get out of combat) and call stop_account(), otherwise the bot will kill the account loop.

###### on_load()

This event is called when the client is fully loaded in the game world. It gets called every time the bot logs into an account.

###### on_server_tick(tick)

Called every time the npc update packet comes in, which coincides with every server tick.

###### on_system_update(seconds)
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
###### on_fatigue_update(fatigue, accurate_fatigue)

## Base

###### stop_account()

Tells the bot to close the account loop, terminating the running account. Usually used with on_kill_signal() after the account is out of combat and the script prepares everything to be able to kill the account loop.

###### in_rect(area_x, area_z, width, height)

Checks whether the player is in a rectangular area. `area_x` and `area_z` are the starting point which should be the north-west coordinate. The width extends east, while the height extends south. Note that just subtracting the coordinates will not work to calculate the width or height. For example, width should be calculated as `x1 - x2 + 1` (x1 is first because the x coordinate decreases in RSC as you move east).

###### point_in_rect(point_x, point_z, area_x, area_z, width, height)

Same as `in_rect` except instead of using the client coordinates it checks the supplied `point_x` and `point_z`.

###### point_in_polygon(x, z, polygon)

Checks whether a point is in the supplied polygon. `polygon` should be an array of ints in the format `[x1, z1, x2, z2, x3, z3, ...]`.

###### at(x, z)

Checks whether the client coords are equal to `x` and `z`.

###### set_wakeup_at(int)

Changes the default (0) wakeup fatigue setting to whatever you want.

###### set_fatigue_tricking(True)

Sets the wakeup fatigue setting to 99.

###### stop_script()
###### log(str)

Prints to the console with user prefix.

###### debug(str)

Prints to the console if the debug flag is set to true.

###### walk_to(x, z)
###### walk_to_entity(x, z)
###### walk_path_to(x, z)

Calculates a path to the specified coordinates and walks to the first point.

Will fail to path in certain circumstances where the distance is a lot shorter than the path distance.

This method will fail if distance > 200. Use `calculate_path` instead or try supplying your own depth with `walk_path_depth_to`.

It's preferable to use `calculate_path` for long distances as this method will calculate the path on every call which can be expensive. Also the depth for the A* algorithm will probably be calculated more efficiently with `calculate_path`.

###### walk_path_depth_to(x, z, depth)

Same as `walk_path_to` except you must specify a depth for the internal A* algorithm. This will be more efficient than calling `walk_path_to` if you choose something less than `distance * 4`.

This method will fail to path if depth > 800.

###### open_bank()

Code was taken from the AA series of APOS scripts, aka Chomp's scripts. Talks to banker and answers. Returns a wait time.

Proper use is like this:

```python
    if not is_bank_open():
        return open_bank()
```
###### disconnect_for(seconds)
###### get_pid()
###### logout()
###### get_fatigue()
###### get_accurate_fatigue()
###### get_x()
###### get_z()
###### in_combat()
###### is_skilling()
###### is_sleeping()

Can only be used in event handlers.

###### get_combat_style()
###### set_combat_style(id)
###### get_max_stat(id)
###### get_current_stat(id)
###### get_experience(id)
###### get_hp_percent()
###### get_item_name(id)
###### skip_tutorial()
###### is_system_update()
###### is_reachable(x, z)
###### is_ground_item_reachable(ground_item)
###### is_object_reachable(object)
###### is_wall_object_reachable(wall_object)
###### distance_to(x, z)
###### distance(x1, z1, x2, z2)
###### in_radius_of(x, z, radius)
###### set_autologin(True)
###### is_appearance_screen()
###### send_appearance_update(head_gender, head_type, body_type, hair_colour, top_colour, pants_colour, skin_colour)
###### get_equipment_stat(id)
###### set_debug(True)
###### random(min, max)

Returns a random integer between min and max, inclusive. Min and max must be >= 0.

## Paths

#### Path Object

Contains the following methods: `complete()`, `reset()`, `reverse()`, `process()`, `walk()`, `next_x()`, `next_z()`, `set_nearest()`, `length()`.

The `process` method must be called before calling `walk`. This is to make handling obstacles with `next_x` and `next_z` cleaner. For example you may want to call `process`, check if `next_z` is beyond some gate or other obstacle, then handle the obstacle before continuing to walk.

The `walk` method returns false if the next tile is unreachable.

The `set_nearest` method can be called on the path to set the path index to the nearest point to the player. This is useful if you calculate a path between two points and you're somewhere in between the two points because you stopped your script before the player reached the destination.

###### calculate_path_to(x, z, depth, skip_local=True, max_depth=512)

Calculates a path from the client's current coordinates to the specified coordinates. Returns a path object or `None` if a path was not found.

The `depth` argument is optional. Use a depth for the A* algorithm for optimal use of this method. Otherwise it will test depths sequentially, wasting CPU.

The `skip_local` argument is optional. Must use kwargs to use it. If set to true, will skip local blocked tiles and allow you to calculate a blocked path, for example.

The `max_depth` argument is optional. Valid values for it are `1, 2, 4, 8, 16, 32, 64, 128, 256, 512`. Must use kwargs to use it. It's only useful if you didn't already set the depth. Using it is preferable because the algorithm won't use tons of memory testing depths on failed attempts if you want to calculate a relatively short blocked path.

Correct usage is something like this:

```python
path = None

def loop():
    global path

    if path != None:
        path.process()
        if not path.complete():
            path.walk()
            return 600
        else:
            path = None

    if get_x() != 128 or get_z() != 640:
        path = calculate_path_to(128, 640)
        if path == None:
            log("failed to path")
            stop_script()
            return 1000

        return 250

    return 1000
```

###### calculate_path(start_x, start_z, x, z, depth, skip_local=True, max_depth=512)

The same as `calculate_path_to` except it takes starting coordinates.

###### calculate_path_through(points=[(x1, z1), (x2, z2), ....], start_x=648, start_z=632, depth=32, skip_local=True, max_depth=512)

This method must be used with kwargs. You may not pass normal arguments.

Calculates paths from `start_x` and `start_z` or the client's current coordinates if they are not supplied, through all the points supplied. It then merges them into one path. Returns a path object and should be handled the same way as `calculate_path_to`. 

`start_x` and `start_z` are optional. `depth` is also optional and if you supply it, it will be used between each set of coordinates. It is not the depth of the entire path. 

The `skip_local` argument is optional. Must use kwargs to use it. If set to true, will skip local blocked tiles and allow you to calculate a blocked path, for example.

The `max_depth` argument is optional. Valid values for it are `1, 2, 4, 8, 16, 32, 64, 128, 256, 512`. Must use kwargs to use it. It's only useful if you didn't already set the depth. Using it is preferable because the algorithm won't use tons of memory testing depths on failed attempts if you want to calculate a relatively short blocked path. Remember that the `max_depth` applies only between points with `calculate_path_through`, not the entire path.

Example: 

```python
path = calculate_path_through(points=[(447, 482), (445, 474)])
```

###### create_path([x1, z1, x2, z2, ...])

One way to use the path object is to create your own path using this method. Points in the path should be at most around 20 tiles apart. You may want to stay within 10-15 tiles apart to be safe.

## Ground items

#### Ground Item Object

Contains the following fields: `id`, `x`, `z`.

###### get_ground_item_count()
###### get_ground_item_at_index(index)
###### get_ground_items()
###### get_nearest_ground_item_by_id(id OR ids=[...], reachable=True, x=216, z=636, radius=10)

`reachable`, `x`, `z`, and `radius` are optional. `x` and `z` are used as the start point for the radius. Otherwise they're unused.

###### get_nearest_ground_item_by_id_in_rect(id OR ids=[...], x=218, z=636, width=10, height=10, reachable=True)

Reachable is optional.

###### is_ground_item_at(id, x, z)
###### pickup_item(ground_item)
###### use_item_on_ground_item(item, ground_item)
###### cast_on_ground_item(spell_id, ground_item)

## Inventory items

#### Inventory Item Object

Contains the following fields: `id`, `amount`, `equipped`, `slot`. 

###### get_total_inventory_count()
###### get_inventory_item_at_index(index)
###### get_inventory_items()
###### get_inventory_item_by_id(id OR ids=[...])
###### get_inventory_count_by_id(id OR ids=[...])
###### get_inventory_item_except([545, 631, ...])

Returns the first item that doesn't match the supplied array of ids.

###### has_inventory_item(id)
###### get_empty_slots()
###### is_inventory_item_equipped(id)
###### use_item(item)
###### drop_item(item)
###### equip_item(item)
###### unequip_item(item)
###### use_item_with_item(item, item)
###### cast_on_item(spell_id, item)
###### use_item_on_object(item, object)
###### use_item_on_wall_object(item, wall_object)
###### use_item_on_player(item, player)
###### use_item_on_npc(item, npc)
###### use_sleeping_bag()

## NPCs

#### NPC Object

Contains the following fields: `id`, `x`, `z`, `sid`, `sprite`, `current_hp`, `max_hp`. 

Contains the following methods: `in_combat()`, `is_talking()`.

###### get_npc_count()
###### get_npc_at_index(index)
###### get_npcs()
###### get_nearest_npc_by_id(id OR ids=[...], in_combat=False, talking=False, reachable=True, x=218, z=636, radius=10)

`in_combat`, `talking`, `reachable`, `radius`, `x`, and `z` are optional. `x` and `z` are used as the start point for the radius. Otherwise they're unused.

###### get_nearest_npc_by_id_in_rect(id OR ids=[...], in_combat=False, talking=False, reachable=True, x=218, z=636, width=10, height=10)
###### attack_npc(npc)
###### talk_to_npc(npc)
###### thieve_npc(npc)
###### cast_on_npc(spell_id, npc)

## Quest menu

###### is_option_menu()
###### get_option_menu()
###### get_option_menu_count()
###### get_option_menu_option(index)
###### get_option_menu_index(str)
###### answer(index)

## Players

#### Player Object

Contains the following fields: `x`, `z`, `pid`, `sprite`, `current_hp`, `max_hp`, `username`, `combat_level`.

Contains the following methods: `in_combat()`, `is_talking()`.

###### get_my_player()
###### get_player_count()
###### get_player_at_index(index)
###### get_players()
###### get_player_by_name(str)
###### attack_player(player)
###### cast_on_player(spell_id, player)
###### trade_player(player)
###### follow_player(player)
###### cast_on_self(spell_id)

## Chat

#### Friend Object

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

###### get_object_count()
###### get_object_at_index(index)
###### get_objects()
###### get_nearest_object_by_id(id OR ids=[...], reachable=True, x=218, z=636, radius=10)

`reachable`, `x`, `z`, and `radius` are optional. `x` and `z` are used as the start point for the radius. Otherwise they're unused.

###### get_nearest_object_by_id_in_rect(id OR ids=[...], x=218, z=636, width=10, height=10, reachable=True)

Reachable is optional.

###### get_object_from_coords(x, z)
###### is_object_at(x, z)
###### at_object(object)
###### at_object2(object)
###### cast_on_object(spell_id, object)

## Wall Objects

#### Wall Object Object

Contains the following fields: `id`, `x`, `z`, `dir`.

###### get_wall_object_count()
###### get_wall_object_at_index(index)
###### get_wall_objects()
###### get_nearest_wall_object_by_id(id OR ids=[...], reachable=True, x=218, z=636, radius=10)

`reachable`, `x`, `z`, and `radius` are optional. `x` and `z` are used as the start point for the radius. Otherwise they're unused.

###### get_nearest_wall_object_by_id_in_rect(id OR ids=[...], x=218, z=636, width=10, height=10, reachable=True)

Reachable is optional.

###### get_wall_object_from_coords(x, z)
###### at_wall_object(wall_object)
###### at_wall_object2(wall_object)
###### cast_on_wall_object(spell_id, wall_object)

## Bank Items

#### Bank Item Object

Contains the following fields: `id`, `amount`.

###### get_bank_size()
###### get_bank_item_at_index(index)
###### get_bank_items()
###### deposit(id, amount)
###### withdraw(id, amount)
###### get_bank_count(id OR ids=[...])
###### has_bank_item(id)
###### is_bank_open()
###### close_bank()

## Prayers

###### enable_prayer(id)
###### disable_prayer(id)
###### is_prayer_enabled(id)

## Shop

#### Shop Item Object

Contains the following fields: `id`, `amount`, `price`.

###### get_shop_item_count()
###### get_shop_item_at_index(index)
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
###### get_quest(id)
###### is_quest_complete(id)
###### get_quest_points()

## Raw Packet Operations

###### create_packet(opcode)
###### write_byte(int)
###### write_short(int)
###### write_int(int)
###### write_bytes(bytearray)
###### send_packet()