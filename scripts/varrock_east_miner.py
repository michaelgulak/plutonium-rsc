# Plutonium version of AA_VarrockEastMiner by RLN
#
# Script block should look like this:
# [script]
# name            = "varrock_east_miner.py"
# progress_report = "20m"
#
# [script.settings]
# ore_type  = "iron"
# powermine = false
# pickup    = false
#
# ore_type can be set to copper, tin, or iron.

import time

KEPT_ITEMS = [1263, 1262, 1261, 1260, 1259, 1258, 156]
ORES       = [202, 151, 150]

rocks = None
coord = None

previous_rock         = None
current_rock          = None
click_timeout         = 0
previous_rock_timeout = 0
move_x                = -1
move_z                = -1

ore_banked = 0

init = False
start_time = 0
start_mining_xp = 0

class Rock:
    def __init__(self, id, x, z):
        self.id = id
        self.x = x
        self.z = z

class Coord:
    def __init__(self, x, z):
        self.x = x
        self.z = z

if settings.ore_type == "copper":
    rocks = [Rock(101, 73, 547), \
             Rock(101, 72, 549), \
             Rock(101, 73, 549), \
             Rock(101, 74, 549)]
    coord = Coord(73, 548)
elif settings.ore_type == "tin":
    rocks = [Rock(105, 78, 545), \
             Rock(105, 78, 546), \
             Rock(105, 79, 546)]
    coord = Coord(79, 545)
elif settings.ore_type == "iron":
    rocks = [Rock(102, 75, 543), \
             Rock(102, 76, 543), \
             Rock(103, 76, 544)]
    coord = Coord(75, 544)

if rocks == None:
    raise RuntimeError("Invalid settings")

def bank_next_item():
    global ore_banked

    item_to_bank = get_inventory_item_except(KEPT_ITEMS)

    if item_to_bank != None:
        inv_count = get_inventory_count_by_id(item_to_bank.id)
        if item_to_bank.id in ORES:
            ore_banked = get_bank_count(item_to_bank.id) + inv_count
        deposit(item_to_bank.id, inv_count)

    
    return 2000

def mine_rock(x, z):
    debug("ATOBJECT (%d, %d)" % (x, z))
    create_packet(136)
    write_short(x)
    write_short(z)
    send_packet()

def mine():
    global click_timeout, current_rock, move_x, move_z

    if move_x != -1:
        if get_x() == move_x and get_z() == move_z:
            move_x = -1
            move_z = -1
        else:
            walk_to(move_x, move_z)
            return 700

    if get_x() != coord.x or get_z() != coord.z:
        if in_radius_of(102, 509, 15):
            door = get_object_from_coords(102, 509)
            if door != None and door.id == 64:
                at_object(door)
                return 1300

        walk_path_to(coord.x, coord.z)
        return 3000
    
    if get_fatigue() > 95:
        use_sleeping_bag()
        return 3000

    if click_timeout != 0 \
        and time.time() <= click_timeout:
        return 250
        
    if hasattr(settings, "pickup") and settings.pickup == True:
        if settings.ore_type == "iron":
            pickup_items = [151, 160, 159, 158, 157]
        else:
            pickup_items = [150, 202, 160, 159, 158, 157]
            
        ground_item = get_nearest_ground_item_by_id(ids=pickup_items, reachable=True, x=get_x(), z=get_z(), radius=3)
        if ground_item != None:
            pickup_item(ground_item)
            return 1000
    
    for rock in rocks:
        if previous_rock == rock \
            and time.time() <= previous_rock_timeout:
            continue
        obj = get_object_from_coords(rock.x, rock.z)
        if obj == None or obj.id != rock.id:
            continue
        
        current_rock = rock
        mine_rock(obj.x, obj.z)
        click_timeout = time.time() + 5
        return 700
    
    return 700

def bank():
    if in_rect(106, 510, 9, 6): # in bank
        if not is_bank_open():
            return open_bank()
        else:
            if get_total_inventory_count() != 2:
                return bank_next_item()
            else:
                close_bank()
                return 1000
    else:
        if in_radius_of(102, 509, 15):
            door = get_object_from_coords(102, 509)
            if door != None and door.id == 64:
                at_object(door)
                return 1300
            
        walk_path_to(102, 512)
        return 3000

def loop():
    global init, start_mining_xp, start_time
    
    if not init:
        init = True
        start_mining_xp = get_experience(14)
        start_time = time.time()
    
    if is_bank_open() \
        or ((get_total_inventory_count() == 30) \
            and not settings.powermine):

        return bank()
    
    return mine()

def get_adjacent_coord():
    if is_reachable(get_x()+1, get_z()):
        return (get_x()+1, get_z())
    elif is_reachable(get_x(), get_z()+1):
        return (get_x(), get_z()+1)
    elif is_reachable(get_x()-1, get_z()):
        return (get_x()-1, get_z())
    else:
        return (get_x(), get_z()-1)

def xp_per_hour(gained_xp, start_time_seconds):
    if gained_xp == 0:
        return 0
    elapsed_time = time.time() - start_time_seconds
    gained_per_second = gained_xp / elapsed_time
    return int(gained_per_second * 3600)

def on_server_message(msg):
    global click_timeout, previous_rock_timeout, previous_rock, move_x, move_z

    if msg.startswith("You only") or msg.startswith("There is"):
        click_timeout = 0
    elif msg.startswith("You manage") or msg.startswith("You just"):
        click_timeout = 0
        previous_rock = current_rock
        previous_rock_timeout = time.time() + 1
    elif msg.startswith("@cya@You have been standing"):
        move_x, move_z = get_adjacent_coord()


def on_progress_report():
    mining_xp_gained = get_experience(14) - start_mining_xp
    xp_to_max_level = 13034431 - get_experience(14)
    xp_per_hr = xp_per_hour(mining_xp_gained, start_time)
    
    status_report = { "Mining Level": get_max_stat(14) }

    if not settings.powermine:
        status_report["Ore Banked"] = ore_banked

    if xp_per_hr != 0:
        status_report["Mining XP/HR"] = xp_per_hr

        if get_max_stat(14) < 99 and xp_to_max_level > 0 and xp_per_hr > 0:
            status_report["Est time to 99 (days)"] = "%.2f" % (xp_to_max_level / xp_per_hr) / 24

    return status_report
