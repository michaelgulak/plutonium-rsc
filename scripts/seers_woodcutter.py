# Seers woodcutter by Space
#
# Start at seers with an axe and a sleeping bag in your inventory.
# Script settings should look like this:

# [script]
# name            = "seers_woodcutter.py"
# progress_report = "20m"
#
# [script.settings]
# tree_type = "yew"

# Where tree_type can be either "yew" or "magic".

import time

KEPT_ITEMS  = [1263, 87, 12, 88, 203, 204, 405]
LOGS        = [635, 636]

logs_banked   = 0
click_timeout = 0
move_x = -1
move_z = -1

if settings.tree_type != "yew" and settings.tree_type != "magic":
    raise RuntimeError("tree_type must be either yew or magic")

CHOPPING_YEWS = settings.tree_type == "yew"

def chop_yews():
    global click_timeout, move_x, move_z

    if not in_rect(526, 469, 13, 10):
        if in_radius_of(500, 454, 15):
            door = get_object_from_coords(500, 454)
            if door != None and door.id == 64:
                at_object(door)
                return 1300

        walk_path_to(518, 471)
        return 3000
    
    if move_x != -1:
        if at(move_x, move_z):
            move_x = -1
            move_z = -1
        else:
            walk_to(move_x, move_z)
            return 700

    if get_fatigue() > 99:
        use_sleeping_bag()
        return 5000

    if click_timeout != 0 and time.time() < click_timeout:
        return 250
    
    obj = get_nearest_object_by_id(309)
    if obj != None:
        at_object(obj)
        click_timeout = time.time() + 5
        return 700
    
    return 700

def chop_magics():
    global click_timeout, move_x, move_z

    if not in_rect(551, 481, 38, 17):
        if in_radius_of(500, 454, 15):
            door = get_object_from_coords(500, 454)
            if door != None and door.id == 64:
                at_object(door)
                return 1300

        walk_path_to(519, 491)
        return 3000
    
    if move_x != -1:
        if at(move_x, move_z):
            move_x = -1
            move_z = -1
        else:
            walk_to(move_x, move_z)
            return 700

    if get_fatigue() > 95:
        use_sleeping_bag()
        return 5000

    if click_timeout != 0 and time.time() < click_timeout:
        return 250
    
    obj = get_nearest_object_by_id(310)
    if obj != None:
        at_object(obj)
        click_timeout = time.time() + 5
        return 700
    
    if get_x() != 531 or get_z() != 487:
        walk_path_to(531, 487)
        return 3000

    return 700

def bank_next_item():
    global logs_banked

    item_to_bank = get_inventory_item_except(KEPT_ITEMS)

    if item_to_bank != None:
        inv_count = get_inventory_count_by_id(item_to_bank.id)
        if item_to_bank.id in LOGS:
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
        return 3000

def loop():
    if is_bank_open() or get_total_inventory_count() == 30:
        return bank()
    
    if CHOPPING_YEWS:
        return chop_yews()
    else:
        return chop_magics()

def get_adjacent_coord():
    if is_reachable(get_x()+1, get_z()):
        return (get_x()+1, get_z())
    elif is_reachable(get_x(), get_z()+1):
        return (get_x(), get_z()+1)
    elif is_reachable(get_x()-1, get_z()):
        return (get_x()-1, get_z())
    else:
        return (get_x(), get_z()-1)

def on_server_message(msg):
    global click_timeout, move_x, move_z

    if msg.startswith("You slip") or msg.startswith("You get"):
        click_timeout = 0
    elif msg.startswith("@cya@You have been standing"):
        move_x, move_z = get_adjacent_coord()

def on_progress_report():
    return {"Woodcutting Level": get_max_stat(8), \
            "Logs Banked":       logs_banked, \
            "Woodcutting xp":    get_experience(8)}