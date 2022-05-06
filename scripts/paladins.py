# Start near or inside paladin room
#
# This script only uses cakes.

import time

kill_wait = 0
coins_banked = 0
chaos_banked = 0
thieve_timeout = 0
move_timer = False

def get_adjacent_coord():
    if is_reachable(get_x()+1, get_z()):
        return (get_x()+1, get_z())
    elif is_reachable(get_x(), get_z()+1):
        return (get_x(), get_z()+1)
    elif is_reachable(get_x()-1, get_z()):
        return (get_x()-1, get_z())
    elif is_reachable(get_x(), get_z()-1):
        return (get_x(), get_z()-1)

    return (None, None)

def bank():
    global coins_banked, chaos_banked

    if get_z() >= 1548:
        door = get_nearest_wall_object_by_id(97)
        if door != None:
            at_wall_object(door)
            return 1000

        return 600

    elif get_z() >= 1543:
        stairs = get_object_from_coords(611, 1545)
        if stairs != None:
            at_object(stairs)
            return 1000
        
        return 600

    if in_radius_of(607, 603, 15) and get_x() >= 608:
        door = get_object_from_coords(607, 603)
        if door != None and door.id == 64:
            at_object(door)
            return 600
    
    if in_radius_of(598, 603, 15) and get_x() >= 599:
        gate = get_object_from_coords(598, 603)
        if gate != None and gate.id == 57:
            at_object(gate)
            return 600
        
    if not in_rect(585, 572, 8, 4):
        walk_path_to(581, 572)
        return 5000
    
    if not is_bank_open():
        return open_bank()
    
    bank_cake_count = get_bank_count(330)
    if bank_cake_count == 0:
        stop_script()
        set_autologin(False)
        logout()
        return 5000

    coin_count = get_inventory_count_by_id(10)
    if coin_count > 0:
        deposit(10, coin_count)
        coins_banked += coin_count
        return 2000
    
    chaos_rune_count = get_inventory_count_by_id(41)
    if chaos_rune_count > 0:
        deposit(41, chaos_rune_count)
        chaos_banked += chaos_rune_count
        return 2000

    empty_slots = get_empty_slots()
    if empty_slots < 2:
        deposit(330, 2-empty_slots)
        return 2000
    elif empty_slots > 2:
        withdraw(330, empty_slots-2)
        return 2000
    
    close_bank()
    return 1000
    
def thieve():
    global thieve_timeout, move_timer
    
    if get_fatigue() > 95:
        use_sleeping_bag()
        return 4000

    if get_current_stat(3) <= 27:
        food = get_inventory_item_by_id(ids=[335,333,330])
        if food != None:
            use_item(food)
            return 1000
        
        return 1000
    
    if get_z() < 1543:
        if in_radius_of(598, 603, 15) and get_x() <= 598:
            gate = get_object_from_coords(598, 603)
            if gate != None and gate.id == 57:
                at_object(gate)
                return 600
        if in_radius_of(607, 603, 15) and get_x() <= 607:
            door = get_object_from_coords(607, 603)
            if door != None and door.id == 64:
                at_object(door)
                return 600
        
        if get_x() != 613 or get_z() != 601:
            walk_path_to(613, 601)
            return 4000
        else:
            stairs = get_nearest_object_by_id(342)
            if stairs != None:
                at_object(stairs)
                return 600
        return 1000
    elif get_z() <= 1547:
        door = get_nearest_wall_object_by_id(97)
        if door != None:
            at_wall_object2(door)
            return 1000

        return 600
    else:
        if thieve_timeout != 0 and time.time() <= thieve_timeout:
            return 250

        if move_timer:
            adj_x, adj_z = get_adjacent_coord()
            if adj_x == None or adj_z == None:
                move_timer = False
                return 1000
            walk_to(adj_x, adj_z)
            move_timer = False
            return 1000

        paladin = get_nearest_npc_by_id(323, in_combat=False)
        if paladin != None:
            thieve_npc(paladin)
            thieve_timeout = time.time() + 3
            return 800
        
        return 700
    
    return 1000

def loop():
    if in_combat():
        walk_to(get_x(), get_z())
        return 600

    if kill_wait != 0:
        if time.time() - kill_wait > 5:
            stop_account()
            return 5000
        
        return 650

    if get_combat_style() != 3:
        set_combat_style(3)
        return 1000
    
    if is_bank_open() \
        or get_inventory_count_by_id(ids=[330,333,335]) == 0:

        return bank()
    
    return thieve()

def on_server_message(msg):
    global thieve_timeout, move_timer

    if msg.startswith("You pick the") or msg.startswith("I can't get"):
        thieve_timeout = 0
    elif msg.startswith("You fail to"):
        thieve_timeout = time.time() + 0.65
    elif msg.startswith("@cya@You have been standing"):
        move_timer = True

def on_progress_report():
    return {"Thieving Level": get_max_stat(17), \
            "Coins": get_inventory_count_by_id(10), \
            "Chaos": get_inventory_count_by_id(41), \
            "Coins banked": coins_banked, \
            "Chaos banked": chaos_banked}

def on_kill_signal():
    global kill_wait

    kill_wait = time.time()