import time

# Simplified version of AA_VarrockEastMiner that only mines 2 rocks written by RLN
#
# Script block should look like this:
# [script]
# name            = "varrock_east_miner.py"
# progress_report = "20m"
#
#   [script.settings]
#   ore_type  = "iron"
#   powermine = false
#
# ore_type can be set to copper, tin, or iron.

PICKAXES     = [1262, 1261, 1260, 1259, 1258, 156]
ORES         = [202, 151, 150]

ONE_SECOND   = 1
FIVE_SECONDS = 5

rocks = None
coord = None

previous_rock         = None
current_rock          = None
click_timeout         = 0
previous_rock_timeout = 0

ore_banked = 0

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
             Rock(101, 73, 549)]
    coord = Coord(73, 548)
elif settings.ore_type == "tin":
    rocks = [Rock(105, 79, 546), \
             Rock(105, 78, 545)]
    coord = Coord(79, 545)
elif settings.ore_type == "iron":
    rocks = [Rock(103, 76, 544), \
             Rock(102, 75, 543)]
    coord = Coord(75, 544)

if rocks == None:
    raise RuntimeError("Invalid settings")

def bank_next_item():
    global ore_banked

    items = get_inventory_items()
    item_to_bank = None
    for item in items:
        if item.id == SLEEPING_BAG or item.id in PICKAXES:
            continue
        item_to_bank = item
        break

    if item_to_bank != None:
        inv_count = get_inventory_count_by_id(item_to_bank.id)
        if item_to_bank.id in ORES:
            ore_banked = get_bank_count(item_to_bank.id) + inv_count
        deposit(item_to_bank.id, inv_count)

    
    return 2000

def mine():
    global click_timeout

    if get_fatigue() >= 95:
        use_sleeping_bag()
        return 3000

    if get_x() != coord.x or get_z() != coord.z:
        walk_path_to(coord.x, coord.z)
        return 5000
    
    if click_timeout != 0 \
        and time.time() <= click_timeout:
        return 250
    
    for rock in rocks:
        if previous_rock == rock \
            and time.time() <= previous_rock_timeout:
            continue
        obj = get_object_from_coords(rock.x, rock.z)
        if obj == None or obj.id != rock.id:
            continue
        
        current_rock = rock
        at_object(obj)
        click_timeout = time.time() + FIVE_SECONDS
        return 700
    
    return 700

def bank():
    if in_rect(106, 510, 8, 5): # in bank
        if not is_bank_open():
            return open_bank()
        else:
            if get_total_inventory_count() != 2:
                return bank_next_item()
            else:
                close_bank()
                return 1000
    else:
        walk_path_to(102, 512)
        return 5000

def loop():
    if is_bank_open() \
        or ((get_total_inventory_count() == 30) \
        and not settings.powermine):

        return bank()
    
    return mine()

def on_server_message(msg):
    global click_timeout, previous_rock_timeout, previous_rock

    if msg.startswith("You only") or msg.startswith("There is"):
        click_timeout = 0
    elif msg.startswith("You manage") or msg.startswith("You just"):
        click_timeout = 0
        previous_rock = current_rock
        previous_rock_timeout = time.time() + ONE_SECOND

def on_progress_report():
    return {"Mining Level": get_max_stat(14), \
            "Ore Banked":   ore_banked}
