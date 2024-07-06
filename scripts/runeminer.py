# Rune miner v1.0 by O N I O N

# Banks at Falador West. Walks back on death. Banks every 3 ore

# Suitable for lvl 3s as long as somebody is killing spiders at rocks

# Optional settings:
# ore_count = 3 (change number to bank more or less ore per trip)

# If you get an error about "calculate_path_through" you probably need to update plutonium

path_to_rocks = [(298, 504), (304, 436), (333, 417), (335, 381), (335, 342), (335, 311), (335, 298), (335, 255), (335, 240), (335, 206), (335, 173), (257, 157)]
path_to_bank = [(335, 173), (335, 206), (335, 240), (335, 255), (335, 298), (335, 311), (335, 342), (335, 381), (333, 417), (304, 436), (298, 504), (330, 553)]
rune_rocks = [[256,158],[258,156]]
PICKAXE = 156
RUNE_ORE = 409
rune_ore_banked = -1
move_timer = False

import time

path = None

ore_count = None
if hasattr(settings, "ore_count"):
    ore_count = settings.ore_count
else:
    ore_count = 3

def loop():
    global path, move_timer

    if in_combat():
        walk_to(get_x(), get_z())
        return 300

    if path != None:
        return walk_path()
    elif get_inventory_count_by_id(PICKAXE) < 1:
        return get_pick()
    elif get_inventory_count_by_id(RUNE_ORE) >= ore_count:
        return bank()
    elif get_inventory_count_by_id(RUNE_ORE) > 0 and get_z() > 425:
        return bank()
    elif get_fatigue() > 0 and get_z() > 425:
        return sleep()
    elif get_z() > 1000:
        return sleep()
    elif move_timer:
        if get_x() == 257 and get_z() == 157:
            walk_to(257, 158)
            return 640
        else:
            move_timer = False
            return 17
    elif get_x() != 257 or get_z() != 157:
        return walk()
    else:
        return mine()

def sleep():
    global path

    if get_fatigue() == 0:
        ladder = get_object_from_coords(308, 1466)
        if ladder != None:
            at_object(ladder)
            return 1000
    elif get_fatigue() > 0:
        bed = get_object_from_coords(310, 1467)
        if bed != None:
            at_object(bed)
            return 1000
        else:
            ladder = get_object_from_coords(308, 522)
            if ladder != None:
                at_object(ladder)
                return 1000
            else:
                path = calculate_path_to(308, 523)
                return 100

    return 100

def bank():
    global path, rune_ore_banked

    if not in_rect(334, 549, 7, 9):
        path = calculate_path_through(points=path_to_bank)
        path.set_nearest()
        return 100
    elif in_rect(334, 549, 7, 9):
        if not is_bank_open():
            return open_bank()
        elif is_bank_open():
            rune_inv_count = get_inventory_count_by_id(RUNE_ORE)
            rune_bank_count = get_bank_count(RUNE_ORE)
            rune_ore_banked = rune_inv_count + rune_bank_count
            deposit(RUNE_ORE, rune_inv_count)
            close_bank()
            return 640

def walk():
    global path

    if get_x() == 257 and get_z() == 158:
        walk_to(257, 157)
        return 640

    path = calculate_path_through(points=path_to_rocks)
    path.set_nearest()
    return 100

def walk_path():
    global path

    if path != None:
        path.process()
        if not path.complete():
            path.walk()
            return 600
        else:
            path = None
            return 100
    else:
        return 100

def get_pick():
    global path

    door = get_wall_object_from_coords(230, 511)
    pickaxe = get_nearest_ground_item_by_id(PICKAXE, x=231, z=508, radius=5)
    if door != None and door.id == 2:
        at_wall_object(door)
        return 1000
    elif pickaxe != None:
        pickup_item(pickaxe)
        return 1000
    elif distance_to(231,509) > 5:
        path = calculate_path_to(230, 511)
        return 100
    else:
        walk_to(231, 509)
        return 1000

def mine():
    for rune_rock in rune_rocks:
        rock = get_object_from_coords(rune_rock[0],rune_rock[1])
        if rock != None and rock.id != 98:
            debug("mining rock at: " + str(rune_rock[0]) + "," + str(rune_rock[1]))
            create_packet(136)
	    write_short(rock.x)
	    write_short(rock.z)
	    send_packet()
            return 640
    return 150


def on_progress_report():    
    return {"Rune Ore Banked": rune_ore_banked}

def on_server_message(msg):
    global move_timer

    if msg.startswith("@cya@You have been standing"):
        if get_x() == 257 and get_z() == 157:
            move_timer = True