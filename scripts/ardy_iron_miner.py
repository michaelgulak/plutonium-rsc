# Ardougne iron miner by Space
#
# Must have sleeping bag and pickaxe in inventory.
#
# Script block should look like this:
# [script]
# name            = "ardy_iron_miner.py"
# progress_report = "20m"
#
# [script.settings]
# spot  = "north"
# powermine = false
# pickup = false
#
# spot can be set to "north" or "south". North has 4 iron rocks, while south has 3.

import time

KEPT_ITEMS   = [1263, 1262, 1261, 1260, 1259, 1258, 156]
PICKUP_IRON  = [151, 160, 159, 158, 157]
ORE          = 151

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

if settings.spot == "north":
    rocks = [Rock(102, 618, 655),
             Rock(102, 617, 655),
             Rock(102, 616, 656),
             Rock(102, 616, 657)]
    coord = Coord(617, 656)
elif settings.spot == "south":
    rocks = [Rock(102, 619, 657),
             Rock(102, 618, 658),
             Rock(102, 617, 658)]
    coord = Coord(618, 657)

if rocks == None:
    raise RuntimeError("Incorrect spot chosen")

def bank_next_item():
    global ore_banked

    item_to_bank = get_inventory_item_except(KEPT_ITEMS)

    if item_to_bank != None:
        inv_count = get_inventory_count_by_id(item_to_bank.id)
        if item_to_bank.id == ORE:
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
        if at(move_x, move_z):
            move_x = -1
            move_z = -1
        else:
            walk_to(move_x, move_z)
            return 700

    if not at(coord.x, coord.z):
        if in_radius_of(550, 612, 15):
            door = get_object_from_coords(550, 612)
            if door != None and door.id == 64:
                at_object(door)
                return 1300

        walk_path_to(coord.x, coord.z)
        return 3000
    
    if get_fatigue() > 99:
        use_sleeping_bag()
        return 3000

    if click_timeout != 0 and time.time() < click_timeout:
        return 250

    if hasattr(settings, "pickup") and settings.pickup:
        ground_item = get_nearest_ground_item_by_id(ids=PICKUP_IRON,\
                                                    reachable=True,\
                                                    x=coord.x,\
                                                    z=coord.z,\
                                                    radius=1)
        if ground_item != None:
            pickup_item(ground_item)
            return 1000
    
    for rock in rocks:
        if previous_rock == rock and time.time() < previous_rock_timeout:
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
    if in_rect(554, 609, 4, 8): # in bank
        if not is_bank_open():
            return open_bank()
        else:
            if get_total_inventory_count() != 2:
                return bank_next_item()
            else:
                close_bank()
                return 1000
    else:
        if in_radius_of(550, 612, 15):
            door = get_object_from_coords(550, 612)
            if door != None and door.id == 64:
                at_object(door)
                return 1300

        walk_path_to(551, 612)
        return 3000

def loop():
    global init, start_mining_xp, start_time
    
    if not init:
        init = True
        start_mining_xp = get_experience(14)
        start_time = time.time()
        
    if is_bank_open() \
        or (get_total_inventory_count() == 30 \
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
    elapsed_time = time.time() - start_time_seconds
    if elapsed_time == 0:
        return 0
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
    
    prog_report = {"Mining Level": get_max_stat(14),
                   "Mining XP/HR": xp_per_hour(mining_xp_gained, start_time)}

    if not settings.powermine:
        prog_report["Ore Banked"] = ore_banked

    return prog_report