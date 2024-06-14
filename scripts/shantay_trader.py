# Shantay Trader by O N I O N

# Start at shantay Chest. Use settings as below

#############SETTINGS#############

#####[script.settings] FOR RECEIVING ITEMS from a single player
# role = "receiver"
# giver_name = "username"

#####[script.settings] FOR RECEIVING ITEMS from multple givers
# role = "receiver"

#####[script.settings] FOR GIVING ITEMS
# role = "giver"
# receiver_name = "username"
# tradeitemids = [151,153,155]

#########END OF SETTINGS##########

##############NOTES###############

# First It will deposit all your inventory items (so dont worry about whats in your inventory to start with)
# It handles trading stackables
# It will trade all items, leaving 0 in bank
# There are no options for setting quantities
# Should Trade around 8k-8.5k items per hour

###########END NOTES##############

import time

role = settings.role

partner_name = ""
if hasattr(settings, "receiver_name"):
    partner_name = settings.receiver_name
if hasattr(settings, "tradeitemids"):
    tradeitemids = settings.tradeitemids
if hasattr(settings, "giver_name"):
    partner_name = settings.giver_name
    

init = False
start_time = time.time()
move_z                = -1
use_bank_timeout = time.time()
chestclicked_timeout = time.time()
senttradetime = time.time()
accepted_timeout = time.time()
partner_name_timeout = time.time()
senttrade = False
accepted_trade = False
total_items_traded = 0
walking = False
walking_timeout = time.time()
confirmed = False
confirmed_timeout = time.time()
loot_bank_map  = {}
received_items = {}
withdrawn = False
lasttradeitems = False
stackable_ids = [10,11,31,32,33,34,35,36,37,38,39,40,41,42,43,46,190,280,281,380,381,419,517,518,519,520,521,528,529,530,531,532,533,534,535,536,574,592,619,628,629,630,631,637,638,639,640,641,642,643,644,645,646,647,669,670,671,672,673,674,711,712,713,715,723,783,784,785,786,790,796,825,985,1013,1014,1015,1024,1030,1041,1062,1063,1064,1065,1066,1067,1068,1069,1070,1071,1118,1122,1123,1124,1125,1126,1127,1182,1254,1270,1271,1272,1273,1274,1275]
class Item:
    def __init__(self, id, amount, received=0):
        self.id = id
        self.amount = amount
        self.received = received

def move():
    global move_z, chestclicked, walking, walking_timeout

    if time.time() > walking_timeout:
        walking = False

    if move_z != -1 and not walking:
        if get_z() == move_z:
            move_z = -1
            chestclicked = False
            walking = False
            return 1
        else:
            walk_to(59, move_z)
            walking = True
            walking_timeout = time.time() +1
            return 300
    return 17

def clear_inventory():
    global chestclicked
    global use_bank_timeout, used_bank, chestclicked_timeout, start_time, init

    if time.time() > use_bank_timeout:
        used_bank = False

    if time.time() > chestclicked_timeout:
        chestclicked = False

    if not is_bank_open() and not chestclicked:
        create_packet(136)
        write_short(58)
        write_short(731)
        send_packet()
        chestclicked = True
        chestclicked_timeout = time.time() + 10
        return 100
    if is_bank_open() and not used_bank:
        chestclicked = False
        inv_items = get_inventory_items()

        processed_ids = set()
        for item in inv_items:
            if item.id not in processed_ids:
                deposit(item.id, get_inventory_count_by_id(item.id))
                processed_ids.add(item.id)

        used_bank = True
        use_bank_timeout = time.time() + 5
        start_time = time.time()
        init = True
    return 100

def loop():
    global start_time, init, move_z, lasttradeitems

    if move_z != -1:
        return move()

    if get_x() != 59 or (get_z() != 731 and get_z() != 732):
        walk_to(59, 731)
        return 1000

    if role == "receiver":
        if not init:
            if get_total_inventory_count() != 0:
                return clear_inventory()
            else:
                start_time = time.time()
                init = True

        if is_bank_open() or get_total_inventory_count() > 18:
            return bank_as_receiver()
        return trade_as_receiver()

    elif role == "giver":
        if not init:
            if get_total_inventory_count() != 0:
                return clear_inventory()
            else:
                start_time = time.time()
                init = True
        if lasttradeitems and get_total_inventory_count() != 0:
            return trade_as_giver()
        elif is_bank_open() or get_total_inventory_count() < 12 or get_inventory_count_by_id(ids=tradeitemids) == 0:
            return bank_as_giver()
        return trade_as_giver()

def bank_as_receiver():
    global chestclicked
    global use_bank_timeout, used_bank, chestclicked_timeout
    global senttradetime, senttrade
    global partner_name
    global loot_bank_map, received_items

    if time.time() > use_bank_timeout:
        used_bank = False

    if time.time() > chestclicked_timeout:
        chestclicked = False

    if not is_bank_open() and not chestclicked:
        create_packet(136)
        write_short(58)
        write_short(731)
        send_packet()
        chestclicked = True
        chestclicked_timeout = time.time() + 10
        return 100
    if is_bank_open() and not used_bank:
        chestclicked = False
        inv_items = get_inventory_items()

        processed_ids = set()
        for item in inv_items:
            if item.id not in processed_ids:
                item_id_str = str(item.id)
                if item_id_str not in received_items:
                    received_items[item_id_str] = Item(item.id, 0, 0)
                received_items[item_id_str].amount = get_bank_count(item.id) + get_inventory_count_by_id(item.id)
                received_items[item_id_str].received += get_inventory_count_by_id(item.id)
                deposit(item.id, get_inventory_count_by_id(item.id))
                processed_ids.add(item.id)

        used_bank = True
        use_bank_timeout = time.time() + 5
        close_bank()
        partner = get_player_by_name(partner_name)
        if partner != None:
            trade_player(partner)
            senttrade = True
            senttradetime = time.time()
            return 100
    return 100

def trade_as_receiver():
    global senttrade, senttradetime
    global accepted
    global deposited
    global accepted_trade, accepted_timeout
    global confirmed, confirmed_timeout
    global partner_name, partner_name_timeout

    if not hasattr(settings, "giver_name") and time.time() > partner_name_timeout:
        partner_name = ""

    if is_trade_confirm_screen() and time.time() > confirmed_timeout:
        confirmed = False

    if is_trade_confirm_screen() and not confirmed:
        confirmed = True
        confirmed_timeout = time.time() + 1
        confirm_trade()
        return 100

    if senttradetime + 5 < time.time() and senttrade == True:
        senttrade = False

    if time.time() > accepted_timeout and accepted_trade == True:
        accepted_trade = False
    if not is_trade_offer_screen() and not is_trade_confirm_screen() and not senttrade:
        partner = get_player_by_name(partner_name)
        if partner != None:
            trade_player(partner)
            senttrade = True
            senttradetime = time.time()
            return 100
    if is_trade_offer_screen() and not is_trade_accepted() and is_recipient_trade_accepted() and not accepted_trade:
        accepted_trade = True
        accepted_timeout = time.time() + 1
        accept_trade_offer()
        return 100
    return 17

def bank_as_giver():
    global chestclicked, use_bank_timeout, used_bank, chestclicked_timeout, senttradetime, senttrade, loot_bank_map, withdrawn, lasttradeitems

    if time.time() > use_bank_timeout:
        used_bank = False

    if time.time() > chestclicked_timeout:
        chestclicked = False

    if not is_bank_open() and not chestclicked:
        create_packet(136)
        write_short(58)
        write_short(731)
        send_packet()
        chestclicked = True
        chestclicked_timeout = time.time() + 10
        return 100
    if is_bank_open() and not used_bank:
        inv_items = get_inventory_items()
        for inv_item in inv_items:
            if inv_item.id not in tradeitemids:
                deposit(inv_item.id, 30)
                return 300
            
        chestclicked = False
        used_bank = True
        use_bank_timeout = time.time() + 5
        withdrawn = False
        finished = True
        if not lasttradeitems:
            for itemid in tradeitemids:
                if itemid not in stackable_ids and get_bank_count(itemid) >= 24:
                    withdraw(itemid, 24)
                    withdrawn = True
                    finished = False
                    break
        if not withdrawn or lasttradeitems:
            log("we must be out of items with 24+ quantity")
            lasttradeitems = True
            for itemid in tradeitemids:
                if get_bank_count(itemid) > 0:
                    withdraw(itemid, get_bank_count(itemid))
                    finished = False
        for id_ in tradeitemids:
            loot_bank_map[str(id_)] = get_bank_count(id_)
        if finished and get_total_inventory_count() == 0:
            log("Finished trading")
            stop_account()
            return 5000
        close_bank()
        partner = get_player_by_name(partner_name)
        if partner != None:
            trade_player(partner)
            senttrade = True
            senttradetime = time.time()
            return 100
    return 100

def trade_as_giver():
    global senttrade, senttradetime
    global lastitemid
    global accepted_trade, accepted_timeout
    global confirmed, confirmed_timeout

    if time.time() > confirmed_timeout:
        confirmed = False

    if is_trade_confirm_screen() and not confirmed:
        confirmed = True
        confirmed_timeout = time.time() + 1
        confirm_trade()
        return 100

    if senttradetime + 5 < time.time() and senttrade == True:
        senttrade = False

    if not senttrade:
        partner = get_player_by_name(partner_name)
        if partner != None:
            run_this = True
            trade_player(partner)
            senttrade = True
            senttradetime = time.time()
            return 100

    tradeitem = None

    for tradeitemid in tradeitemids:
        if get_inventory_count_by_id(tradeitemid) != 0:
            tradeitem = get_inventory_item_by_id(tradeitemid)
            break
    if tradeitem == None:
        tradeid = lastitemid
    else:
        tradeid = tradeitem.id
        lastitemid = tradeitem.id

    if not is_trade_offer_screen() and not is_trade_confirm_screen():
        return 100

    if is_trade_offer_screen() and time.time() > accepted_timeout:
        accepted_trade = False

    #trade_count = min(12, get_inventory_count_by_id(tradeid))
    trade_count = get_inventory_count_by_id(tradeid)
    
    if is_trade_offer_screen() and has_my_offer(tradeid, 1) and not accepted_trade:
        accepted_trade = True
        accepted_timeout = time.time() + 5
        accept_trade_offer()
        return 100
    elif is_trade_offer_screen() and not has_my_offer(tradeid, 1):
        trade_offer_item(trade_count, tradeitem)
        create_packet(55)
        send_packet()
        return 100

    return 17

def on_server_message(msg):
    global senttrade, senttradetime
    global chestclicked, deposited
    global move_z
    global confirmed, accepted_trade
    global total_items_traded

    if msg.startswith("@cya@You have been standing"):
        if get_z() == 731:
            move_z = 732
        elif get_z() == 732:
            move_z = 731

    if msg.startswith("Trade completed"):
        chestclicked = False
        deposited = False
        senttrade = False
        confirmed = False
        accepted_trade = False
        total_items_traded += 12

        #if not hasattr(settings, "giver"):
        #    partner_name = ""
        if role == "giver" and get_total_inventory_count() >= 12:
            return trade_as_giver()
        if role == "receiver" and get_total_inventory_count() <= 18:
            return trade_as_receiver()

    elif msg.startswith("Other player has declined"):
        senttrade = False
        senttradetime = time.time()
        chestclicked = False
        deposited = False

def on_trade_request(name):
    global senttrade, senttradetime, partner_name, partner_name_timeout

    if time.time() > senttradetime + 2:
        senttrade = False

    if not hasattr(settings, "giver"):
        partner_name = name
        partner_name_timeout = time.time() + 60

    if not hasattr(settings, "giver") and not senttrade:
        partner = get_player_by_name(partner_name)
        if partner != None:
            trade_player(partner)
            senttrade = True
            senttradetime = time.time()

    if role == "giver" and get_total_inventory_count() >= 12:
        senttrade = False
        partner = get_player_by_name(partner_name)
        if partner != None:
            trade_player(partner)
            senttrade = True
            senttradetime = time.time()
    if role == "receiver" and get_total_inventory_count() <= 18 and hasattr(settings, "giver"):
        senttrade = False
        partner = get_player_by_name(partner_name)
        if partner != None:
            trade_player(partner)
            senttrade = True
            senttradetime = time.time()

def per_hour(gained, start_time_seconds):
    elapsed_time = time.time() - start_time_seconds
    if elapsed_time == 0:
        return 0
    gained_per_second = gained / elapsed_time
    return int(gained_per_second * 3600)

def on_progress_report():    
    prog_report = {"Traded Items Per Hour": per_hour(total_items_traded, start_time)}

    if role == "giver":
        for id_ in tradeitemids:
            prog_report[get_item_name(id_) + " In Bank"] = \
                loot_bank_map[str(id_)] if str(id_) in loot_bank_map else 0
    elif role == "receiver":
        for item_id, item in received_items.items():
            prog_report[get_item_name(int(item_id)) + " In Bank"] = item.amount
        for item_id, item in received_items.items():
            prog_report[get_item_name(int(item_id)) + " Received So Far"] = item.received

    return prog_report