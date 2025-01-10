# This script requires the character to have a pick axe and sleeping bag in inventory.
# It also requires mining level 60 (mining guild), and smithing 30 (smelt steel).

# The script will mine iron until it has *bars_per_loop* number of ore at Varrock East mine.
# It will then mine 2x *bars_per_loop* at the mining guild in Falador.
# It will then smelt the bars at Falador to make steel bars, and go back to mining Iron.

# [Mandatory] bars_per_loop                     is how many iron/coal it mines.
# [Optional]  pickup                            if true, will pickup ore (i.e. from powerminers) in addition to mining.
# [Optional]  trade_to_cannonball_smelter       is the name of a partner that's running millz_falador_cannonball.py. 
#                                               If the value is defined, the script will smelt the bars, and trade them to the smelter character.

# Credit to Space as lots of the mining code has come from his varrock east miner/guild miner scripts.

# [script.settings]
# bars_per_loop = 1000
# pickup = true
# trade_to_cannonball_smelter = "Millz8"

import time
import millz_common

DEBUG = False

COAL_ROCKS = [110, 111]
COAL_ORE = 155
IRON_ORE = 151
STEEL_BAR = 171
IRON_BAR = 170
GEMS = [160, 159, 158, 157]

init = False
bank_items_checked = False
iron_ore_bank = 0
coal_ore_bank = 0
steel_bars_bank = 0
current_activity = ""

start_time = 0
path = None
destination = [0, 0]

smelt_timeout = 0

# From Space's Varrock East Miner:
previous_rock         = None
current_rock          = None
click_timeout         = 0
previous_rock_timeout = 0
move_x                = -1
move_z                = -1

class Rock:
    def __init__(self, id, x, z):
        self.id = id
        self.x = x
        self.z = z

class Coord:
    def __init__(self, x, z):
        self.x = x
        self.z = z

iron_rocks = [Rock(102, 75, 543), \
             Rock(102, 76, 543), \
             Rock(103, 76, 544)]
iron_coord = Coord(75, 544)


def mine_rock(x, z):
    debug("ATOBJECT (%d, %d)" % (x, z))
    create_packet(136)
    write_short(x)
    write_short(z)
    send_packet()

last_debug_msg = ""
def debug_log(msg):
    global last_debug_msg
    if last_debug_msg != msg:
        if DEBUG:
            log(msg)
        else:
            debug(msg)
        last_debug_msg = msg


def handle_banking(task):
    '''
    task: 
      - "mining" will deposit all iron/coal ores. 
      - "smelting" will deposit steel and withdraw iron/coal
      - "init" will only check for item counts
    '''
    global iron_ore_bank, coal_ore_bank, steel_bars_bank, bank_items_checked
    global path, destination

    if millz_common.is_in_bank():
        if not is_bank_open():
            return open_bank()
        else:
            iron_ore_bank = get_bank_count(IRON_ORE)
            coal_ore_bank = get_bank_count(COAL_ORE)
            steel_bars_bank = get_bank_count(STEEL_BAR)
            bank_items_checked = True

            if get_inventory_count_by_id(IRON_BAR) > 0:
                debug_log("Deposit accidental iron bars")
                deposit(IRON_BAR, get_inventory_count_by_id(IRON_BAR))
                return 1000
            
            gems = get_inventory_item_by_id(ids=GEMS)
            if gems != None:
                debug_log("Deposit Gems")
                deposit(gems.id, get_inventory_count_by_id(gems.id))
                return 1000
            
            if get_inventory_count_by_id(STEEL_BAR) > 0:
                debug_log("Deposit steel bars")
                deposit(STEEL_BAR, get_inventory_count_by_id(STEEL_BAR))
                return 1000

            if task == "mining":
                debug_log("Banking - Mining")
                
                ore_item = get_inventory_item_by_id(ids=[IRON_ORE, COAL_ORE])
                if ore_item != None:
                    deposit(ore_item.id, get_inventory_count_by_id(ore_item.id))
                    return 1000
                
                close_bank()
                return 1000

            elif task == "smelting":
                debug_log("Banking - Smelting")
                
                if get_inventory_count_by_id(IRON_ORE) > 9:
                    debug_log("Too much iron ore, depositing excess")
                    deposit(IRON_ORE, get_inventory_count_by_id(IRON_ORE) - 9)
                    return 1000
                
                if get_inventory_count_by_id(COAL_ORE) > 18:
                    debug_log("Too much coal, depositing excess")
                    deposit(COAL_ORE, get_inventory_count_by_id(COAL_ORE) - 18)
                    return 1000


                if get_inventory_count_by_id(IRON_ORE) < 9:
                    debug_log("Withdraw iron ore")
                    withdraw(IRON_ORE, 9 - get_inventory_count_by_id(IRON_ORE))
                    return 1000
                
                if get_inventory_count_by_id(COAL_ORE) < 18:
                    debug_log("Withdraw coal ore")
                    withdraw(COAL_ORE, 18 - get_inventory_count_by_id(COAL_ORE))
                    return 1000
                    
                debug_log("Finished banking, close bank")
                close_bank()
                return 1000
            else:
                debug_log("Banking - Init")

    else:
        # Mining Guild
        if get_z() > 3000:
            ladder = get_object_from_coords(274, 3398)
            if ladder != None:
                at_object(ladder)
                return 700
        else:
            if current_activity == "Smelting":
                debug_log("Banking at Falador West")
                NEAREST_BANK = millz_common.get_bank_by_name("Falador West")
            elif current_activity == "Mining Iron":
                debug_log("Banking at Varrock East")
                NEAREST_BANK = millz_common.get_bank_by_name("Varrock East")
            elif current_activity == "Mining Coal":
                debug_log("Banking at Falador East")
                NEAREST_BANK = millz_common.get_bank_by_name("Falador East")
            else:
                NEAREST_BANK = millz_common.get_nearest_bank(get_x(), get_z())
                debug_log("Banking at closest bank: %s" % NEAREST_BANK.name)

            destination = [NEAREST_BANK.x, NEAREST_BANK.z]
            path = calculate_path_to(NEAREST_BANK.x, NEAREST_BANK.z)
            if path == None:
                log("Failed to path to bank - stopping script")
                stop_account()
    return 650


def smelt():
    global path, destination, smelt_timeout, current_activity, move_x, move_z

    if smelt_timeout != 0 and time.time() <= smelt_timeout:
        return 50
    
    if coal_ore_bank < 18 or iron_ore_bank < 9:
        log("Changing activity from %s to Mining Iron" % current_activity)
        current_activity = "Mining Iron"
        return 500

    if move_x != -1:
        if get_x() == move_x and get_z() == move_z:
            move_x = -1
            move_z = -1
        else:
            walk_to(move_x, move_z)
            return 700

    iron_count = get_inventory_count_by_id(IRON_ORE)
    coal_count = get_inventory_count_by_id(COAL_ORE)

    if iron_count >= 1 and coal_count >= 2:
        debug_log("Iron: " + str(iron_count) + " - Coal: " + str(coal_count))
        # Smelt
        furnace = get_object_from_coords(310, 546)
        coal = get_inventory_item_by_id(COAL_ORE)
        if furnace != None and coal != None:
            use_item_on_object(coal, furnace)
            smelt_timeout = time.time() + 5
            return 400
        else:
            destination = [311, 545]
            path = calculate_path_to(destination[0], destination[1])
            if path == None:
                log("Failed to path to falador furnaces - stopping script")
                set_autologin(False)
                stop_account()
                logout()
            return 650
    else:
        if is_trade_confirm_screen():
            if is_trade_confirm_accepted():
                return 650
        
            debug_log("Confirm Trade")
            confirm_trade()
            return 650

        if is_trade_offer_screen():
            inventory_steel_count = get_inventory_count_by_id(STEEL_BAR)
            if inventory_steel_count > 12:
                trade_item_count = 12
            else:
                trade_item_count = inventory_steel_count

            if has_my_offer(STEEL_BAR, trade_item_count):
                if not is_trade_confirm_accepted():
                    debug_log("Accept trade offer")
                    accept_trade_offer()
                    return 650
        
                debug_log("Waiting for partner to accept")
                return 650
            else:
                debug_log("Offer item")
                trade_offer_item(trade_item_count, get_inventory_item_by_id(STEEL_BAR))
                return 650
        
        if hasattr(settings, "trade_to_cannonball_smelter") and get_inventory_count_by_id(STEEL_BAR) >= 9:
            player = get_player_by_name(settings.trade_to_cannonball_smelter)
            if player != None:
                debug_log("Trading bars to partner %s" % settings.trade_to_cannonball_smelter)
                trade_player(player)
                return 650

        debug_log("Smelt() -> Banking")
        return handle_banking("smelting")


def mine_iron():
    global click_timeout, current_rock, move_x, move_z, path, destination, current_activity

    if iron_ore_bank >= settings.bars_per_loop:
        log("Changing activity from %s to Mining Coal" % current_activity)
        current_activity = "Mining Coal"
        return 50
    
    if is_bank_open() or (get_total_inventory_count() == 30):
            return handle_banking("mining")

    if move_x != -1:
        if get_x() == move_x and get_z() == move_z:
            move_x = -1
            move_z = -1
        else:
            walk_to(move_x, move_z)
            return 700

    if get_x() != iron_coord.x or get_z() != iron_coord.z:
        if in_radius_of(102, 509, 15):
            door = get_object_from_coords(102, 509)
            if door != None and door.id == 64:
                at_object(door)
                return 1300

        destination = [iron_coord.x, iron_coord.z]
        path = calculate_path_to(destination[0], destination[1])
        if path == None:
            log("Failed to path to varrock east mine - stopping script")
            set_autologin(False)
            stop_account()
            logout()
        return 3000
    
    if get_fatigue() > 95:
        use_sleeping_bag()
        return 3000

    if click_timeout != 0 \
        and time.time() <= click_timeout:
        return 250
        
    if hasattr(settings, "pickup") and settings.pickup == True:
        pickup_items = [IRON_ORE, COAL_ORE, 160, 159, 158, 157]
        ground_item = get_nearest_ground_item_by_id(ids=pickup_items, reachable=True, x=get_x(), z=get_z(), radius=3)
        if ground_item != None:
            pickup_item(ground_item)
            return 1000
    
    for rock in iron_rocks:
        if previous_rock == rock \
            and time.time() <= previous_rock_timeout:
            continue
        obj = get_object_from_coords(rock.x, rock.z)
        if obj == None or obj.id != rock.id:
            continue
        
        current_rock = rock
        mine_rock(obj.x, obj.z)
        click_timeout = time.time() + 5
        return 700
    
    return 700


def mine_coal():
    global path, destination, current_activity

    if coal_ore_bank >= (settings.bars_per_loop * 2):
        log("Changing activity from %s to Smelting" % current_activity)
        current_activity = "Smelting"
        return 50

    if is_bank_open() or (get_total_inventory_count() == 30):
            return handle_banking("mining")
    
    if get_z() < 3000: # Above ground, done banking.

        ladder = get_object_from_coords(274, 566)
        if ladder != None:
            at_object(ladder)
            return 1000
        
        destination = [274,565]
        path = calculate_path_to(274,565)
        if path == None:
            log("Failed to path to mining guild - stopping script")
            set_autologin(False)
            stop_account()
            logout()

    else:
        if hasattr(settings, "pickup") and settings.pickup == True:
            pickup_items = [IRON_ORE, COAL_ORE, 160, 159, 158, 157]
            ground_item = get_nearest_ground_item_by_id(ids=pickup_items, reachable=True, x=get_x(), z=get_z(), radius=3)
            if ground_item != None:
                pickup_item(ground_item)
                return 1000
    
        obj = get_nearest_object_by_id_in_rect(ids=COAL_ROCKS, x=277, z=3381, width=14, height=19)
        if obj != None:
            at_object(obj)
            return 700 

    return 0


def check_level_requirements():
    if get_max_stat(14) < 60:
        log("Mining level is less than 60!")
        return False
    
    if get_max_stat(13) < 30:
        log("Smithing level is less than 30!")
        return False
    return True

def loop():
    global init, bank_items_checked, current_activity, path, destination, start_time

    if in_combat():
        log("Warning: in combat at %d, %d" % (get_x(), get_z()))

    if path != None:
        debug_log("Walking...")
        door = get_nearest_object_by_id(64, x=get_x(), z=get_z(), reachable=True, radius=8)
        if door != None:
            log("Opening bank style door")
            at_object(door)
            return 1500
        
        wall_door = get_nearest_wall_object_by_id(2, x=get_x(), z=get_z(), reachable=True, radius=8)
        if wall_door != None:
            log("Opening normal style door")
            at_wall_object(wall_door)
            return 1500

        path.process()
        if not path.complete():
            if not path.walk() and destination[0] != 0:
                path = calculate_path_to(destination[0], destination[1])
                if path == None:
                    stop_script()
                    set_autologin(False)
                    logout()
                    log("Could not path to " + str(destination[0]) + ", " + str(destination[1]) + ". Stopping.")
            return 1000
        else:
            path = None

    if not init:
        start_time = time.time()

        if not check_level_requirements():
            stop_account()
            set_autologin(False)
            logout()

        if not bank_items_checked:
            debug_log("Check counts of items in bank")
            return handle_banking("init")
            
        if current_activity == "":
            if iron_ore_bank < settings.bars_per_loop:
                log("Setting initial activity to Mining Iron")
                current_activity = "Mining Iron"
            elif coal_ore_bank < (settings.bars_per_loop * 2):
                log("Setting initial activity to Mining Coal")
                current_activity = "Mining Coal"
            else:
                log("Setting initial activity to Smelting")
                current_activity = "Smelting"

        init = True
    
    if get_fatigue() > 98:
        use_sleeping_bag()
        return 1000
    
    if current_activity == "Smelting":
        return smelt()

    if current_activity == "Mining Iron":
        return mine_iron()

    if current_activity == "Mining Coal":
        return mine_coal()

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
    global click_timeout, previous_rock_timeout, previous_rock, move_x, move_z, smelt_timeout

    if msg.startswith("You only") or msg.startswith("There is"):
        click_timeout = 0
    elif msg.startswith("You manage") or msg.startswith("You just"):
        click_timeout = 0
        previous_rock = current_rock
        previous_rock_timeout = time.time() + 1
    elif msg.startswith("@cya@You have been standing"):
        move_x, move_z = get_adjacent_coord()
    elif msg.startswith("You retrieve a bar of"):
        smelt_timeout = 0


def get_friendly_location():
    global path
    if get_z() > 3000:
        return "Mining Guild"
    
    if distance_to(iron_coord.x, iron_coord.z) < 3:
        return "Varrock East Mine"
    
    if distance_to(310, 546) < 3:
        return "Falador Furnace"
    
    bank = millz_common.get_current_bank()
    if bank != None:
        return "In Bank: %s" % bank.name
    
    if path != None:
        return "Walking @ (%d, %d)" % (get_x(), get_z())

    return "%d, %d" % (get_x(), get_z())

def on_progress_report():
    status_report = { "Status: Task": current_activity, \
        "Bank: Iron Ore": "%d/%d" % (iron_ore_bank, settings.bars_per_loop), \
        "Bank: Coal Ore": "%d/%d" % (coal_ore_bank, (settings.bars_per_loop * 2)), \
        "Bank: Steel Bar": steel_bars_bank, \
        "Status: Location": get_friendly_location()
    }

    inv_items = get_inventory_items()
    for item in inv_items:
        status_report["Inventory Item: " + get_item_name(int(item.id))] = get_inventory_count_by_id(item.id)

    return status_report
