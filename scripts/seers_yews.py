# Seers Yews by RLN
#
# Go near the yew trees at seers with an axe and sleeping bag.
# No script settings required for this script.

import time

KEPT_ITEMS  = [1263, 87, 12, 88, 203, 204, 405]
YEW_LOGS    = 635

logs_banked   = 0
click_timeout = 0

def chop():
    global click_timeout

    if not in_rect(526, 469, 13, 10):
        if in_radius_of(500, 454, 15):
            door = get_object_from_coords(500, 454)
            if door != None and door.id == 64:
                at_object(door)
                return 1300

        walk_path_to(518, 471)
        return 5000
    
    if get_fatigue() > 95:
        use_sleeping_bag()
        return 5000

    if click_timeout != 0 \
        and time.time() <= click_timeout:
        return 250
    
    obj = get_nearest_object_by_id(309)
    if obj != None:
        at_object(obj)
        click_timeout = time.time() + 5
        return 700
    
    return 700

def bank_next_item():
    global logs_banked

    item_to_bank = get_inventory_item_except(KEPT_ITEMS)

    if item_to_bank != None:
        inv_count = get_inventory_count_by_id(item_to_bank.id)
        if item_to_bank.id == YEW_LOGS:
            logs_banked = get_bank_count(item_to_bank.id) + inv_count
        deposit(item_to_bank.id, inv_count)

    return 2000

def bank():
    if in_rect(504, 447, 7, 7): # in bank
        if not is_bank_open():
            return open_bank()
        else:
            if get_total_inventory_count() != 2:
                return bank_next_item()
            else:
                close_bank()
                return 1000
    else:
        if in_radius_of(500, 454, 15):
            door = get_object_from_coords(500, 454)
            if door != None and door.id == 64:
                at_object(door)
                return 1300

        walk_path_to(501, 451)
        return 5000

def loop():
    if is_bank_open() \
        or get_total_inventory_count() == 30:

        return bank()
    
    return chop()

def on_server_message(msg):
    global click_timeout

    if msg.startswith("You slip") or msg.startswith("You get"):
        click_timeout = 0

def on_progress_report():
    return {"Woodcutting Level": get_max_stat(8), \
            "Logs Banked":       logs_banked}