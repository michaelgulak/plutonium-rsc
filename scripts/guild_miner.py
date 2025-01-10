# Mining guild script by Space that mines mith ore and coal.

# Make sure you start it with a pickaxe and sleeping bag.

COAL       = 155
MITH       = 153
COAL_ROCKS = [110, 111]
KEPT_ITEMS = [1263, 1262, 1261, 1260, 1259, 1258, 156]

coal_banked = 0
mith_banked = 0

def bank_next_item():
    global coal_banked, mith_banked
 
    item_to_bank = get_inventory_item_except(KEPT_ITEMS)
 
    if item_to_bank != None:
        inv_count = get_inventory_count_by_id(item_to_bank.id)
        if item_to_bank.id == COAL:
            coal_banked = get_bank_count(item_to_bank.id) + inv_count
        elif item_to_bank.id == MITH:
            mith_banked = get_bank_count(item_to_bank.id) + inv_count

        deposit(item_to_bank.id, inv_count)
    
    return 2000
 
 
def bank():
    if in_rect(286, 564, 7, 10): # in bank
        if not is_bank_open():
            return open_bank()
        else:
            if get_total_inventory_count() != 2:
                return bank_next_item()
            else:
                close_bank()
                return 1000
    else:
        if get_z() > 3000:
            ladder = get_object_from_coords(274, 3398)
            if ladder != None:
                at_object(ladder)
                return 700
        else:
            if in_radius_of(274, 563, 15):
                door = get_wall_object_from_coords(274, 563)
                if door != None and door.id == 2:
                    at_wall_object(door)
                    return 1300
            
            if in_radius_of(287, 571, 15):
                door = get_object_from_coords(287, 571)
                if door != None and door.id == 64:
                    at_object(door)
                    return 1300

            walk_path_to(286, 571)
            return 1000

    return 1000
 
def mine():
    if get_fatigue() > 99:
        use_sleeping_bag()
        return 3000

    if not in_rect(277, 3381, 14, 19):
        if not in_rect(277, 563, 6, 5):
            if in_radius_of(274, 563, 15):
                door = get_wall_object_from_coords(274, 563)
                if door != None and door.id == 2:
                    at_wall_object(door)
                    return 1300

            if in_radius_of(287, 571, 15):
                door = get_object_from_coords(287, 571)
                if door != None and door.id == 64:
                    at_object(door)
                    return 1300

            walk_path_to(274,565)
            return 1000
        else:
            ladder = get_object_from_coords(274, 566)
            if ladder != None:
                at_object(ladder)
                return 1000
        
        return 1000
          
    obj = get_nearest_object_by_id_in_rect(107, x=277, z=3381, width=14, height=19)
    if obj != None:
        at_object(obj)
        return 700
    
    obj = get_nearest_object_by_id_in_rect(ids=COAL_ROCKS, x=277, z=3381, width=14, height=19)
    if obj != None:
        at_object(obj)
        return 700 

    return 1000
 
def loop():
    if is_bank_open() or get_total_inventory_count() == 30:
        return bank()
        
    return mine()

def on_progress_report():
    return {"Mining Level": get_max_stat(14), \
            "Coal Banked":  coal_banked, \
            "Mith Banked":  mith_banked} 
       
