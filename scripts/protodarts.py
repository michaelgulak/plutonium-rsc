# Prototype dart fletcher by Space
#
# Have feathers, sleeping bag, technical plans, and a hammer in your inventory
# and start beside the desert anvil. Set the fatigue_trick setting to true
# at high fletching levels to speed up xp gains. If your inventory is full
# it will drop your last item (gnomeball support).
#
# [script]
# name            = "protodarts.py"
# progress_report = "1h"
#
# [script.settings]
# fatigue_trick = false

import time

TIP = 1071
FEATHERS = 381
DART = 1014
ANVIL_X = 170
ANVIL_Z = 792
SLEEP_TIME = 50

answer_timeout = 0
anvil_timeout = 0
bag_timeout = 0
attach_timeout = 0
start_fletching_xp = 0
start_time = 0
start_feathers = 0
move_timeout = 0
move_x = -1
move_z = -1
init = True

def at_object_no_walk(x, z):
    debug("ATOBJECTNOWALK (%d, %d)" % (x, z))
    create_packet(136)
    write_short(x)
    write_short(z)
    send_packet()

def loop():
    global init, start_fletching_xp, start_time, start_feathers, \
        answer_timeout, bag_timeout, anvil_timeout, attach_timeout, \
            move_timeout, move_x, move_z
    
    if init:
        start_fletching_xp = get_experience(9)
        start_time = time.time()
        start_feathers = get_inventory_count_by_id(FEATHERS)  
        set_fatigue_tricking(settings.fatigue_trick)
        init = False
    
    if not in_rect(171, 791, 1, 2):
        walk_to(171, 792)
        return 700

    if get_total_inventory_count() == 30:
        item = get_inventory_item_at_index(29)
        if item.id != TIP:
            if item.id == FEATHERS:
                log("Inventory full")
                stop_account()
                return 5000
            else:
                # Yes, it will drop your last item if your inventory is full.
                # Gnomeball/cracker support
                drop_item(item)
                return 1200
    
    t = time.time()
    
    if answer_timeout != 0 and t < answer_timeout:
        return SLEEP_TIME
    
    if is_option_menu():
        answer(0)
        answer_timeout = t + 10
        return SLEEP_TIME
    
    if move_x != -1:
        if at(move_x, move_z):
            move_x, move_z = (-1, -1)
            bag_timeout = 0
        else:
            if move_timeout != 0 and t < move_timeout:
                return SLEEP_TIME
            
            walk_to(move_x, move_z)
            move_timeout = time.time() + 0.7
            return SLEEP_TIME
        
    if bag_timeout != 0 and t < bag_timeout:
        return SLEEP_TIME
     
    if get_fatigue() > 99:
        bag_timeout = t + 5
        use_sleeping_bag()
        return SLEEP_TIME
    
    tip = get_inventory_item_by_id(TIP)
    if tip != None:
        if attach_timeout != 0 and t < attach_timeout:
            return SLEEP_TIME
        
        feathers = get_inventory_item_by_id(FEATHERS)
        if feathers == None or feathers.amount < 10:
            log("Out of feathers")
            stop_account()
            return 5000
        
        use_item_with_item(tip, feathers)
        attach_timeout = time.time() + 20
        return SLEEP_TIME
    else:
        if anvil_timeout != 0 and t < anvil_timeout:
            return SLEEP_TIME
        
        at_object_no_walk(ANVIL_X, ANVIL_Z)
        anvil_timeout = time.time() + 20
        return SLEEP_TIME
    
    return SLEEP_TIME
    
def on_server_message(msg):
    global move_x, move_z, bag_timeout, anvil_timeout, answer_timeout, attach_timeout
    
    if msg.startswith("@cya@You have been standing"):
        if at(171, 792):
            move_x, move_z = (171, 791)
        elif at(171, 791):
            move_x, move_z = (171, 792)
    elif msg.startswith("You wake up") or msg.startswith("You are unexpectedly awoken"):
        bag_timeout = 0
        anvil_timeout = 0
        attach_timeout = 0
    elif msg.startswith("You waste the bronze bar"):
        anvil_timeout = 0
        answer_timeout = 0
    elif msg.startswith("You need to attach"):
        attach_timeout = 0
        answer_timeout = 0
    elif msg.startswith("But you feel"):
        attach_timeout = 0
    elif msg.startswith("You succesfully"):
        anvil_timeout = 0
        answer_timeout = 0

def per_hour(gained, elapsed_time):
    if elapsed_time == 0:
        return 0
    return int(gained / elapsed_time * 3600)

def on_progress_report():
    fletching_xp = get_experience(9)
    fletch_diff = fletching_xp - start_fletching_xp
        
    feathers_count = get_inventory_count_by_id(FEATHERS)
    feathers_diff = start_feathers - feathers_count
    
    t_diff = time.time() - start_time
    
    return {"Fletching xp/hr": per_hour(fletch_diff, t_diff),
            "Fletching xp":    fletching_xp,
            "Feathers left":   feathers_count,
            "Feathers/hr":     per_hour(feathers_diff, t_diff)}
    