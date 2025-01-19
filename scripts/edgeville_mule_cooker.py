# Edgeville mule cooker by Space
#
# Start mule in Edgeville bank or upstairs in the Edgeville general store.
#
# Make sure mule has raw shark in the bank.
#
# Start cooker upstairs in the Edgeville general store. Should have
# cooking gauntlets in inventory. The script will work if you don't have
# cooking gauntlets as well. Do not start cooker with valuable items,
# the script will drop all unnecessary items.
#
# [script]
# name = "edgeville_mule_cooker.py"
# progress_report = "20m"
#
# [script.settings]
# role = "cook"
# partner = "mule_username"
#
# or
#
# [script.settings]
# role = "mule"
# partner = "cook_username"

import time

if settings.role == "cook":
    ROLE_COOK = True
elif settings.role == "mule":
    ROLE_COOK = False
else:
    raise RuntimeError("invalid role")

RAW_SHARK         = 545
COOKED_SHARK      = 546
COOKING_GAUNTLETS = 700
COOK_KEPT_ITEMS   = [COOKING_GAUNTLETS, RAW_SHARK, COOKED_SHARK]
MULE_KEPT_ITEMS   = [RAW_SHARK]

shark_left       = 0
start_cooking_xp = 0
shark_cooked     = 0
start_time       = 0
cook_timeout     = 0
init             = True
trading          = False
trade_timeout    = 0

def cook():
    global cook_timeout
    
    if get_fatigue() > 99:
        bed = get_object_from_coords(226, 1386)
        if bed != None:
            at_object(bed)
            return 1200
    
    junk = get_inventory_item_except(COOK_KEPT_ITEMS)
    if junk != None:
        drop_item(junk)
        return 800
    
    if cook_timeout != 0 and time.time() < cook_timeout:
        return 100
    
    raw_shark = get_inventory_item_by_id(RAW_SHARK)
    rang = get_object_from_coords(222, 1385)
    if raw_shark != None and rang != None:
        use_item_on_object(raw_shark, rang)
        cook_timeout = time.time() + 5
        return 1000
    
    return 1000

def bank():
    global shark_left
    
    if get_z() > 1000:
        ladder = get_object_from_coords(226, 1383)
        if ladder != None:
            at_object(ladder)
            return 600
        
        return 1000
    
    if not in_rect(220, 448, 9, 6):
        door = get_wall_object_from_coords(225, 444)
        if door != None and door.id == 2:
            at_wall_object(door)
            return 800
        
        if in_radius_of(217, 447, 15):
            door = get_object_from_coords(217, 447)
            if door != None and door.id == 64:
                at_object(door)
                return 1300
        
        walk_to(218, 448)
        return 600

    if not is_bank_open():
        return open_bank()
    
    junk = get_inventory_item_except(MULE_KEPT_ITEMS)
    if junk != None:
        deposit(junk.id, get_inventory_count_by_id(junk.id))
        return 1200
    
    inv_count = get_inventory_count_by_id(RAW_SHARK)
    bank_count = get_bank_count(RAW_SHARK)
    
    shark_left = inv_count + bank_count
    
    if bank_count < 24 - inv_count:
        log("Out of raw shark")
        stop_account()
        return 1000

    withdraw(RAW_SHARK, 24 - inv_count)
    return 1200

def trade_partner():
    global trade_timeout
    
    if trade_timeout != 0 and time.time() < trade_timeout:
        return 250
    
    player = get_player_by_name(settings.partner)
    if player != None:
        trade_player(player)
        trade_timeout = time.time() + 5
        return 800
        
    return 500

def trade_from_mule():
    if in_rect(220, 448, 9, 6):
        door = get_object_from_coords(217, 447)
        if door != None and door.id == 64:
            at_object(door)
            return 1300

    if get_z() < 1000:
        door = get_wall_object_from_coords(225, 444)
        if door != None and door.id == 2:
            at_wall_object(door)
            return 800
        
        ladder = get_object_from_coords(226, 439)
        if ladder != None:
            at_object(ladder)
            return 600

    if not at(226, 1384):
        walk_to(226, 1384)
        return 600
    
    if is_trade_confirm_screen():
        confirm_trade()
        return 800

    if not is_trade_offer_screen():
        return trade_partner()
    
    trade_count = min(12, get_inventory_count_by_id(RAW_SHARK))
    if trade_count == 0 or has_my_offer(RAW_SHARK, trade_count):
        accept_trade_offer()
        return 800
    else:
        shark = get_inventory_item_by_id(RAW_SHARK)
        if shark != None:
            trade_offer_item(trade_count, shark)
            return 800
    
    return 1000

def trade_from_cook():
    junk = get_inventory_item_except(COOK_KEPT_ITEMS)
    if junk != None:
        drop_item(junk)
        return 800

    if is_trade_confirm_screen():
        confirm_trade()
        return 800
    
    if not is_trade_offer_screen():
        return trade_partner()
    
    trade_count = min(12, get_inventory_count_by_id(COOKED_SHARK))
    if trade_count == 0 or has_my_offer(COOKED_SHARK, trade_count):
        accept_trade_offer()
        return 800
    else:
        shark = get_inventory_item_by_id(COOKED_SHARK)
        if shark != None:
            trade_offer_item(trade_count, shark)
            return 800
    
    return 1000

def init_script():
    global trading, start_time, start_cooking_xp
    
    if ROLE_COOK:    
        start_time = time.time()
        start_cooking_xp = get_experience(7)
    
        if get_inventory_count_by_id(RAW_SHARK) == 12 and \
            get_inventory_count_by_id(COOKED_SHARK) == 0:

            trading = True
    else:
        if 24 > get_inventory_count_by_id(COOKED_SHARK) > 0 or \
            (get_inventory_count_by_id(RAW_SHARK) == 12 and \
                get_total_inventory_count() == 12):
            
            trading = True
def loop():
    global init, trading
    
    if init:
        if not is_friend(settings.partner):
            add_friend(settings.partner)
            return 1200
        
        init_script()
        init = False
            
    if ROLE_COOK:
        if get_inventory_count_by_id(RAW_SHARK) == 0 or trading:
            trading = True
            return trade_from_cook()
        else:
            return cook()
    else:
        if get_inventory_count_by_id(RAW_SHARK) >= 24 or trading:
            trading = True
            return trade_from_mule()
        else:
            return bank()

def on_private_message(msg, from_name):
    global trading
    
    if ROLE_COOK and from_name == settings.partner and msg == "Banking":
        trading = False
        
def on_server_message(msg):
    global cook_timeout, shark_cooked, trading, trade_timeout
    
    if msg.startswith("You accidentally"):
        cook_timeout = 0
    elif msg.endswith("successfully"):
        trade_timeout = 0
        if ROLE_COOK:
            if get_inventory_count_by_id(RAW_SHARK) >= 24:
                trading = False
        else:
            if get_inventory_count_by_id(RAW_SHARK) < 12:
                trading = False
                send_private_message(settings.partner, "Banking")

def on_fatigue_update(fatigue, accurate_fatigue):
    global cook_timeout, shark_cooked
    
    cook_timeout = 0
    if fatigue != 0:
        shark_cooked += 1

def per_hour(gained, start_time_seconds):
    elapsed_time = time.time() - start_time_seconds
    if elapsed_time == 0:
        return 0
    gained_per_second = gained / elapsed_time
    return int(gained_per_second * 3600)

def on_progress_report():
    if not ROLE_COOK:
        return {"Raw shark left": shark_left}
    
    cooking_xp = get_experience(7)
    return {"Cooking xp/hr": per_hour(cooking_xp-start_cooking_xp, start_time),
            "Cooking xp":    cooking_xp,
            "Shark cooked":  shark_cooked,
            "Shark/hr":      per_hour(shark_cooked, start_time)}
            