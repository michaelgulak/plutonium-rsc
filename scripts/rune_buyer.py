# Rune buyer by Space

# Start with coins near the npc you choose.

# Note that if you have competition, how many runes you get will be determined by pid swaps.
# For the time being this just logs out instead of moving for the 5 min move timer because
# of the npc desync bug.

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

BUYING_SOULS = 825 in settings.buy_ids

if BUYING_SOULS and settings.npc_id != 514:
    raise RuntimeError("Trying to buy souls from someone other than magic store owner")

move_x       = -1
move_z       = -1
soul_timeout = 0

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
    global move_x, move_z, soul_timeout
    
    coin_count = get_inventory_count_by_id(10)
    if (BUYING_SOULS and coin_count < 6000) or coin_count < 7:
        log("Out of money, stopping")
        stop_account()
        return 1000

    # This is removed until the npc desync bug is fixed
    #if move_x != -1:
    #    if at(move_x, move_z):
    #        move_x, move_z = (-1, -1)
    #    else:
    #        walk_to(move_x, move_z)
    #        return 700

    if is_shop_open():
        for buy_id in settings.buy_ids:
            if buy_id == 825:
                if get_shop_item_by_id(825).amount == 30:
                    if not (soul_timeout != 0 and time.time() < soul_timeout):
                        buy_shop_item(825, 1)
                        soul_timeout = time.time() + 1
            else:
                buy_shop_item(buy_id, 100)

        return 400

    if is_option_menu():
        answer(0)
        return 3000
    
    if settings.npc_id == 54 and get_my_player().combat_level < 21:
        log("A mugger can kill you at Aubury's before level 21, stopping")
        stop_account()
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
    cs = [(1, 0), (0, 1), (-1, 0), (0, -1)]
    mx = get_x()
    mz = get_z()
    for dx, dz in cs:
        nx = mx + dx
        nz = mz + dz
        path = calculate_path_to(nx, nz, 1)
        if path != None and path.length() == 1:
            return (nx, nz)

    raise RuntimeError("no adjacent coordinate found")

def on_server_message(msg):
    global move_x, move_z

    if msg.startswith("@cya@You have been standing"):
        move_x, move_z = get_adjacent_coord()

def on_progress_report():
    prog_report = { "Coins": get_inventory_count_by_id(10) }
    
    for buy_id in settings.buy_ids:
        prog_report[get_item_name(buy_id)] = get_inventory_count_by_id(buy_id)
    
    return prog_report