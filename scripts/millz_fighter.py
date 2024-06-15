# Script block should look like this:
# Fight Modes: 0 = Controlled, 1 = Strength, 2 = Attack, 3 = Defense
# Eat/Bank/Loot/Doors: true or false
# Bank choices: "Al Kharid", "Ardougne North", "Ardougne South", "Catherby", "Draynor", "Edgeville", "Falador East", "Falador West", "Seers Village", "Varrock East", "Varrock West", "Yanille"
# food_count: Will only withdraw up to 29 inventory items. Specifying 30, and holding 5 equipment, will withdraw 24 pieces of food. 

# [script]
# name = "millz_fighter.py"

# [script.settings]
# fight_mode = 0
# npc_ids = [407, 86]
# loot_ids = [413, 604, 814, 31, 32, 34, 35, 36, 37, 38, 40, 41, 42, 46, 157, 158, 159, 160, 436, 438, 439, 440, 441, 442, 443, 526, 527, 1092, 1277]
# eat = true
# bank = true
# loot = true
# open_doors = true
# fight_location = [659, 639]
# walkback = 20
# bank_name = "Ardougne North"
# food_count = 30

import millz_common
import time

INIT = False
NEAREST_BANK = None

START_X = settings.fight_location[0]
START_Z = settings.fight_location[1]
start_time = time.time()

start_attack_xp = 0
start_defense_xp = 0
start_strength_xp = 0
start_hits_xp = 0

start_attack_level = 0
start_defense_level = 0
start_strength_level = 0
start_hits_level = 0

loot = {}
spawn_map = {}
path = None
destination = [0, 0]

def loop():
    global INIT, NEAREST_BANK, START_X, START_Z, start_attack_xp, start_defense_xp, start_strength_xp, start_hits_xp
    global start_attack_level, start_defense_level, start_strength_level, start_hits_level, loot, spawn_map, path, destination
    
    if INIT is False:
        #log("Initialising")
        if settings.bank is True:
			# NEAREST_BANK = millz_common.get_nearest_bank(START_X, START_Z)
            NEAREST_BANK = millz_common.get_bank_by_name(settings.bank_name)
        
        start_attack_xp = get_experience(0)
        start_defense_xp = get_experience(1)
        start_strength_xp = get_experience(2)
        start_hits_xp = get_experience(3)
        
        start_attack_level = get_max_stat(0)
        start_defense_level = get_max_stat(1)
        start_strength_level = get_max_stat(2)
        start_hits_level = get_max_stat(3)
        
        INIT = True
        
        
    if get_combat_style() != settings.fight_mode:
        set_combat_style(settings.fight_mode)
        return 1000

    if settings.eat is True and millz_common.use_food():
        return 1500
    
    # Not using START_X, START_Z here as we want to be able to open a bank door etc while pathing out of our grind area.
    if settings.open_doors is True:
        door = get_nearest_object_by_id(64, x=get_x(), z=get_z(), reachable=True, radius=8)
        if door != None:
            log("Opening bank style door")
            at_object(door)
            return 1500
        
        wall_door = get_nearest_wall_object_by_id(2, x=get_x(), z=get_z(), reachable=True, radius=8)
        if wall_door != None:
            log("Opening normal style door")
            at_wall_object(wall_door)
            return 1500
        
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

    if get_hp_percent() < 20:
        log("Critically low on health, pausing")
        return 5000

    if in_combat():
        return 500

    if settings.bank is True and (is_bank_open() or (settings.eat is True and settings.food_count > 0 and millz_common.food_count() == 0) or get_total_inventory_count() == 30):
        if distance_to(NEAREST_BANK.x, NEAREST_BANK.z) > 5:
            # walk_path_to(NEAREST_BANK.x, NEAREST_BANK.z)
            destination = [NEAREST_BANK.x, NEAREST_BANK.z]
            path = calculate_path_to(NEAREST_BANK.x, NEAREST_BANK.z)
            if path == None:
                log("Failed to path to bank")
                stop_script()
            return 3000
        
        if not is_bank_open():
            return open_bank()
            
        if settings.loot is True: 
            loot_item = get_inventory_item_by_id(ids=settings.loot_ids)
            if loot_item != None:
                count_in_inventory = get_inventory_count_by_id(loot_item.id)
                #log("Deposit Loot - ID: " + str(loot_item.id) + " *" + str(count_in_inventory))
                
                loot[str(loot_item.id)] = loot.get(str(loot_item.id), 0) + count_in_inventory
                    
                deposit(loot_item.id, count_in_inventory)
                return 1500
            
        if settings.eat is True and settings.food_count > 0:
            if millz_common.withdraw_food(settings.food_count, True):
                return 2000
                
        # log("Close bank")
        spawn_map = {} # Reset spawns since they'll be out of sync after leaving area.
        close_bank()
        return 1000
                

    if get_fatigue() > 95:
        use_sleeping_bag()
        return 1000
        
    
    if distance_to(START_X, START_Z) >= settings.walkback:
        # log("Walking to fight location " + str(distance_to(START_X, START_Z)) + "yds away")
        #walk_path_to(START_X, START_Z)
        destination = [START_X, START_Z]
        path = calculate_path_to(START_X, START_Z)
        if path == None:
            log("Failed to path to fighting location")
            stop_script()
        return 4000

    if settings.loot is True:
        ground_item = get_nearest_ground_item_by_id(ids=settings.loot_ids, reachable=True, x=START_X, z=START_Z, radius=settings.walkback)
        if ground_item != None:
            pickup_item(ground_item)
            return 1000

    npc = get_nearest_npc_by_id(ids=settings.npc_ids, in_combat=False, reachable=True, x=START_X, z=START_Z, radius=settings.walkback)
    if npc != None:
        # log("Attack NPC (" + str(npc.x) + ", " + str(npc.z) + ")")
        attack_npc(npc)
        return 500

    next_spawn = get_next_spawn()
    if next_spawn is None:
        return 500
    
    if distance_to(next_spawn.get_coord()[0], next_spawn.get_coord()[1]) > 0:
        walk_to(next_spawn.get_coord()[0], next_spawn.get_coord()[1])
        
    return 500

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
    if npc.id in settings.npc_ids and distance(npc.x, npc.z, START_X, START_Z) <= settings.walkback:
        spawn_map[str(npc.sid)] = Spawn(npc.sid, npc.x, npc.z, time.time())
        # log("Spawn map now contains " + str(len(spawn_map)) + " entries")

def on_npc_despawned(npc):
    global spawn_map
    if npc.id in settings.npc_ids and distance(npc.x, npc.z, START_X, START_Z) <= settings.walkback:
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
            "Spawns Tracked": len(spawn_map)
            }
    
    # Remove blank empty lines from report
    for key, value in full_report.items():
        if value != 0:
            status_report[key] = value
    
    # Add loot items to report
    if len(loot) != 0:
        for key, value in loot.items():
            if value != 0:
                status_report["Banked Loot: " + get_item_name(int(key))] = value
    else:
        inv_items = get_inventory_items()
        for item in inv_items:
            if item.id in settings.loot_ids:
                status_report["Inventory Loot: " + get_item_name(int(item.id))] = get_inventory_count_by_id(item.id)
    
    return status_report
