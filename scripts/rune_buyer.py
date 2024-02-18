# Rune buyer by RLN

# Note that if you have competition, how many runes you get will be determined by pid swaps.

# Script block should look like this:
# [script]
# name            = "rune_buyer.py"
# progress_report = "20m"
#
# [script.settings]
# npc_id  = 514
# buy_ids = [33, 31, 32, 34]

# npc_id should be one of:
# 514 - Magic store owner (Magic guild)
# 149 - Betty (Port Sarim)
# 54 - Aubury (Varrock)
# 793 - Lundail (Mage bank)

# buy_ids should have one or more of:
# 33 - Air-Rune
# 31 - Fire-Rune
# 32 - Water-Rune
# 34 - Earth-Rune
# 35 - Mind-Rune
# 36 - Body-Rune
# 825 - Soul-Rune

move_timer = False
stopped    = False

class Door:
    def __init__(self, id, x, z):
        self.id = id
        self.x = x
        self.z = z

DOORS = [
    Door(2, 600, 1703),
    Door(2, 271, 632),
    Door(2, 104, 525),
]

def loop():
    global stopped, move_timer

    if stopped:
        logout()
        return 1000
    
    if get_inventory_count_by_id(10) < 7:
        log("Out of money, stopping")
        set_autologin(False)
        logout()
        stopped = True
        return 1000

    if move_timer:
        x, z = get_adjacent_coord()
        walk_to(x, z)
        move_timer = False
        return 800

    if is_shop_open():
        for buy_id in settings.buy_ids:
            buy_shop_item(buy_id, 100)

        return 400

    if is_option_menu():
        answer(0)
        if settings.npc_id == 793:
            return 7000

        return 3000
    
    if settings.npc_id == 54 and get_my_player().combat_level < 21:
        log("A mugger can kill you at Aubury's before level 21, stopping")
        set_autologin(False)
        logout()
        stopped = True
        return 1000

    for door in DOORS:
        if in_radius_of(door.x, door.z, 15):
            door_ = get_wall_object_from_coords(door.x, door.z)
            if door_ != None and door_.id == door.id:
                at_wall_object(door_)
                return 800

    npc = get_nearest_npc_by_id(settings.npc_id, talking=False)
    if npc != None:
        talk_to_npc(npc)
        return 1000
    
    return 1000

def get_adjacent_coord():
    if is_reachable(get_x()+1, get_z()):
        return (get_x()+1, get_z())
    elif is_reachable(get_x(), get_z()+1):
        return (get_x(), get_z()+1)
    elif is_reachable(get_x()-1, get_z()):
        return (get_x()-1, get_z())
    else:
        return (get_x(), get_z()-1)

def on_server_message(msg):
    global move_timer

    if msg.startswith("@cya@You have been standing"):
        move_timer = True

def on_progress_report():
    prog_report = {}
    
    for buy_id in settings.buy_ids:
        prog_report[get_item_name(buy_id)] = get_inventory_count_by_id(buy_id)
    
    return prog_report