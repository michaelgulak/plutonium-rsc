# Catherby/Ardougne walker by Space
#
# You can get to Catherby/Ardougne as a level 3 with this script.
# This script can be run with or without script settings. If you want 
# to walk to ardy then set stop_at_ardy = true.
#
# Script block should look like this:
#
# [script]
# name = "catherby_walker.py"
#
# [script.settings]
# stop_at_ardy = false

STATE_COLLECT_COINS      = 0
STATE_WALK_TO_PORT_SARIM = 1
STATE_TAKE_BOAT_1        = 2
STATE_PICK_BANANAS       = 3
STATE_WALK_TO_BRIM_BOAT  = 4
STATE_TAKE_BOAT_2        = 5
STATE_WALK_TO_CATHERBY_1 = 6
STATE_WALK_TO_CATHERBY_2 = 7

BOAT2_OPTIONS = ["Can I board this ship?", "Search away I have nothing to hide", "Ok"]

class BananaTree:
    def __init__(self, x, z):
        self.x = x
        self.z = z

state       = STATE_COLLECT_COINS
path        = None
banana_tree = None

def collect_coins():
    global state, path

    if in_combat():
        walk_to(get_x(), get_z())
        return 600

    if get_inventory_count_by_id(10) >= 60:
        state = STATE_WALK_TO_PORT_SARIM
        return 100

    if path != None:
        path.process()
        if not path.complete():
            if not path.walk():
                path = calculate_path_to(124, 647)
                if path == None:
                    log("Could not path to Lumbridge 2, stopping")
                    stop_account()

            return 1000
        else:
            path = None

    if not in_radius_of(120, 648, 15):
        path = calculate_path_to(124, 647)
        if path == None:
            log("Could not path to Lumbridge, stopping")
            stop_account()
            return 1000

        return 100
    
    coins = get_nearest_ground_item_by_id(10, reachable=True, x=120, z=648, radius=15)
    if coins != None:
        pickup_item(coins)
        return 700
        
    man = get_nearest_npc_by_id(11, in_combat=False, reachable=True, x=120, z=648, radius=15)
    if man != None:
        thieve_npc(man)
        return 700
    
    debug("Waiting for man to be in range to thieve for gold")
    
    return 1000

def walk_to_port_sarim():
    global state, path

    if path != None:
        path.process()
        if not path.complete():
            if not path.walk():
                path = calculate_path_to(269, 651)
                if path == None:
                    log("Could not path to Port Sarim 2, stopping")
                    stop_account()

            return 1000
        else:
            path = None

    if not in_radius_of(269, 651, 10):
        path = calculate_path_to(269, 651)
        if path == None:
            log("Could not path to Port Sarim, stopping")
            stop_account()
            return 1000
    else:
        state = STATE_TAKE_BOAT_1
        return 100
    
    return 1000

def take_boat_1():
    global state

    if in_radius_of(324, 713, 10):
        state = STATE_PICK_BANANAS
        return 100

    if is_option_menu():
        answer(0 if is_quest_complete(16) else 1)
        return 1000
    
    npc = get_nearest_npc_by_id(166, talking=False)
    if npc != None:
        talk_to_npc(npc)
        return 2000
    
    return 1000

def pick_next_tree(tree):
    global banana_tree

    if tree == None:
        banana_tree = BananaTree(344, 707)
    else:
        if tree.x == 360 and tree.z == 703:
            banana_tree = BananaTree(344, 707)
        else:
            nx = tree.x + ((4 if tree.x == 344 else 3) if tree.z == 703 else 0)
            nz = 707 if tree.z == 703 else tree.z - 2
            banana_tree = BananaTree(nx, nz)

def pick_bananas():
    global state

    if get_current_stat(3) < get_max_stat(3):
        banana = get_inventory_item_by_id(249)
        if banana != None:
            use_item(banana)
            return 1000
    else:
        if get_max_stat(3) >= 22 or get_inventory_count_by_id(249) >= 6:
            state = STATE_WALK_TO_BRIM_BOAT
            return 100
    
    if not in_rect(365, 702, 24, 7):
        walk_path_to(346, 708)
        return 3000
    
    if banana_tree == None:
        pick_next_tree(None)
    
    if banana_tree != None:
        tree = get_object_from_coords(banana_tree.x, banana_tree.z)
        if tree != None:
            at_object(tree)
            return 800
    
    return 1000

def walk_to_brim_boat():
    global state, path

    if not in_combat() and get_max_stat(3) < 22 and get_max_stat(3) - get_current_stat(3) >= 2:
        banana = get_inventory_item_by_id(249)
        if banana != None:
            use_item(banana)
            return 1000

    if path != None:
        path.process()
        if in_radius_of(434, 682, 15) and get_x() <= 434 and path.next_x() >= 435:
            if in_combat():
                walk_to(get_x(), get_z())
                return 300

            gate = get_object_from_coords(434, 682)
            if gate != None and gate.id == 254:
                at_object(gate)
            return 1000
                
        if not path.complete():
            path.walk()
            return 600
        else:
            path = None

    if not in_radius_of(467, 657, 10):
        path = calculate_path_to(467, 657)
        if path == None:
            log("Could not path to Brimhaven dock, stopping")
            stop_account()
            return 1000
    else:
        state = STATE_TAKE_BOAT_2
        return 100
    
    return 1000

def take_boat_2():
    global state

    if in_radius_of(538, 617, 10):
        if hasattr(settings, "stop_at_ardy") and settings.stop_at_ardy:
            log("Made it to Ardougne")
            stop_account()
            return 1000
            
        state = STATE_WALK_TO_CATHERBY_1
        return 100

    if is_option_menu():
        for opt in BOAT2_OPTIONS:
            idx = get_option_menu_index(opt)
            if idx != -1:
                answer(idx)
                return 1200
    
    npc = get_nearest_npc_by_id(317, talking=False)
    if npc != None:
        talk_to_npc(npc)
        return 2000
    
    return 1000

def walk_to_catherby_1():
    global state, path

    if path != None:
        path.process()
        if not path.complete():
            path.walk()
            return 1000
        else:
            path = None

    if not in_radius_of(555, 480, 10):
        path = calculate_path_to(555, 480)
        if path == None:
            log("Could not path to Catherby 1, stopping")
            stop_account()
            return 1000
    else:
        state = STATE_WALK_TO_CATHERBY_2
        return 100
    
    return 1000

def walk_to_catherby_2():
    global path

    if path != None:
        path.process()
        if not path.complete():
            path.walk()
            return 1000
        else:
            path = None

    if not in_radius_of(438, 481, 10):
        path = calculate_path_to(438, 481)
        if path == None:
            log("Could not path to Catherby 2, stopping")
            stop_account()
            return 1000
    else:
        log("Made it to Catherby")
        stop_account()
        return 1000
    
    return 1000
    
def loop():
    if state == STATE_COLLECT_COINS:
        return collect_coins()
    elif state == STATE_WALK_TO_PORT_SARIM:
        return walk_to_port_sarim()
    elif state == STATE_TAKE_BOAT_1:
        return take_boat_1()
    elif state == STATE_PICK_BANANAS:
        return pick_bananas()
    elif state == STATE_WALK_TO_BRIM_BOAT:
        return walk_to_brim_boat()
    elif state == STATE_TAKE_BOAT_2:
        return take_boat_2()
    elif state == STATE_WALK_TO_CATHERBY_1:
        return walk_to_catherby_1()
    elif state == STATE_WALK_TO_CATHERBY_2:
        return walk_to_catherby_2()

def on_server_message(msg):
    if msg.endswith("last banana") or msg.startswith("there are no bananas left"):
        pick_next_tree(banana_tree)
    
def on_death():
    global state

    state = STATE_COLLECT_COINS