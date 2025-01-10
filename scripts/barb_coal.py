# Barbarian Village coal miner by Space
#
# Must have sleeping bag and pickaxe in inventory.
#
# No settings needed

import time

KEPT_ITEMS = [1263, 1262, 1261, 1260, 1259, 1258, 156]
COAL       = 155

click_timeout = 0
coal_banked   = 0

def bank_next_item():
    global coal_banked

    item_to_bank = get_inventory_item_except(KEPT_ITEMS)

    if item_to_bank != None:
        inv_count = get_inventory_count_by_id(item_to_bank.id)
        if item_to_bank.id == COAL:
            coal_banked = get_bank_count(item_to_bank.id) + inv_count
        deposit(item_to_bank.id, inv_count)

    return 2000

def bank():
    if in_rect(220, 448, 9, 6): # in bank
        if not is_bank_open():
            return open_bank()
        else:
            if get_total_inventory_count() != 2:
                return bank_next_item()
            else:
                close_bank()
                return 1000
    else:
        if in_radius_of(217, 447, 15):
            door = get_object_from_coords(217, 447)
            if door != None and door.id == 64:
                at_object(door)
                return 1300
        
        walk_path_to(218, 448)
        return 5000

def mine():
    global click_timeout

    if get_fatigue() > 99:
        use_sleeping_bag()
        return 3000
    
    if not in_rect(230, 503, 8, 5):
        if in_radius_of(217, 447, 15):
            door = get_object_from_coords(217, 447)
            if door != None and door.id == 64:
                at_object(door)
                return 1300

        walk_path_to(226, 504)
        return 5000
    
    if click_timeout != 0 and time.time() < click_timeout:
        return 250
    
    obj = get_nearest_object_by_id(110)
    if obj != None:
        at_object(obj)
        click_timeout = time.time() + 5
        return 700
    
    return 700

def loop():
    if is_bank_open() or get_total_inventory_count() == 30:
        return bank()
    
    return mine()

def on_server_message(msg):
    global click_timeout

    if msg.startswith("You only") or msg.startswith("There is"):
        click_timeout = 0
    elif msg.startswith("You manage") or msg.startswith("You just"):
        click_timeout = 0

def on_progress_report():
    return {"Mining Level": get_max_stat(14),
            "Coal Banked":  coal_banked}