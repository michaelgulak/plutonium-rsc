# Hero thiever by Space

# Make sure you have at least 28 hp or this script won't work.
# Start with armour equipped and optionally sleeping bag.

# Script block should look like this:

# [script]
# name            = "hero.py"
# progress_report = "20m"
#
# [script.settings]
# food_ids  = [142, 330, 333, 335] # wine and cake
# food_count = 5
# sleep = true

import time

LOOT_IDS = [10, 161, 619, 38, 612, 152]
REVERSED_FOOD_IDS = settings.food_ids[::-1]

kill_wait      = 0
thieve_timeout = 0
move_x         = -1
move_z         = -1
init           = True
kept_items     = []
loot_bank_map  = {}

if settings.sleep:
    kept_items.append(SLEEPING_BAG)

def set_kept_items():
    global kept_items

    for i in range(get_total_inventory_count()):
        item = get_inventory_item_at_index(i)
        if item != None and item.equipped:
            kept_items.append(item.id)
    
    for id_ in settings.food_ids:
        kept_items.append(id_)

def get_adjacent_coord():
    cs = [(1, 0), (0, 1), (-1, 0), (0, -1)]
    mx = get_x()
    mz = get_z()
    for dx, dz in cs:
        nx = mx + dx
        nz = mz + dz
        path = calculate_path_to(nx, nz, 1)
        if path != None and path.length() == 1:
            return (nx, nz)

    return (None, None)

def deposit_food():
    total_count = 0
    for id_ in settings.food_ids:
        count = get_inventory_count_by_id(id_)

        if total_count+count > settings.food_count:
            if count > 0:
                to_deposit = count - (settings.food_count - total_count)
                deposit(id_, to_deposit)
                return 1000
        else:
            total_count += count
    
    return 1000

def withdraw_food(num):
    for id_ in settings.food_ids:
        bcount = get_bank_count(id_)
        if bcount >= num:
            withdraw(id_, num)
            return 1000
    
    log("Out of food")
    stop_account()
    return 1000

def bank():
    jug = get_inventory_item_by_id(140)
    if jug != None:
        drop_item(jug)
        return 1000

    if in_rect(554, 609, 4, 8): # in bank
        if not is_bank_open():
            return open_bank()
        else:
            loot_item = get_inventory_item_except(kept_items)
                            
            if loot_item != None:
                count = get_inventory_count_by_id(loot_item.id)
                deposit(loot_item.id, count)
                return 1000
            
            food_count = get_inventory_count_by_id(ids=settings.food_ids)

            if food_count > settings.food_count:
                return deposit_food()
            elif food_count < settings.food_count:
                return withdraw_food(settings.food_count-food_count)
            else:
                for id_ in LOOT_IDS:
                    loot_bank_map[str(id_)] = get_bank_count(id_)

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

    return 1000


def thieve():
    global thieve_timeout, move_x, move_z

    jug = get_inventory_item_by_id(140)
    if jug != None:
        drop_item(jug)
        return 1000

    if get_current_stat(3) <= 27:
        food = get_inventory_item_by_id(ids=REVERSED_FOOD_IDS)
        if food != None:
            use_item(food)
            return 1000
        else:
            raise RuntimeError("Invalid state, thieving without food")

    #if 142 in settings.food_ids:
    #    wine_count = get_inventory_count_by_id(142)
    #    if wine_count > settings.food_count:
    #        wine = get_inventory_item_by_id(142)
    #        if wine != None:
    #            drop_item(wine)
    #            return 1200
    
    if in_rect(554, 609, 4, 8):
        door = get_object_from_coords(550, 612)
        if door != None and door.id == 64:
            at_object(door)
            return 1300

    if thieve_timeout != 0 and time.time() < thieve_timeout:
        return 250

    if move_x != -1 and move_x != None:
        if at(move_x, move_z):
            move_x, move_z = (-1, -1)
        else:
            walk_to(move_x, move_z)
            return 700
    
    if settings.sleep and get_fatigue() > 99:
        use_sleeping_bag()
        return 2000

    hero = get_nearest_npc_by_id(324)
    if hero != None:
        if hero.in_combat():
            return 250

        thieve_npc(hero)
        thieve_timeout = time.time() + 5
        return 800
    else:
        if get_x() != 548 or get_z() != 600:
            walk_path_to(548, 600)
            return 2000

    return 1000

def loop():
    global init

    if init:
        set_kept_items()
        init = False
        if get_combat_style() != 3:
            set_combat_style(3)
            return 1000

    if in_combat():
        walk_to(get_x(), get_z())
        return 600

    if kill_wait != 0:
        if time.time() - kill_wait > 5:
            stop_script()
            stop_account()
            return 5000
        
        return 650
    
    if is_bank_open() \
        or get_total_inventory_count() == 30 \
        or get_inventory_count_by_id(ids=settings.food_ids) == 0:

        return bank()
    
    return thieve()

def on_server_message(msg):
    global thieve_timeout, move_x, move_z

    if msg.startswith("You pick the") or msg.startswith("I can't get"):
        thieve_timeout = 0
    elif msg.startswith("You fail to"):
        thieve_timeout = 0
    elif msg.startswith("@cya@You have been standing"):
        move_x, move_z = get_adjacent_coord()

def on_progress_report():
    prog_report = {"Thieving Level": get_max_stat(17),
                   "Thieving XP":    get_experience(17)}

    for id_ in LOOT_IDS:
        prog_report[get_item_name(id_)] = loot_bank_map.get(str(id_), 0)

    return prog_report

def on_kill_signal():
    global kill_wait

    if is_sleeping():
        stop_script()
        stop_account()
        return

    kill_wait = time.time()
    
def on_death():
    log("Died")
    stop_account()