# Fishing guild raw shark fisher & certer by Millz (Adapted from Space's Catherby)

# Will walk to, and enter the fishing guild. Doesn't handle gates, but moves from catherby etc.

# [Optional] Set master like below, if bot receives a trade request from the master character, bot will trade over its certs.
# [Optional] Set walk_to_guild to true if the bot isn't already at the fishing guild and you'd like to make it walk there. Disable after for better stability.

# [script.settings]
# master = "Millz"
# walk_to_guild = false

import time
import millz_common

DEBUG = False

SPOT_X = 593
SPOT_Z = 501
MENU_OPTIONS = ["I have some fish to trade in", "Raw shark", "Twentyfive"]
KEPT_ITEMS = [1263, 379, 631, 545]

move_timer = False
fish_timeout = 0
talk_timeout = 0
last_login_time = time.time()

path = None
destination = [0, 0]

init = False
start_time = 0
start_fish_xp = 0

perform_trade = False

def debug_log(msg):
    global DEBUG
    if DEBUG:
        log(msg)
    else:
        debug(msg)

def talk_to_npc_no_walk(npc):
    create_packet(153)
    write_short(npc.sid)
    send_packet()

def enter_fishing_guild():
    global path
    
    if path != None:
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
    
    guild_door = get_wall_object_from_coords(586, 524)
    if guild_door != None and guild_door.id == 112:
        log("Opening fishing guild door")
        at_wall_object(guild_door)
        return 1000
    
    if distance_to(586, 524) > 5:
        log("We're not in the fishing guild. Attempting to walk there (no gate logic so ensure it's walkable!)")
        destination = [586, 524]
        path = calculate_path_to(586, 524)
        if path == None:
            log("Failed to path to fishing guild")
            set_autologin(False)
            stop_script()
            logout()
        return 1000
        
    return 1000

def cert():
    global talk_timeout

    if not in_rect(605, 500, 4, 7):
        debug_log("Walking to certing hut")
        walk_path_to(604, 503)
        return 1000

    if is_option_menu():
        debug_log("Options menu")
        for opt in MENU_OPTIONS:
            idx = get_option_menu_index(opt)
            if idx != -1:
                debug_log("Option: %d" % idx)
                answer(idx)
                talk_timeout = time.time() + 5
                return 2000
    else:
        if talk_timeout != 0 and time.time() < talk_timeout:
            return 250

        padik = get_nearest_npc_by_id(370)

        if padik == None:
            log("Padik is missing - relog")
            set_autologin(True)
            logout()
            return 2000
        
        #if padik.x > 605:
        #    log("Padik is outside of the fishing guild @ %d,%d - relog" % (padik.x, padik.z))
        #    set_autologin(True)
        #    logout()
        #    return 2000
        #
        #if not is_reachable(padik.x, padik.z):
        #    log("Padik exists but isn't reachable @ %d,%d - relog" % (padik.x, padik.z))
        #    set_autologin(True)
        #    logout()
        #    return 2000

        debug_log("Talk to Padik")
        talk_to_npc_no_walk(padik)
        talk_timeout = 0
        return 1000

    return 1000

def fish():
    global fish_timeout, move_timer, path, destination

    if fish_timeout != 0 and time.time() <= fish_timeout:
        return 50
    
    if move_timer:
        if get_x() != SPOT_X or get_z() != SPOT_Z + 2:
            debug_log("Moving 5min timer")
            walk_to(SPOT_X, SPOT_Z + 2)
            return 600
        else:
            move_timer = False

    if get_fatigue() > 98:
        debug_log("Sleeping")
        use_sleeping_bag()
        return 1000
    
    item_to_drop = get_inventory_item_except(KEPT_ITEMS)
    if item_to_drop != None:
        debug_log("Drop item: %s" % item_to_drop.id)
        drop_item(item_to_drop)
        return 1200
        
    if get_x() != SPOT_X or get_z() != SPOT_Z + 1:
        debug_log("Walking to fishing spot")
        destination = [SPOT_X, SPOT_Z + 1]
        path = calculate_path_to(SPOT_X, SPOT_Z + 1)
        if path == None:
            log("Failed to path to fishing spot")
            set_autologin(False)
            stop_script()
            logout()
        return 600
    
    fish = get_object_from_coords(SPOT_X, SPOT_Z)

    if fish != None:
        debug_log("Fishing")
        at_object2(fish)
        fish_timeout = time.time() + 5
        return 500

    return 5000  


def trade():
    global perform_trade
    if get_inventory_count_by_id(631) == 0:
        log("No certs left to trade")
        perform_trade = False
        return 600
    
    if is_trade_confirm_screen():
        log("Confirm Trade")
        confirm_trade()
        return 1000

    if is_trade_offer_screen():
        if has_my_offer(631, get_inventory_count_by_id(631)):
            if not is_trade_confirm_accepted():
                log("Accept trade offer")
                accept_trade_offer()
                return 1000

            log("Waiting for master to accept")
            return 1000
        else:
            log("Offer item")
            trade_offer_item(get_inventory_count_by_id(631), get_inventory_item_by_id(631))
            return 1000
    else:
        player = get_player_by_name(settings.master)
        if player != None:
            log("Trading master")
            trade_player(player)
            return 1000
        
    perform_trade = False  
    return 600

def loop():
    global init, start_time, start_fish_xp, perform_trade, path, destination, last_login_time
    
    if init is False:
        init = True
        start_time = time.time()
        start_fish_xp = get_experience(10)
     
    if hasattr(settings, "master") and perform_trade:
        return trade()
                
    if hasattr(settings, "walk_to_guild") and settings.walk_to_guild and not millz_common.in_fishing_guild(get_x(), get_z()):
        return enter_fishing_guild()
    
    if in_radius_of(603, 506, 8):
        door = get_wall_object_from_coords(603, 506)
        if door != None and door.id == 2:
            log("Opening certer door")
            at_wall_object(door)
            return 700
        
    if path != None:
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

    #if (time.time() - last_login_time) > 86400:
    #    log("24hrs since last login, relogging to prevent desync bug")
    #    set_autologin(True)
    #    logout()
    #    return 2000

    if get_inventory_count_by_id(545) >= 25:
        return cert()
    
    return fish()

def on_server_message(msg):
    global fish_timeout, move_timer, perform_trade, last_login_time

    if msg.startswith("You fail") or msg.startswith("You catch"):
        fish_timeout = 0
    elif msg.startswith("@cya@You have been standing"):
        move_timer = True
    elif "decline" in msg:
        perform_trade = False
    elif msg.startswith("Welcome to"):
        last_login_time = time.time()

def on_progress_report():
    gained_fish_xp = get_experience(10) - start_fish_xp
    fishing_level = get_max_stat(10)

    if fishing_level == 99:
        return {"Raw shark certs": get_inventory_count_by_id(631),
            "Fishing XP/HR": millz_common.xp_per_hour(gained_fish_xp, start_time)}
    
    return {"Raw shark certs": get_inventory_count_by_id(631),
            "Fishing Level": fishing_level, \
            "Fishing XP Next Level:": millz_common.exp_until_next_level(get_experience(10)), \
            "Fishing XP/HR": millz_common.xp_per_hour(gained_fish_xp, start_time)}

def on_trade_request(name):
    global perform_trade

    log("Trade request from: " + name)

    if hasattr(settings, "master"):
        master = settings.master
        if master != None and master is name:
            perform_trade = True