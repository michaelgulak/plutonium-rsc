# Rune miner v1.1 by O N I O N

# Banks at Falador West. Walks back on death. Banks every 3 ore

# Suitable for lvl 3s as long as somebody is killing spiders at rocks

# Optional settings:
# ore_count = 6 (change number to bank more or less ore per trip)
# fight_mode = 3 (0 is controlled, 1 is strength, 2 is attack, 3 is defense)

# If you get an error about "calculate_path_through" you probably need to update plutonium

# v1.1 changes
# changed walk_path to use walker.py logic to handle gates & recalc at al kharid gate
# added door checks for bank door & sleep room door
# added deposit for uncut gems
# added a check for full inventory & banks if it is full
# added IDs for all pickaxes, so you can start with any type of pickaxe
# added deathcount to track how many times you die
# added check for 100 fatigue so it will bank if it somehow maxes fatigues on spiders
# added an attack for spider if it is attacking miners but only if hp is 10 or higher. (protecting those with critically low hp that have just run)
# added optional fight_mode setting with default as defense

path_to_rocks = [(298, 504), (304, 436), (333, 417), (335, 381), (335, 342), (335, 311), (335, 298), (335, 255), (335, 240), (335, 206), (335, 173), (257, 157)]
path_to_bank = [(335, 173), (335, 206), (335, 240), (335, 255), (335, 298), (335, 311), (335, 342), (335, 381), (333, 417), (304, 436), (298, 504), (330, 553)]
rune_rocks = [[256,158],[258,156]]
PICKAXE = 156
RUNE_ORE = 409
rune_ore_banked = -1
move_timer = False
deathcount = 0

PICKAXES = [156, 1258, 1259, 1260, 1261, 1262]
UNCUT_GEMS = [157,158,159,160]

import time

ore_count = None
if hasattr(settings, "ore_count"):
    ore_count = settings.ore_count
else:
    ore_count = 3

fight_mode = 3
if hasattr(settings, "fight_mode"):
    fight_mode = settings.fight_mode

path = None

class Gate:
    def __init__(self, id, x, z, is_x, thresh_to, thresh_from):
        self.id = id
        self.x = x
        self.z = z
        self.is_x = is_x
        self.thresh_to = thresh_to
        self.thresh_from = thresh_from

GATES = [
    Gate(137, 341, 487, True, 341, 342),
    Gate(254, 434, 682, True, 434, 435),
    Gate(138, 343, 581, False, 580, 581),
    Gate(346, 331, 142, False, 141, 142),
    Gate(346, 111, 142, False, 141, 142),
]

def loop():
    global path, move_timer

    if get_combat_style() != fight_mode:
        set_combat_style(fight_mode)

    if in_combat():
        walk_to(get_x(), get_z())
        return 300

    if path != None:
        return walk_path()
    elif get_inventory_count_by_id(ids=PICKAXES) < 1:
        return get_pick()
    elif get_inventory_count_by_id(RUNE_ORE) >= ore_count:
        return bank()
    elif get_inventory_count_by_id(RUNE_ORE) > 0 and get_z() > 425:
        return bank()
    elif get_total_inventory_count() == 30:
        return bank()
    elif get_fatigue() > 0 and get_z() > 425:
        return sleep()
    elif get_z() > 1000:
        return sleep()
    elif get_fatigue() == 100:
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
    door = get_wall_object_from_coords(309, 525)
    if door != None:
        if door.id == 2:
            at_wall_object(door)
            return 1280
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
    bank_door = get_object_from_coords(327, 552)
    if bank_door != None:
        if bank_door.id == 64:
            at_object(bank_door)
            return 640
    if not in_rect(334, 549, 7, 9):
        debug("not in bank")
        path = calculate_path_through(points=path_to_bank)
        if path != None:
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
            uncut_gems = get_inventory_item_by_id(ids=UNCUT_GEMS)
            for GEM_ID in UNCUT_GEMS:
                uncut_gem_count = get_inventory_count_by_id(GEM_ID)
                if uncut_gem_count > 0:
                    deposit(GEM_ID, get_inventory_count_by_id(GEM_ID))
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
        for gate in GATES:
            if in_radius_of(gate.x, gate.z, 15):
                pnc = path.next_z()
                mnc = get_z()
                if gate.is_x:
                    pnc = path.next_x()
                    mnc = get_x()
                    
                if (pnc <= gate.thresh_to and mnc >= gate.thresh_from) or \
                    (pnc >= gate.thresh_from and mnc <= gate.thresh_to):
                
                    gate_ = get_object_from_coords(gate.x, gate.z)
                    if gate_ != None and gate_.id == gate.id:
                        at_object(gate_)

                    return 800

        if not path.complete():
            if not path.walk():
                path = None
                return 100
            return 800
        else:
            path = None
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

    spider = get_nearest_npc_by_id_in_rect(99, in_combat=False, reachable=True, x=257, z=157, width=1, height=2)
    if spider != None and get_current_stat(3) >= 10:
        attack_npc(spider)
        return 640

    return 150


def on_progress_report():    
    return {"Rune Ore Banked": rune_ore_banked,
            "Times Died": deathcount}

def on_server_message(msg):
    global move_timer

    if msg.startswith("@cya@You have been standing"):
        if get_x() == 257 and get_z() == 157:
            move_timer = True

def on_death():
    global deathcount
    deathcount += 1