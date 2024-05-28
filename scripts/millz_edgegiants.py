# Script block should look like this:
# Fight Modes: 0 = Controlled, 1 = Strength, 2 = Attack, 3 = Defense
# Loot ID 413 for big bones.

# [script]
# name = "millz_edgegiants.py"

# [script.settings]
# fight_mode = 0
# loot_ids = [31,32,33,34,35,36,37,38,40,41,42,46,413,526,527,1276,1277,220,436,438,439,440,441,442,443,1092,157,158,159,160,1289]
# eat = false
# food_count = 0
# bury_bones = false

import millz_common
import time

DEBUG = False
INIT = False

GIANT_ID = 61
BRASS_KEY = 99
HUT_DOOR_ID = 23 # Wall Object

start_time = time.time()

start_attack_xp = 0
start_defense_xp = 0
start_strength_xp = 0
start_hits_xp = 0

start_attack_level = 0
start_defense_level = 0
start_strength_level = 0
start_hits_level = 0

bank_bone_count = 0

loot = {}
spawn_map = {}
path = None
destination = [0, 0]


def debug_log(msg):
    global DEBUG

    if DEBUG is False:
        return 0
    
    log(msg)
    return 0

def is_inside_hut():
    return in_rect(204, 481, 4, 4)

def open_hut_door():
    debug_log("Open Hut Door?")
    hut_door = get_wall_object_from_coords(202, 485)
    if hut_door != None and hut_door.id == HUT_DOOR_ID:
        debug_log("Opening Door...")
        use_item_on_wall_object(get_inventory_item_by_id(BRASS_KEY), hut_door)
    return 2500

def is_dungeon_gate_closed():
    gate = get_object_from_coords(208, 3317)
    return gate != None and gate.id == 57

def open_dungeon_gate():
    gate = get_object_from_coords(208, 3317)
    if gate != None and gate.id == 57:
        debug_log("Open Dungeon Gate")
        at_object(gate)
        return 1000

def is_inside_bank():
    return in_rect(153, 498, 7, 10) # area_x, area_z, width, height

def need_to_bank():
    if settings.eat is True and millz_common.food_count() == 0:
        return True
    
    if get_inventory_count_by_id(ids=settings.loot_ids) > 0:
        return True
    
    return False

def fight():
    # Called when inventory isn't full, have food, and below ground.
    if get_combat_style() != settings.fight_mode:
        debug_log("Set fight mode")
        set_combat_style(settings.fight_mode)
        return 1000
    
    if settings.eat is True and millz_common.use_food():
        return 1500
    
    if in_combat():
        return 500
    
    if get_fatigue() > 95:
        debug_log("Sleep")
        use_sleeping_bag()
        return 1000
    
    if settings.bury_bones is True:
        big_bones = get_inventory_item_by_id(413)
        if big_bones != None:   
            debug_log("Bury Bones")   
            use_item(big_bones)
            return 1000

    ground_item = get_nearest_ground_item_by_id(ids=settings.loot_ids, reachable=True)
    if ground_item != None:
        debug_log("Pickup Item")
        pickup_item(ground_item)
        return 1000

    npc = get_nearest_npc_by_id(GIANT_ID, in_combat=False, reachable=True)
    if npc != None:
        debug_log("Attack NPC")
        attack_npc(npc)
        return 500

    next_spawn = get_next_spawn()
    if next_spawn is None:
        debug_log("Spawn is none, waiting")
        return 500
    
    if distance_to(next_spawn.get_coord()[0], next_spawn.get_coord()[1]) > 0:
        debug_log("Walk to next spawn")
        walk_to(next_spawn.get_coord()[0], next_spawn.get_coord()[1])
    
    debug_log("Waiting")
    return 500

def banking():
    # Above ground

    global spawn_map, loot, destination, path, bank_bone_count

    bank_door = get_object_from_coords(150, 507)
    if bank_door != None and bank_door.id == 2:
        at_object(bank_door)
        return 1000


    if is_inside_hut():
        if need_to_bank():
            return open_hut_door()
        
        ladder_down = get_object_from_coords(203, 482)
        if ladder_down != None:
            debug_log("Go Down Ladder")
            at_object(ladder_down)
            return 1000
    
    banker = get_nearest_npc_by_id(95)
    if is_inside_bank() or banker != None:
        if is_bank_open():
            debug_log("Bank Open")
            loot_item = get_inventory_item_by_id(ids=settings.loot_ids)
            if loot_item != None:
                count_in_inventory = get_inventory_count_by_id(loot_item.id)
                loot[str(loot_item.id)] = loot.get(str(loot_item.id), 0) + count_in_inventory
                deposit(loot_item.id, count_in_inventory)
                return 1500

            if settings.eat is True and settings.food_count > 0:
                if millz_common.withdraw_food(settings.food_count, True):
                    return 2000
                
            
            bank_bone_count = get_bank_count(413)
            close_bank()

        if need_to_bank() and not is_bank_open():
            debug_log("Open Bank")
            return open_bank()
        
        debug_log("Set Path to Hut")
        destination = [202, 486]
        path = calculate_path_to(destination[0], destination[1])
        if path == None:
            log("Failed to path to hut")
            stop_script()
            set_autologin(False)
            logout()
            return 1000
        
    if not is_inside_hut() and need_to_bank():
        debug_log("Set Path to Bank")
        destination = [154, 508]
        path = calculate_path_to(destination[0], destination[1])
        if path == None:
            log("Failed to path to bank")
            stop_script()
            set_autologin(False)
            logout()
            return 1000
    
    if not is_inside_hut() and not need_to_bank():
        return open_hut_door()

    return 650

def leave_dungeon():
    # Underground, inventory is full or we need food.
    ladder_up = get_object_from_coords(203, 3314)
    if ladder_up != None:
        debug_log("Ladder Up")
        at_object(ladder_up)
    else:
        debug_log("Walk to Ladder")
        walk_to(204, 3314)

    return 1000



def loop():
    global INIT, start_attack_xp, start_defense_xp, start_strength_xp, start_hits_xp
    global start_attack_level, start_defense_level, start_strength_level, start_hits_level, loot, spawn_map, path, destination

    if INIT is False:
        debug_log("Init")

        brass_key = get_inventory_item_by_id(BRASS_KEY)
        if brass_key == None:      
            log("No Brass Key in Inventory! Stopping.")
            stop_script()
            set_autologin(False)
            logout()
            return 1000

        start_attack_xp = get_experience(0)
        start_defense_xp = get_experience(1)
        start_strength_xp = get_experience(2)
        start_hits_xp = get_experience(3)
        
        start_attack_level = get_max_stat(0)
        start_defense_level = get_max_stat(1)
        start_strength_level = get_max_stat(2)
        start_hits_level = get_max_stat(3)
        
        INIT = True

    if get_hp_percent() < 20:
        log("Critically low on health, pausing")
        return 5000
    
    if get_z() > 3000:
        if not in_combat() and is_dungeon_gate_closed():
            return open_dungeon_gate()

        if get_total_inventory_count() == 30 or (settings.eat is True and millz_common.food_count() == 0):
            debug_log("Leave Dungeon")
            return leave_dungeon()
        
        debug_log("Fight")
        return fight()
    

    if path != None:
        debug_log("Pathing")
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

    debug_log("Banking")
    return banking()


## Spawn Camp Logic

class Spawn:
    def __init__(self, server_index, x, z, timestamp):
        self.server_index = server_index
        self.x = x
        self.z = z
        self.timestamp = timestamp
    
    def get_index(self):
        return self.server_index
        
    def get_coord(self):
        return [self.x, self.z]
        
    def get_timestamp(self):
        return self.timestamp
        
    def set_timestamp(self, timestamp):
        self.timestamp = timestamp

def on_npc_spawned(npc):
    global spawn_map
    if npc.id == GIANT_ID:
        spawn_map[str(npc.sid)] = Spawn(npc.sid, npc.x, npc.z, time.time())
        # log("Spawn map now contains " + str(len(spawn_map)) + " entries")

def on_npc_despawned(npc):
    global spawn_map
    if npc.id == GIANT_ID:
        if str(npc.sid) in spawn_map:
            spawn_map[str(npc.sid)].set_timestamp(time.time())

# returns None or Spawn
def get_next_spawn():
    global spawn_map

    if not spawn_map or len(spawn_map) == 0:
        return None
        
    # Remove any spawns from the map that haven't spawned within the past 5 minutes.
    for key in list(spawn_map.keys()):
        spawn = spawn_map[key]
        if (time.time() - spawn.get_timestamp()) > 300:
            del spawn_map[key]  # Remove the entry from the dictionary
    
    # This check again to ensure we haven't wiped out all entries.
    if not spawn_map or len(spawn_map) == 0:
        return None
    
    next_spawn = min(spawn_map.values(), key=lambda spawn: spawn.get_timestamp())
    #x = next_spawn.get_coord()[0]
    #z = next_spawn.get_coord()[1]
    #distance = distance_to(x, z)
    #if distance > 0:
    #    log("Next spawn (" + str(x) + "," + str(z) + ") - Distance: " + str(distance) + "yds")
    
    return next_spawn

## End Spawn Camp Logic
    
def on_server_message(msg):
    global start_time, start_attack_xp, start_defense_xp, start_strength_xp, start_hits_xp
    
    if "advanced" in msg:
        start_time = time.time()
        start_attack_xp = get_experience(0)
        start_defense_xp = get_experience(1)
        start_strength_xp = get_experience(2)
        start_hits_xp = get_experience(3)


def on_death():
    log("You have died! Stopping script.")
    stop_account()


def on_progress_report():
    atk_gained = get_experience(0) - start_attack_xp
    def_gained = get_experience(1) - start_defense_xp
    str_gained = get_experience(2) - start_strength_xp
    hp_gained = get_experience(3) - start_hits_xp
    
    atk_level = get_max_stat(0) - start_attack_level
    def_level = get_max_stat(1) - start_defense_level
    str_level = get_max_stat(2) - start_strength_level
    hp_level = get_max_stat(3) - start_hits_level
    
    current_stats = str(get_max_stat(0)) + "-" + str(get_max_stat(1)) + "-" + str(get_max_stat(2)) + "-" + str(get_max_stat(3))
    
    status_report = {}
    
    full_report = {"Current Stats": current_stats,
            "Attack Levels Gained": atk_level, \
            "Attack XP/HR": millz_common.xp_per_hour(atk_gained, start_time), \
            "Defense Levels Gained": def_level, \
            "Defense XP/HR": millz_common.xp_per_hour(def_gained, start_time), \
            "Strength Levels Gained": str_level, \
            "Strength XP/HR": millz_common.xp_per_hour(str_gained, start_time), \
            "Hits Levels Gained": hp_level, \
            "Hits XP/HR": millz_common.xp_per_hour(hp_gained, start_time), \
            "Spawns Tracked": len(spawn_map), \
            "Bones in Bank": bank_bone_count
            }
    
    # Remove blank empty lines from report
    for key, value in full_report.items():
        if value != 0 and value != "0":
            status_report[key] = value
    
    # Add loot items to report
    if len(loot) != 0:
        for key, value in loot.items():
            if value != 0:
                status_report["Banked Loot: " + get_item_name(int(key))] = value
            if key == "413":
                status_report["Bones Banked Per Hour"] = millz_common.xp_per_hour(value, start_time)
    else:
        inv_items = get_inventory_items()
        for item in inv_items:
            if item.id in settings.loot_ids:
                status_report["Inventory Loot: " + get_item_name(int(item.id))] = get_inventory_count_by_id(item.id)
    
    return status_report
