# Fishing guild raw shark fisher & certer by Millz (Adapted from RLN's Catherby)

# No script settings required.

# Will walk to, and enter the fishing guild. Doesn't handle gates, but moves from catherby etc.

import time
import millz_common

SPOT_X = 593
SPOT_Z = 501
MENU_OPTIONS = ["I have some fish to trade in", "Raw shark", "Twentyfive"]
KEPT_ITEMS = [1263, 379, 631, 545]

move_timer = False
fish_timeout = 0
talk_timeout = 0

path = None
destination = [0, 0]

init = False
start_time = time.time()
start_fish_xp = 0

def enter_fishing_guild():
    global path
    
    if path != None:
        path.process()
        if not path.complete():
            if not path.walk() and destination[0] != 0:
                path = calculate_path_to(destination[0], destination[1])
                if path == None:
                    stop_script()
                    set_autologin(False)
                    logout()
                    log("Could not path to " + str(destination[0]) + ", " + str(destination[1]) + ". Stopping.")
            return 1000
        else:
            path = None
    
    guild_door = get_wall_object_from_coords(586, 524)
    if guild_door != None and guild_door.id == 112:
        log("Opening fishing guild door")
        at_wall_object(guild_door)
        return 1000
    
    if distance_to(586, 524) > 5:
        log("We're not in the fishing guild. Attempting to walk there (no gate logic so ensure it's walkable!)")
        destination = [586, 524]
        path = calculate_path_to(586, 524)
        if path == None:
            log("Failed to path to fishing guild")
            stop_script()
        return 1000
        
    return 1000

def cert():
    global talk_timeout

    if is_option_menu():
        #log("Certing")
        for opt in MENU_OPTIONS:
            idx = get_option_menu_index(opt)
            if idx != -1:
                answer(idx)
                talk_timeout = time.time() + 5
                return 2000
    else:
        if talk_timeout != 0 and time.time() < talk_timeout:
            return 250

        padik = get_nearest_npc_by_id(370)
        if padik != None:
            #log("Talk to Padik")
            talk_to_npc(padik)
            talk_timeout = 0
            return 400
        else:
            #log("Walk to Padik")
            walk_path_to(601, 506)
            return 600

    return 5000

def fish():
    global fish_timeout, move_timer

    if move_timer:
        #log("Moving 5min timer")
        if get_x() != SPOT_X or get_z() != SPOT_Z + 2:
            walk_to(SPOT_X, SPOT_Z + 2)
            return 600
        else:
            move_timer = False

    if get_fatigue() > 95:
        use_sleeping_bag()
        return 1000

    if fish_timeout != 0 and time.time() <= fish_timeout:
        return 250
    
    item_to_drop = get_inventory_item_except(KEPT_ITEMS)
    if item_to_drop != None:
        drop_item(item_to_drop)
        return 1200
        
    if get_x() != SPOT_X or get_z() != SPOT_Z + 1:
        #log("Walking to fishing spot")
        walk_path_to(SPOT_X, SPOT_Z + 1)
        return 600
    
    fish = get_object_from_coords(SPOT_X, SPOT_Z)

    if fish != None:
        #log("Fishing")
        at_object2(fish)
        fish_timeout = time.time() + 5
        return 900

    return 5000  

def loop():
    global init, start_fish_xp
    
    if init is False:
        init = True
        start_fish_xp = get_experience(10)
        
    if not millz_common.in_fishing_guild(get_x(), get_z()):
        return enter_fishing_guild()
    
    if in_radius_of(603, 506, 8):
        door = get_wall_object_from_coords(603, 506)
        if door != None and door.id == 2:
            at_wall_object(door)
            return 700
            
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
    gained_fish_xp = get_experience(10) - start_fish_xp
    
    return {"Raw shark certs": get_inventory_count_by_id(631),
            "Fishing Level": get_max_stat(10), \
            "Fishing XP Next Level:": millz_common.exp_until_next_level(get_experience(10)), \
            "Fishing XP/HR": millz_common.xp_per_hour(gained_fish_xp, start_time) 
            }
