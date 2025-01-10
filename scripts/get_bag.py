# Get sleeping bag script by Space

# Pickpockets men in Lumbridge until it gets enough gold for a sleeping bag
# and then buys one.

# If there are no men in a narrow area it will wait until there is one.

MAN          = 11
SLEEPING_BAG = 1263
MEAT         = 132
SPAWN_PT     = 120, 648
SHOP_NPCS    = [83, 55]
SHOP_PT      = (133, 643)

def loop():
    if in_combat():
        walk_to(get_x(), get_z())
        return 650

    meat = get_inventory_item_by_id(MEAT)
    if meat != None:
        use_item(meat)
        return 1300

    if has_inventory_item(SLEEPING_BAG):
        log("finished objective")
        stop_account()
        return 10000
    
    coin_count = get_inventory_count_by_id(10)
    log("# Coins = %d" % coin_count)
    if coin_count >= 39:
        if in_radius_of(132, 641, 15):
            door = get_wall_object_from_coords(132, 641)
            if door != None and door.id == 2:
                at_wall_object(door)
                return 1300

        if not in_radius_of(SHOP_PT[0],SHOP_PT[1],8):
            walk_path_to(SHOP_PT[0],SHOP_PT[1])
            return 650

        if not is_shop_open():
            if is_option_menu():
                answer(0)
                return 3000
            else:
                npc = get_nearest_npc_by_id(ids=SHOP_NPCS, talking=False, reachable=True)
                if npc != None:
                    talk_to_npc(npc)
                    return 3000
        else:
            buy_shop_item(SLEEPING_BAG, 1)
            return 3000
    else:
        if not in_radius_of(SPAWN_PT[0], SPAWN_PT[1], 15):
            walk_path_to(120, 648)
            return 5000
    
        coins = get_nearest_ground_item_by_id(10, reachable=True, x=SPAWN_PT[0], z=SPAWN_PT[1], radius=15)
        if coins != None:
            pickup_item(coins)
            return 700
        
        man = get_nearest_npc_by_id(MAN, in_combat=False, reachable=True, x=SPAWN_PT[0], z=SPAWN_PT[1], radius=15)
        if man != None:
            thieve_npc(man)
            return 700
    
    log("waiting for a man to be in area")
    return 700

