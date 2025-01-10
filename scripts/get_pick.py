# Get pick by Space

# Walks to Barbarian Village to grab a bronze pickaxe.

HATCHET   = 87
TINDERBOX = 166
COINS     = 10
PICK      = 156
MEAT      = 132

def loop():
    hatchet = get_inventory_item_by_id(HATCHET)
    if hatchet != None:
        drop_item(hatchet)
        return 1000

    tinderbox = get_inventory_item_by_id(TINDERBOX)
    if tinderbox != None:
        drop_item(tinderbox)
        return 1000
    
    coins = get_inventory_item_by_id(COINS)
    if coins != None:
        drop_item(coins)
        return 1000

    meat = get_inventory_item_by_id(MEAT)
    if meat != None:
        use_item(meat)
        return 1300
    
    if not has_inventory_item(PICK):
        if in_radius_of(230, 511, 15):
            door = get_wall_object_from_coords(230, 511)
            if door != None and door.id == 2:
                at_wall_object(door)
                return 1300

        if not in_radius_of(230, 510, 5):
            walk_path_to(230, 510)
            return 5000
        else:
            pick = get_nearest_ground_item_by_id(PICK)
            if pick != None:
                pickup_item(pick)
                return 2000
    else:
        log("objective complete")
        stop_account()
        return 10000
    
    return 2000