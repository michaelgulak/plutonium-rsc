HATCHET   = 87
TINDERBOX = 166
COINS     = 10
PICK      = 156

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
    
    if not has_inventory_item(PICK):
        if not in_radius_of(230, 510, 5):
            walk_path_to(230, 510)
            return 5000
        else:
            pick = get_nearest_ground_item_by_id(PICK)
            if pick != None:
                pickup_item(pick)
                return 2000
    else:
        stop_script()
        set_autologin(False)
        logout()
        print("objective complete")
        return 10000