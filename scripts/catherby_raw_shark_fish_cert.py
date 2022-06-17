import time

# Catherby raw shark fisher & certer by RLN

# No script settings required.

SPOT_X = 406
SPOT_Z = 505
MENU_OPTIONS = ["I have some fish to trade in", "Raw shark", "Twentyfive"]

move_timer = False
fish_timeout = 0

def cert():
    if not in_rect(430, 481, 4, 5):
        walk_path_to(427, 485)
        return 3000
    else:
        if is_option_menu():
            for opt in MENU_OPTIONS:
                idx = get_option_menu_index(opt)
                if idx != -1:
                    answer(idx)
                    return 1200
        else:
            owen = get_nearest_npc_by_id(299)
            if owen != None:
                talk_to_npc(owen)
                return 1200

    return 5000

def fish():
    global fish_timeout, move_timer

    if move_timer:
        if get_x() != SPOT_X or get_z() != SPOT_Z - 2:
            walk_to(SPOT_X, SPOT_Z - 2)
            return 600
        else:
            move_timer = False

    if get_x() != SPOT_X or get_z() != SPOT_Z - 1:
        walk_path_to(SPOT_X, SPOT_Z - 1)
        return 5000

    if fish_timeout != 0 and time.time() <= fish_timeout:
        return 250
    
    fish = get_object_from_coords(SPOT_X, SPOT_Z)

    if fish != None:
        at_object2(fish)
        fish_timeout = time.time() + 5
        return 900

    return 5000  

def loop():
    if get_fatigue() > 95:
        use_sleeping_bag()
        return 1000

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