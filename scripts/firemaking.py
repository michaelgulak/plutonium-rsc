# Firemaking by O N I O N
# Inspired by Space
# With special thanks to chatgpt

# The following features:
# 1. moving to the closest tree, including using diagonal points for shorter pathing
# 2. ensures we don't move to & cut from a position where log drops on an object
# 3. sleeps for the duration of the path if path is long enough

# Make sure you have an axe, sleeping bag, and tinderbox.

############# WARNING ################

# There are no safety features for where you might end up!
# It is better to have nothing of value on you

######################################

# Tunables (higher = more cpu/ram usage)
MAX_TREES   = 10   # only consider nearest N trees
MAX_PATH    = 10   # don't accept paths longer than this (tiles/steps)
ASTAR_DEPTH = 16   # good depth for ~<=10-step paths (2^4)

#########################################

import time

CUT_TIMEOUT = 0
FIREMAKING_TIMEOUT = 0
ACTION_TIMEOUT = 0
SLEEP_TIMEOUT = 0

LOGS      = 14
TINDERBOX = 166
TREES     = [0, 1]

init                 = True
start_firemaking_xp  = 0
start_woodcutting_xp = 0
start_time           = 0

WAIT_TICKS = 0

class WalkTile:
    def __init__(self, x, z):
        self.x = x
        self.z = z

def init_script():
    global start_firemaking_xp, start_woodcutting_xp, start_time

    start_woodcutting_xp = get_experience(8)
    start_firemaking_xp = get_experience(11)
    start_time = time.time()

def on_server_tick(tick):
    global WAIT_TICKS
    if WAIT_TICKS > 0:
        WAIT_TICKS -= 1

def at_object_no_walk(obj):
    debug("ATOBJECTNOWALK (%d, %d)" % (obj.x, obj.z))
    create_packet(136)
    write_short(obj.x)
    write_short(obj.z)
    send_packet()

def on_sleepword_sent():
    global SLEEP_TIMEOUT
    set_wakeup_at(0)
    SLEEP_TIMEOUT = time.time() + 1
    return loop()

def loop():
    global init
    global CUT_TIMEOUT, FIREMAKING_TIMEOUT, WAIT_TICKS, ACTION_TIMEOUT, SLEEP_TIMEOUT
    
    if init:
        init_script()
        init = False

    if get_fatigue() == 100 and time.time() > SLEEP_TIMEOUT:
        use_sleeping_bag()
        return 2000

    if WAIT_TICKS > 0:
        return 17

    if time.time() < ACTION_TIMEOUT:
        return 17

    logs = get_nearest_ground_item_by_id(LOGS, x=get_x(),z=get_z(),radius=1)
    if logs != None and not is_object_at(logs.x, logs.z) and time.time() > FIREMAKING_TIMEOUT:
        tinderbox = get_inventory_item_by_id(TINDERBOX)
        if tinderbox != None:
            use_item_on_ground_item(tinderbox, logs)
            FIREMAKING_TIMEOUT = time.time() + 10
            ACTION_TIMEOUT = time.time() + 10
            return 17

    tree = get_nearest_object_by_id(ids=TREES)
    if tree != None and distance_to(tree.x,tree.z) == 1 and not is_object_at(get_x(),get_z()) and time.time() > CUT_TIMEOUT:
        at_object_no_walk(tree)
        CUT_TIMEOUT = time.time() + 10
        ACTION_TIMEOUT = time.time() + 10
        return 17

    walk = None
    nearest = get_nearest_tree()
    if nearest != None:
        tree = nearest["tree"]   # game object with .x / .z
        walk = nearest["walk"]   # dict with "x" / "z"
        log("Tree: " + str(tree.x) + "," + str(tree.z) + " | Walk: " + str(walk.x) + "," + str(walk.z))

    if walk != None:
        path = calculate_path_to(walk.x,walk.z,ASTAR_DEPTH)
        walk_to(walk.x,walk.z)
        if path.length() > 2 and get_fatigue() > 5:
            sleep_ticks = path.length() - 2
            current_fatigue = get_fatigue()
            wake_fatigue = current_fatigue - (sleep_ticks * 5)
            set_wakeup_at(wake_fatigue)
            use_sleeping_bag()
            return 1000
        WAIT_TICKS = path.length()
        return 17

    return 17

def get_nearest_tree():
    # 1) collect all trees, sort by cheap player distance, keep nearest N
    all_trees = [obj for obj in get_objects() if obj.id in TREES]
    all_trees.sort(key=lambda t: distance_to(t.x, t.z))
    tree_candidates = all_trees[:MAX_TREES]

    # 8 neighboring tiles around a tree
    neighbor_offsets = [
        ( 0, -1), ( 1, -1), ( 1,  0), ( 1,  1),
        ( 0,  1), (-1,  1), (-1,  0), (-1, -1),
    ]

    best_len = float("inf")
    best_tree = None
    best_walk = None

    for tree in tree_candidates:
        tx, tz = tree.x, tree.z

        for dx, dz in neighbor_offsets:
            wx, wz = tx + dx, tz + dz

            # Quick rejects
            if is_object_at(wx, wz):
                continue

            # Lower-bound from player to (wx,wz); if already worse than best, skip
            lb = distance_to(wx, wz)
            if lb > MAX_PATH:
                continue
            if lb >= best_len:
                continue

            # Bounded A* to keep CPU/memory in check
            # prefer setting a depth; skip_local=False so we don't route through blocks
            path = calculate_path_to(wx, wz, ASTAR_DEPTH, skip_local=False)
            if not path:
                continue

            plen = path.length()
            if plen > MAX_PATH:
                continue

            if plen < best_len:
                best_len = plen
                best_tree = tree
                best_walk = WalkTile(wx, wz)

                # if we ever find the theoretical best, we can stop early
                if best_len <= 1:
                    return {"tree": best_tree, "walk": best_walk}

    if best_tree is None:
        return None
    return {"tree": best_tree, "walk": best_walk}
            

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

def on_server_message(msg):
    global CUT_TIMEOUT, FIREMAKING_TIMEOUT, ACTION_TIMEOUT

    if msg == "You fail to light a fire":
        FIREMAKING_TIMEOUT = 0
        ACTION_TIMEOUT = 0
    elif msg == "The fire catches and the logs begin to burn":
        FIREMAKING_TIMEOUT = time.time() + 1
        ACTION_TIMEOUT = 0
    elif msg == "You slip and fail to hit the tree":
        CUT_TIMEOUT = 0
        ACTION_TIMEOUT = 0
    elif msg == "You get some wood":
        CUT_TIMEOUT = time.time() + 1
        ACTION_TIMEOUT = 0