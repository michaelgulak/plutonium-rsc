# Al Kharid smelter by Space

# Start with a sleeping bag in Al Kharid bank.

# Script settings should look like this:

# [script]
# name            = "smelter.py"
# progress_report = "20m"
#
# [script.settings]
# bar_type = "steel"

# Where bar_type can be "bronze", "iron", "silver", "steel",
# "gold", "mithril", "adamantite", or "runite".

import time

class Bar:
    def __init__(self, prim_ore_id, sec_ore_id, sec_ore_ratio, prim_withdraw, sec_withdraw):
        self.prim_ore_id = prim_ore_id
        self.sec_ore_id = sec_ore_id
        self.sec_ore_ratio = sec_ore_ratio
        self.prim_withdraw = prim_withdraw
        self.sec_withdraw = sec_withdraw

bar           = None
smelt_timeout = 0
bars_smelted  = 0
ore_remaining = 0

if settings.bar_type == "bronze":
    bar = Bar(150, 202, 1, 14, 14)
elif settings.bar_type == "iron":
    bar = Bar(151, -1, 0, 29, 0)
elif settings.bar_type == "silver":
    bar = Bar(383, -1, 0, 29, 0)
elif settings.bar_type == "steel":
    bar = Bar(151, 155, 2, 9, 18)
elif settings.bar_type == "gold":
    bar = Bar(152, -1, 0, 29, 0)
elif settings.bar_type == "mithril":
    bar = Bar(153, 155, 4, 5, 20)
elif settings.bar_type == "adamantite":
    bar = Bar(154, 155, 6, 4, 24)
elif settings.bar_type == "runite":
    bar = Bar(409, 155, 8, 3, 24)
else:
    raise RuntimeError("Invalid bar_type")

def bank():
    global ore_remaining

    if not in_rect(93, 689, 7, 12):
        if in_radius_of(86, 695, 15):
            door = get_object_from_coords(86, 695)
            if door != None and door.id == 64:
                at_object(door)
                return 1000

        walk_to(88, 695)
        return 1000

    if not is_bank_open():
        return open_bank()
    
    item = get_inventory_item_except([bar.prim_ore_id, bar.sec_ore_id, SLEEPING_BAG])
    if item != None:
        deposit(item.id, get_inventory_count_by_id(item.id))
        return 1300

    
    prim_count = get_inventory_count_by_id(bar.prim_ore_id)
    if prim_count > bar.prim_withdraw:
        deposit(bar.prim_ore_id, prim_count - bar.prim_withdraw)
        return 1300
    
    sec_count = -1
    if bar.sec_ore_id != -1:
        sec_count = get_inventory_count_by_id(bar.sec_ore_id)
        if sec_count > bar.sec_withdraw:
            deposit(bar.sec_ore_id, sec_count - bar.sec_withdraw)
            return 1300

    prim_bank_count = get_bank_count(bar.prim_ore_id)
    if prim_count < bar.prim_withdraw:
        count = bar.prim_withdraw - prim_count
        if prim_bank_count < count:
            log("Out of primary ore")
            stop_account()
            return 1300
        withdraw(bar.prim_ore_id, count)
        return 1300
    
    sec_bank_count = -1
    
    if bar.sec_ore_id != -1:
        sec_bank_count = get_bank_count(bar.sec_ore_id)
        
        if sec_count < bar.sec_withdraw:
            count = bar.sec_withdraw - sec_count
            if sec_bank_count < count:
                log("Out of secondary ore")
                stop_account()
                return 1000
            withdraw(bar.sec_ore_id, count)
            return 1300

    if bar.sec_ore_id != -1:
        ore_remaining = min([prim_bank_count, sec_bank_count // bar.sec_ore_ratio])
    else:
        ore_remaining = prim_bank_count

    close_bank()
    return 1300

def smelt():
    global smelt_timeout

    if not in_radius_of(84, 679, 5):
        if in_radius_of(86, 695, 15):
            door = get_object_from_coords(86, 695)
            if door != None and door.id == 64:
                at_object(door)
                return 1000

        walk_to(84, 679)
        return 1000
    
    if get_fatigue() > 99:
        use_sleeping_bag()
        return 2000

    if smelt_timeout != 0 and time.time() <= smelt_timeout:
        return 250

    ore = get_inventory_item_by_id(bar.prim_ore_id)
    if ore != None:
        furnace = get_object_from_coords(85, 679)
        if furnace != None:
            use_item_on_object(ore, furnace)
            smelt_timeout = time.time() + 5
            return 800

    return 1000

def loop():
    prim_count = get_inventory_count_by_id(bar.prim_ore_id)
    sec_count = -1
    if bar.sec_ore_id != -1:
        sec_count = get_inventory_count_by_id(bar.sec_ore_id)

    if is_bank_open() \
        or (prim_count == 0 \
            or (sec_count != -1 and sec_count // bar.sec_ore_ratio != prim_count)):
        
        return bank()

    return smelt()

def on_server_message(msg):
    global smelt_timeout, bars_smelted, ore_remaining

    if msg.startswith("bar", 15):
        bars_smelted += 1
        if ore_remaining > 0:
            ore_remaining -= 1
        smelt_timeout = 0
    elif msg.startswith("impure", 15):
        smelt_timeout = 0

def on_progress_report():
    return {"Bars Smelted":  bars_smelted,
            "Ore Remaining": ore_remaining}