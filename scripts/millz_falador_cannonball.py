# Script to make cannonballs in Falador. 
# Used in conjunction with many mules running the millz_steel_maker.py script.
# This script only handles incoming trading and smelting. If you want to use mule/master use the Al Kharid version.

# [script.settings]
# mules = ["mule1", "mule2", "mule3"]


import time

CANNONBALL_MOULD = 1057
STEEL_BAR = 171
CANNONBALL = 1041
FURNACE = 118

perform_trade = False
trade_partner = None
steel_bar_count = 0
times_traded = 0
smelt_timeout = 0
move_timer = False
init = False
start_cannonball_count = 0
start_steel_bar_count = 0
start_time = 0


def use_item_on_object_nowalk(item, obj):
    create_packet(115)
    write_short(obj.x)
    write_short(obj.z)
    write_short(item.slot)
    send_packet()


def smelt():
    global perform_trade, smelt_timeout, move_timer, start_time
    global init, start_cannonball_count, trade_partner

    if not init:
        init = True
        start_time = time.time()
        start_cannonball_count = get_inventory_count_by_id(CANNONBALL)

    if move_timer:
        if get_x() != 311 or get_z() != 544:
            walk_to(311, 544)
            return 1000
        else:
            move_timer = False

    if get_fatigue() > 98:
        use_sleeping_bag()
        return 1000
    
    if is_trade_confirm_screen():
        confirm_trade()
        return 1000
    
    if is_trade_offer_screen():
        if not is_trade_confirm_accepted():
            accept_trade_offer()
        return 1000

    # We have a trade partner and we have 12 spaces free to receive bars.
    if (trade_partner != None and (30 - get_total_inventory_count()) >= 12):
        player = get_player_by_name(trade_partner)
        if player != None:
            #log("Trading %s" % trade_partner)
            trade_player(player)
            return 1000
        else:
            # log("Can't find player %s" % trade_partner)
            perform_trade = False
            trade_partner = None
            return 1000
    else:
        if smelt_timeout != 0 and time.time() <= smelt_timeout:
            return 50
        
        # smelt spot = 311, 545
        # move spot = 311, 544
        # furnace = 310, 546 (id: 118)

        if get_x() != 311 or get_z() != 545:
            walk_to(311, 545)
            return 600
    
        # Ran out of steel bars, assume no mules trading.
        if get_inventory_count_by_id(STEEL_BAR) == 0:
            # log("No bars or mule, waiting for trading request")
            return 2000

        # Smelt
        furnace = get_object_from_coords(310,546)
        bar = get_inventory_item_by_id(STEEL_BAR)
        if furnace != None and bar != None:
            #log("Smelting... %d bars remaining" % get_inventory_count_by_id(STEEL_BAR))
            use_item_on_object_nowalk(bar, furnace)
            smelt_timeout = time.time() + 5
            return 400

        return 1200


def loop():
    return smelt()


def on_server_message(msg):
    global smelt_timeout, move_timer, perform_trade, trade_partner

    if msg.startswith("@cya@You have been standing"):
        move_timer = True
    if "very heavy" in msg:
        smelt_timeout = 0
    if "Trade completed successfully" in msg:
        perform_trade = False
        trade_partner = None


def items_per_hour(processed_count, elapsed_time):
    if elapsed_time == 0 or processed_count == 0:
        return 0
    else:
        return (processed_count / elapsed_time) * 3600
    

def on_trade_request(name):
    global perform_trade, trade_partner

    for mule in settings.mules:
        if name == mule:
            trade_partner = mule
            perform_trade = True
            return
    log("Trade request from %s who isn't in our mule list" % name)

def on_progress_report():

    elapsed_time = time.time() - start_time
       
    balls_created = get_inventory_count_by_id(CANNONBALL) - start_cannonball_count

    if balls_created == 0:
        return {"No balls created": ""}

    balls_per_hour = items_per_hour(balls_created, elapsed_time)

    return {"Inventory: Cannonballs": get_inventory_count_by_id(CANNONBALL), \
            "Cannonballs (This Session)": balls_created, \
            "Inventory: Steel Bars": get_inventory_count_by_id(STEEL_BAR), \
            "Balls per hour": "%.2f" % balls_per_hour}

