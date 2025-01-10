# Catherby fish/cook by Space
# Modified by Millz for optional bank raw fish. cook_and_bank, and raw_bank should be set exclusively.
#   Setting both cook_and_bank and raw_bank to false will powerfish (drop items on ground).
#
# Script block should look like this:
# [script]
# name            = "catherby_fisher.py"
# progress_report = "20m"

# [script.settings]
# fish_type = "shrimp/anchovies"
# cook_and_bank = false
# raw_bank = false
#
# Possible values for fish_type are: shrimp/anchovies, sardines/herring,
# tuna/swordfish, lobster, mackarel/cod/bass, shark.

import time

EQUIPMENT       = [[376, 1263], [377, 380, 1263], [379, 1263], [375, 1263], [548, 1263], [379, 1263]]
RAW_FISH_IDS    = [[349,351], [354,361], [366,369], [372], [552,550,554], [545]]
COOKED_FISH_IDS = [[350,352], [355,362], [367,370], [373], [551,553,555], [546]]

SHRIMP_ANCHOVIES  = 0
SARDINES_HERRING  = 1
TUNA_SWORDFISH    = 2
LOBSTER           = 3
MACKEREL_COD_BASS = 4
SHARK             = 5

class FishingSpot:
    def __init__(self, coord, at_type, id_):
        self.coord = coord
        self.raw_fish_ids = RAW_FISH_IDS[id_]
        self.cooked_fish_ids = COOKED_FISH_IDS[id_]
        self.equipment = EQUIPMENT[id_]
        self.at_type = at_type

fishing_spot = None
fish_banked  = {}
fish_timeout = 0
cook_timeout = 0
move_timer = False
init = True
start_time = 0
start_fish_xp = 0
start_cook_xp = 0

if settings.fish_type == "shrimp/anchovies":
    fishing_spot = FishingSpot((418,500), 1, SHRIMP_ANCHOVIES)
elif settings.fish_type == "sardines/herring":
    fishing_spot = FishingSpot((418,500), 2, SARDINES_HERRING)
elif settings.fish_type == "tuna/swordfish":
    fishing_spot = FishingSpot((409,504), 1, TUNA_SWORDFISH)
elif settings.fish_type == "lobster":
    fishing_spot = FishingSpot((409,504), 2, LOBSTER)
elif settings.fish_type == "mackarel/cod/bass":
    fishing_spot = FishingSpot((406,505), 1, MACKEREL_COD_BASS)
elif settings.fish_type == "shark":
    fishing_spot = FishingSpot((406,505), 2, SHARK)
else:
    raise RuntimeError("Unknown fish type")

def fish():
    global init, fish_timeout, move_timer, start_fish_xp, start_cook_xp, start_time
    
    if init:
        start_fish_xp = get_experience(10)
        start_cook_xp = get_experience(7)
        start_time = time.time()
        init = False

    if move_timer:
        if not at(fishing_spot.coord[0], fishing_spot.coord[1] - 2):
            walk_to(fishing_spot.coord[0], fishing_spot.coord[1] - 2)
            return 600
        else:
            move_timer = False

    if not at(fishing_spot.coord[0], fishing_spot.coord[1] - 1):
        if in_radius_of(439, 497, 15):
            door = get_object_from_coords(439, 497)
            if door != None and door.id == 64:
                at_object(door)
                return 1300
            
        walk_path_to(fishing_spot.coord[0], fishing_spot.coord[1] - 1)
        return 5000

    if fish_timeout != 0 and time.time() <= fish_timeout:
        return 250
    
    fish = get_object_from_coords(fishing_spot.coord[0], fishing_spot.coord[1])

    if fish != None:
        if fishing_spot.at_type == 1:
            at_object(fish)
        elif fishing_spot.at_type == 2:
            at_object2(fish)

        fish_timeout = time.time() + 5
        return 900

    return 5000  

def bank_next_item():
    global fish_banked

    item_to_bank = get_inventory_item_except(fishing_spot.equipment)

    if item_to_bank != None:
        inv_count = get_inventory_count_by_id(item_to_bank.id)
        if (item_to_bank.id in fishing_spot.cooked_fish_ids) or \
            (item_to_bank.id in fishing_spot.raw_fish_ids):
            fish_banked[str(item_to_bank.id)] = get_bank_count(item_to_bank.id) + inv_count

        deposit(item_to_bank.id, inv_count)

    return 2000

def cook():
    global cook_timeout

    if get_inventory_count_by_id(ids=fishing_spot.raw_fish_ids) > 0:
        if not in_rect(436,480,5,6):
            if in_radius_of(435,486,15):
                door = get_wall_object_from_coords(435,486)
                if door != None and door.id == 2:
                    at_wall_object(door)
                    return 1300

            walk_path_to(433,483)
            return 5000
        if cook_timeout != 0 and time.time() <= cook_timeout:
            return 250
        fish = get_inventory_item_by_id(ids=fishing_spot.raw_fish_ids)
        rang = get_nearest_object_by_id(11)
        if fish != None and rang != None:
            use_item_on_object(fish, rang)
            cook_timeout = time.time() + 5
            return 1300
    else:
        if not in_rect(443,491,7,6):
            if in_radius_of(435,486,15) and get_z() <= 485:
                door = get_wall_object_from_coords(435,486)
                if door != None and door.id == 2:
                    at_wall_object(door)
                    return 1300
            if in_radius_of(439,497,15):
                door = get_object_from_coords(439, 497)
                if door != None and door.id == 64:
                    at_object(door)
                    return 1300
            
            walk_to(439,496)
            return 1200
        if not is_bank_open():
            return open_bank()
        else:
            if len(fishing_spot.equipment) != get_total_inventory_count():
                return bank_next_item()
            else:
                close_bank()
                return 1000
    
    return 5000

def bank():
    if in_radius_of(439,497,15):
        door = get_object_from_coords(439, 497)
        if door != None and door.id == 64:
            at_object(door)
            return 1300
    
    if distance_to(439, 496) > 10:
        walk_path_to(439,496)
        return 1200

    if not is_bank_open():
        return open_bank()
    else:
        if len(fishing_spot.equipment) != get_total_inventory_count():
            return bank_next_item()
        else:
            close_bank()
            return 1000


def loop():
    if get_fatigue() > 99:
        use_sleeping_bag()
        return 5000
    
    if settings.cook_and_bank \
        and (is_bank_open() or get_total_inventory_count() == 30):

        return cook()
    
    if settings.raw_bank \
        and (is_bank_open() or get_total_inventory_count() == 30):

        return bank()
        
    return fish()


def xp_per_hour(gained_xp, start_time_seconds):
    elapsed_time = time.time() - start_time_seconds
    if elapsed_time == 0:
        return 0
    gained_per_second = gained_xp / elapsed_time
    return int(gained_per_second * 3600)
    
    
def on_server_message(msg):
    global fish_timeout, cook_timeout, move_timer

    if msg.startswith("You accidentally") or msg.endswith("nicely cooked"):
        cook_timeout = 0
    elif msg.startswith("You fail") or msg.startswith("You catch"):
        fish_timeout = 0
    elif msg.startswith("@cya@You have been standing"):
        move_timer = True

def on_progress_report():
    gained_fish_xp = get_experience(10) - start_fish_xp
    gained_cook_xp = get_experience(7) - start_cook_xp
        
    count = 0
    for k in fish_banked:
        count += fish_banked[k]

    if settings.cook_and_bank:
        return {"Fishing Level": get_max_stat(10), \
                "Fishing XP/HR": xp_per_hour(gained_fish_xp, start_time), \
                "Cooking Level": get_max_stat(7), \
                "Cooking XP/HR": xp_per_hour(gained_cook_xp, start_time), \
                "Fish Banked":   count}

    return {"Fishing Level": get_max_stat(10), \
            "Fishing XP/HR": xp_per_hour(gained_fish_xp, start_time), \
            "Fish Banked":   count}