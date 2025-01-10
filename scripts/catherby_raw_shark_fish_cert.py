# Catherby raw shark fisher & certer by Space

# Start with necessary equipment and sleeping bag.

# No script settings required.

import time

SPOT_X = 406
SPOT_Z = 505
MENU_OPTIONS = ["I have some fish to trade in", "Raw shark", "Twentyfive"]
KEPT_ITEMS = [1263, 379, 631, 545]

move_timer = False
fish_timeout = 0
talk_timeout = 0

def cert():
    global talk_timeout

    if in_radius_of(427, 485, 10):
        door = get_wall_object_from_coords(427, 485)
        if door != None and door.id == 2:
            at_wall_object(door)
            return 700
            
    if not in_rect(430, 481, 4, 5):
        walk_path_to(427, 485)
        return 3000
    else:
        if is_option_menu():
            for opt in MENU_OPTIONS:
                idx = get_option_menu_index(opt)
                if idx != -1:
                    answer(idx)
                    talk_timeout = time.time() + 5
                    return 2000
        else:
            if talk_timeout != 0 and time.time() < talk_timeout:
                return 250

            owen = get_nearest_npc_by_id(299)
            if owen != None:
                talk_to_npc(owen)
                talk_timeout = 0
                return 400

    return 5000

def fish():
    global fish_timeout, move_timer

    if move_timer:
        if not at(SPOT_X, SPOT_Z - 2):
            walk_to(SPOT_X, SPOT_Z - 2)
            return 600
        else:
            move_timer = False

    if not at(SPOT_X, SPOT_Z):
        if in_radius_of(427, 485, 10):
            door = get_wall_object_from_coords(427, 485)
            if door != None and door.id == 2:
                at_wall_object(door)
                return 700

        walk_path_to(SPOT_X, SPOT_Z - 1)
        return 5000

    if get_fatigue() > 99:
        use_sleeping_bag()
        return 1000

    if fish_timeout != 0 and time.time() < fish_timeout:
        return 250
    
    item_to_drop = get_inventory_item_except(KEPT_ITEMS)
    if item_to_drop != None:
        drop_item(item_to_drop)
        return 1200
    
    fish = get_object_from_coords(SPOT_X, SPOT_Z)

    if fish != None:
        at_object2(fish)
        fish_timeout = time.time() + 5
        return 900

    return 5000  

def loop():
    if get_inventory_count_by_id(545) >= 25:
        return cert()
    
    return fish()

def on_server_message(msg):
    global fish_timeout, move_timer

    if msg.startswith("You fail") or msg.startswith("You catch"):
        fish_timeout = 0
    elif msg.startswith("@cya@You have been standing"):
        move_timer = True

def on_progress_report():
    return {"Raw shark certs": get_inventory_count_by_id(631)}