# Firepower by Space

# This script cuts trees and lights the logs with a tinderbox.
# Make sure you have an axe, sleeping bag, and tinderbox.

import time

LOGS      = 14
TINDERBOX = 166
TREES     = [0, 1]

init                 = True
start_firemaking_xp  = 0
start_woodcutting_xp = 0
start_time           = 0

def init_script():
    global start_firemaking_xp, start_woodcutting_xp, start_time

    start_woodcutting_xp = get_experience(8)
    start_firemaking_xp = get_experience(11)
    start_time = time.time()

def loop():
    global init
    
    if init:
        init_script()
        init = False

    if get_fatigue() > 99:
        use_sleeping_bag()
        return 2000

    logs = get_nearest_ground_item_by_id(LOGS)
    if logs != None \
        and logs.x == get_x() \
        and logs.z == get_z() \
        and not is_object_at(logs.x, logs.z):

        box = get_inventory_item_by_id(TINDERBOX)
        if box != None:
            use_item_on_ground_item(box, logs)
            return 700
    
    tree = get_nearest_object_by_id(ids=TREES)
    if tree != None:
        at_object(tree)
        return 700
    
    return 500

def xp_per_hour(gained_xp, start_time_seconds):
    elapsed_time = time.time() - start_time_seconds
    if elapsed_time == 0:
        return 0
    gained_per_second = gained_xp / elapsed_time
    return int(gained_per_second * 3600)

def on_progress_report():
    wc_gained = get_experience(8) - start_woodcutting_xp
    fm_gained = get_experience(11) - start_firemaking_xp
    return {"Woodcutting level": get_max_stat(8),
            "Woodcutting xp": get_experience(8),
            "Woodcutting xp/hr": xp_per_hour(wc_gained, start_time),
            "Firemaking level": get_max_stat(11),
            "Firemaking xp": get_experience(11),
            "Firemaking xp/hr": xp_per_hour(fm_gained, start_time)}