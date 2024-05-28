# Script to make cannonballs with a mule at Al Kharid - Created by Millz

# [script.settings]
# mode = "smelt" # mule/smelt
# mule = "mulename"
# master = "smeltername"

import time

CANNONBALL_MOULD = 1057
STEEL_BAR = 171
CANNONBALL = 1041
FURNACE = 118

perform_trade = False
steel_bar_count = 0
times_traded = 0
smelt_timeout = 0
move_timer = False
init = False
start_cannonball_count = 0
start_steel_bar_count = 0
start_time = time.time()

def in_bank():
    return in_rect(93,689,7,12)


def use_item_on_object_nowalk(item, obj):
    create_packet(115)
    write_short(obj.x)
    write_short(obj.z)
    write_short(item.slot)
    send_packet()


def mule():
    global perform_trade, steel_bar_count, start_steel_bar_count

    if perform_trade:
        if get_inventory_count_by_id(STEEL_BAR) == 0:
            perform_trade = False
            return 600
        
        if is_trade_confirm_screen():
            #log("Confirm Trade")
            confirm_trade()
            return 800

        if is_trade_offer_screen():
            if has_my_offer(STEEL_BAR, 12):
                if not is_trade_confirm_accepted():
                    #log("Accept trade offer")
                    accept_trade_offer()
                    return 800

                #log("Waiting for master to accept")
                return 1000
            else:
                #log("Offer item")
                trade_offer_item(12, get_inventory_item_by_id(STEEL_BAR))
                return 800
        else:
            player = get_player_by_name(settings.master)
            if player != None:
                #log("Trading master")
                trade_player(player)
                return 1000
            
        perform_trade = False
        return 600
    
    # Open Bank Door
    door = get_object_from_coords(86,695)
    if door != None and door.id == 64:
        #log("Opening bank door")
        at_object(door)
        return 1000
    
    # Walk to trade waiting spot
    if get_inventory_count_by_id(STEEL_BAR) >= 24:
        #log("Returning to trading spot")
        walk_path_to(82, 679)
        return 1000

    # Withdraw bars from bank
    if get_inventory_count_by_id(STEEL_BAR) < 12:
        #log("Banking")
        if not in_bank():
            walk_path_to(88,695)
            return 1000
        else:
            if not is_bank_open():
                return open_bank()
            else:
                if get_bank_count(STEEL_BAR) < 24:
                    #log("Ran out of steel bars!")
                    stop_account()
                    set_autologin(False)
                    logout()
                    return 1000
            
                if start_steel_bar_count == 0:
                    start_steel_bar_count = get_bank_count(STEEL_BAR)

                steel_bar_count = get_bank_count(STEEL_BAR)
                withdraw(STEEL_BAR, 24 - get_inventory_count_by_id(STEEL_BAR))
                return 1500
    return 1000


def smelt():
    global perform_trade, smelt_timeout, move_timer
    global init, start_cannonball_count

    if not init:
        init = True
        start_cannonball_count = get_inventory_count_by_id(CANNONBALL)

    if move_timer:
        if get_x() != 83 or get_z() != 679:
            walk_to(83, 679)
            return 600
        else:
            move_timer = False

    if get_fatigue() > 98:
        #log("Sleeping")
        use_sleeping_bag()
        return 1000
    
    if perform_trade or get_inventory_count_by_id(STEEL_BAR) == 0:
        if get_inventory_count_by_id(STEEL_BAR) > 12:
            perform_trade = False
            return 250

        # Trade        
        if is_trade_confirm_screen():
            #log("Confirm Trade")
            confirm_trade()
            return 800

        if is_trade_offer_screen():
            if has_recipient_offer(STEEL_BAR, 12):
                if not is_trade_confirm_accepted():
                    #log("Accept trade offer")
                    accept_trade_offer()
                    return 800

                #log("Waiting for mule to offer")
                return 1000
        else:
            player = get_player_by_name(settings.mule)
            if player != None:
                #log("Trading mule")
                trade_player(player)
                return 800
            else:
                #log("Can't find mule?")
                perform_trade = False
                return 1000
            
        #log("No bars, we need to trade for more...")
        return 600

    else:
        if smelt_timeout != 0 and time.time() <= smelt_timeout:
            return 50
        
        if get_x() != 84 or get_z() != 679:
            walk_to(84, 679)
            return 600
    
        # Smelt
        furnace = get_object_from_coords(85,679)
        bar = get_inventory_item_by_id(STEEL_BAR)
        if furnace != None:
            #log("Smelting...")
            use_item_on_object_nowalk(bar, furnace)
            smelt_timeout = time.time() + 5
            return 400

        return 1200


def loop():
    if settings.mode == "mule":
        return mule()
    else:
        return smelt()


def on_server_message(msg):
    global smelt_timeout, move_timer, perform_trade, times_traded

    if msg.startswith("@cya@You have been standing"):
        move_timer = True
    if "very heavy" in msg:
        smelt_timeout = 0
    if msg.startswith("Trade completed successfully"):
        times_traded += 1


def items_per_hour(processed_count, elapsed_time):
    if elapsed_time == 0 or processed_count == 0:
        return 0
    else:
        return (processed_count / elapsed_time) * 3600
    

def on_trade_request(name):
    global perform_trade

    if settings.mode == "mule":
        partner = settings.master
        if partner != None and partner is name:
            perform_trade = True
    
    if settings.mode == "smelt":
        partner = settings.mule
        if partner != None and partner is name:
            perform_trade = True

def on_progress_report():

    elapsed_time = time.time() - start_time

    if settings.mode == "mule":
        if steel_bar_count == 0:
            return {"Waiting for 2 bank trips to calculate stats": 0}

        bars_used = start_steel_bar_count - steel_bar_count
        bars_per_hour = items_per_hour(bars_used, elapsed_time)

        if bars_per_hour == 0:
            return {"Waiting for 2 bank trips to calculate stats": 0}
        
        remaining_time_estimate = steel_bar_count / bars_per_hour

        return {"Steel bars remaining": steel_bar_count, \
                "Est time remaining (hrs)": remaining_time_estimate, \
                "Bars per hour": "%.2f" % bars_per_hour, \
                "Trade count": times_traded}
    else:
        
        balls_created = get_inventory_count_by_id(CANNONBALL) - start_cannonball_count
        balls_per_hour = items_per_hour(balls_created, elapsed_time)

        return {"Cannonballs (Total)": get_inventory_count_by_id(CANNONBALL), \
                "Cannonballs (This Session)": balls_created, \
                "Trade count": times_traded, \
                "Balls per hour": "%.2f" % balls_per_hour}

