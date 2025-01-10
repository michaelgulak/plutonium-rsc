# Ardougne Guard Thiever by O N I O N (based on Space's Falador thiever)

# Pickpockets guards in Ardougne at different spots. Only uses cake.
# Must have cake in bank and sleeping bag in inventory (armour optional).

# Script block should look like this:

# [script]
# name = "ardy_guard_thiever.py"
#
# [script.settings]
# spot = 1
#
# Where spot can be 1, 2, or 3.

import time

class Spot:
    def __init__(self, minx, minz, width, height, walkx, walkz):
        self.minx = minx
        self.minz = minz
        self.width = width
        self.height = height
        self.walkx = walkx
        self.walkz = walkz

CAKE_IDS = [335,333,330]

spot = None

if settings.spot == 1:
    spot = Spot(576, 590, 15, 15, 564, 596)
elif settings.spot == 2:
    spot = Spot(558, 573, 30, 30, 549, 581)
elif settings.spot == 3:
    spot = Spot(569, 563, 10, 10, 563, 569)
else:
    raise RuntimeError('Incorrect spot chosen')

kill_combat_wait = 0

def bank():
    if not in_rect(585, 572, 9, 5):
        walk_path_to(580, 572)
        return 5000
    
    if not is_bank_open():
        return open_bank()
    
    bank_cake_count = get_bank_count(330)
    if bank_cake_count == 0:
        stop_script()
        set_autologin(False)
        logout()
        return 5000

    cake_count = get_inventory_count_by_id(330)
    if cake_count == 0:
        withdraw(330, 22)
        return 2000
    else:
        close_bank()
        return 2000
    
    return 5000

def eat_cake():
    cake = get_inventory_item_by_id(ids=CAKE_IDS)
    if cake != None:
        use_item(cake)
        return 1350
    
    return 2000

def thieve():
    if get_current_stat(3) <= 9:
        return eat_cake()

    if get_fatigue() > 99:
        use_sleeping_bag()
        return 2000
    
    if not in_rect(spot.minx, spot.minz, spot.width, spot.height):
        walk_path_to(spot.walkx, spot.walkz)
        return 5000
    
    guard = get_nearest_npc_by_id_in_rect(321, \
                                          in_combat=False, \
                                          x=spot.minx, \
                                          z=spot.minz, \
                                          width=spot.width, \
                                          height=spot.height)
    if guard != None:
        thieve_npc(guard)
        return 1000
    
    return 1200

def loop():
    if in_combat():
        walk_to(get_x(), get_z())
        return 650
    
    if kill_combat_wait != 0:
        if time.time() - kill_combat_wait > 5:
            stop_script()
            stop_account()
            return 5000
        
        return 650

    if get_combat_style() != 3:
        set_combat_style(3)
        return 2000
    
    if is_bank_open() or get_inventory_count_by_id(ids=CAKE_IDS) == 0:
        return bank()
    
    return thieve()

def on_progress_report():
    return {"Thieving Level": get_max_stat(17), 
            "Coins":          get_inventory_count_by_id(10)}

def on_kill_signal():
    global kill_combat_wait
    
    if is_sleeping():
        stop_script()
        stop_account()
        return

    kill_combat_wait = time.time()