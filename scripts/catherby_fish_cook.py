# Catherby fish/cook by RLN
#
# Script block should look like this:
# [script]
# name            = "catherby_fish_cook.py"
# progress_report = "20m"

# [script.settings]
# fish_type = "shrimp/anchovies"
# powerfish = false
#
# Possible values for fish_type are: shrimp/anchovies, sardines/herring,
# tuna/swordfish, lobster, mackarel/cod/bass, shark.

import time

EQUIPMENT       = [[376, SLEEPING_BAG], \
                   [377, 380, SLEEPING_BAG], \
                   [379, SLEEPING_BAG], \
                   [375, SLEEPING_BAG], \
                   [548, SLEEPING_BAG], \
                   [379, SLEEPING_BAG]]
RAW_FISH_IDS    = [[349,351], [354,361], [366,369], [372], [552,550,554], [545]]
COOKED_FISH_IDS = [[350,352], [355,362], [367,370], [373], [551,553,555], [546]]

SHRIMP_ANCHOVIES  = 0
SARDINES_HERRING  = 1
TUNA_SWORDIES     = 2
LOBSTER           = 3
MACKEREL_COD_BASS = 4
SHARK             = 5

class FishingSpot:
    def __init__(self, coord, at_type, raw_fish_ids, cooked_fish_ids, equipment):
        self.coord = coord
        self.raw_fish_ids = raw_fish_ids
        self.cooked_fish_ids = cooked_fish_ids
        self.equipment = equipment
        self.at_type = at_type

fishing_spot = None
fish_banked  = {}
fish_timeout = 0
cook_timeout = 0
move_timer = False

if settings.fish_type == "shrimp/anchovies":
    fishing_spot = FishingSpot((418,500), 1, RAW_FISH_IDS[SHRIMP_ANCHOVIES], COOKED_FISH_IDS[SHRIMP_ANCHOVIES], EQUIPMENT[SHRIMP_ANCHOVIES])
elif settings.fish_type == "sardines/herring":
    fishing_spot = FishingSpot((418,500), 2, RAW_FISH_IDS[SARDINES_HERRING], COOKED_FISH_IDS[SARDINES_HERRING], EQUIPMENT[SARDINES_HERRING])
elif settings.fish_type == "tuna/swordfish":
    fishing_spot = FishingSpot((409,504), 1, RAW_FISH_IDS[TUNA_SWORDIES], COOKED_FISH_IDS[TUNA_SWORDIES], EQUIPMENT[TUNA_SWORDIES])
elif settings.fish_type == "lobster":
    fishing_spot = FishingSpot((409,504), 2, RAW_FISH_IDS[LOBSTER], COOKED_FISH_IDS[LOBSTER], EQUIPMENT[LOBSTER])
elif settings.fish_type == "mackarel/cod/bass":
    fishing_spot = FishingSpot((406,505), 1, RAW_FISH_IDS[MACKEREL_COD_BASS], COOKED_FISH_IDS[MACKEREL_COD_BASS], EQUIPMENT[MACKEREL_COD_BASS])
elif settings.fish_type == "shark":
    fishing_spot = FishingSpot((406,505), 2, RAW_FISH_IDS[SHARK], COOKED_FISH_IDS[SHARK], EQUIPMENT[SHARK])
else:
    raise RuntimeError("Unknown fish type")

def fish():
    global fish_timeout, move_timer

    if move_timer:
        if get_x() != fishing_spot.coord[0] or get_z() != fishing_spot.coord[1] - 2:
            walk_to(fishing_spot.coord[0], fishing_spot.coord[1] - 2)
            return 600
        else:
            move_timer = False

    if get_x() != fishing_spot.coord[0] or get_z() != fishing_spot.coord[1] - 1:
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
        if item_to_bank.id in fishing_spot.cooked_fish_ids:
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

def loop():
    if get_fatigue() > 95:
        use_sleeping_bag()
        return 5000
    
    if is_bank_open() \
        or ((get_total_inventory_count() == 30) \
        and not settings.powerfish):

        return cook()
    
    return fish()

def on_server_message(msg):
    global fish_timeout, cook_timeout, move_timer

    if msg.startswith("You accidentally") or msg.endswith("nicely cooked"):
        cook_timeout = 0
    elif msg.startswith("You fail") or msg.startswith("You catch"):
        fish_timeout = 0
    elif msg.startswith("@cya@You have been standing"):
        move_timer = True

def on_progress_report():
    count = 0
    for k in fish_banked:
        count += fish_banked[k]
    return {"Fishing Level": get_max_stat(10), \
            "Cooking Level": get_max_stat(7), \
            "Fish Banked":   count}