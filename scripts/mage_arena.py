# Mage_Arena by O N I O N

# Version: 1.0

# Start at Mage Arena bank

# script.settings are optional. If settings are not found, it will use the defaults (saradomin & defense)

# The items below must be found in your bank or inventory:
# * Knife
# * Runes for bolt or blast spells (enough for 250+ spells)
# * Runes for charging staff (200 blood and enough fire/air based on your god)
# * 24 Shark (Or more in the unlikely event the script needs to bank)

######## OPTIONAL SETTINGS ########

# god options: Saradomin = 1 | Guthix = 2 | Zamorak = 3 (DEFAULT IS SARADOMIN)
# combat_style options: Controlled = 0 | Aggressive = 1 | Accurate = 2 | Defensive = 3 (DEFAULT IS DEFENSIVE)

# Script settings should look like this (leave blank for defaults):

# [script.settings]
# combat_style = 3
# god = 1

##################################

import time

# You could potentially change the variables found in "CUSTOMISABLE". These are untested

## CUSTOMISABLE
NUMBER_OF_SPELLS = 250 # This is the number of spells you will take to the arena for killing kolodion
FOOD_REQUIRED    = 24  # do not increase beyond 24 or you will have trouble at the bank
FOOD             = 546 # You could change 546 (shark) to other food ids
EAT_AT_LOST_HP   = 20  # Set this to the value your food heals. e.g. if you used lobster change it to 12

## COMBAT STYLE (OPTIONAL SCRIPT OVERRIDE)
COMBAT_STYLE = 3
if hasattr(settings, "combat_style"):
    COMBAT_STYLE = settings.combat_style

## GOD (OPTIONAL SCRIPT OVERRIDE)
GOD = 1
if hasattr(settings, "god"):
    GOD = settings.god

## ITEM IDs
STAFFS = [101,102,103,197]
ARMOUR = [235,314,315,316,317,522,597]
KNIFE  = 13
RUNES  = [31,33,38,41]

## RUNE IDS
FIRE_RUNE  = 31
AIR_RUNE   = 33
CHAOS_RUNE = 41
DEATH_RUNE = 38
BLOOD_RUNE = 619
FIRE_STAFF = 197
AIR_STAFF  = 101

## KOLODION / BATTLE_MAGE VARIABLES
CURRENT_KOLODION = 713
KOLODION = [713,757,758,759,760]
KILLED_KOLODION = False
BATTLE_MAGE = [789,790,791]

## TIMEOUTS
TALKING = 0
DOOR = 0
LADDER = 0
WEB = 0
CAST = 0
EAT = 0
POOL = 0
CHAMBER = 0
GATE = 0

## OTHER VARIABLES
INV_READY = False
SPELL_IDX = -1
LAST_CHAT = ""
GET_FREE_STAFF = False
GOD_STAFF = []
GOD_CAPE = []
CHARGE_STAFF = False
CAST_COUNT = 0

def loop():
    global INV_READY, KILLED_KOLODION, GET_FREE_STAFF, GOD_STAFF, GOD_CAPE, CHARGE_STAFF, CAST_COUNT, TALKING

    if get_max_stat(6) < 60:
        log("Error: Your magic level is too low")
        stop_account()
        return 1280

    if CHARGE_STAFF and CAST_COUNT < 100 and not in_rect(233,126,11,9) and not in_rect(222,130,3,1) and not in_rect(236,130,3,1) and not at(236,129) and INV_READY:
        return walk_to_arena()

    ## handling our position for moving around the mage arena & returning to bank
    if in_rect(233,126,11,9) or in_rect(222,130,3,1) or in_rect(236,130,3,1) or at(236,129):
        return handle_arena()
    elif at(228, 120):
        main_entrance = get_object_from_coords(228, 119)
        if main_entrance != None:
            at_object(main_entrance)
            return 1280
    elif is_reachable(228, 120):
        walk_to(228, 120)
        return 1280
    elif get_z() < 120:
        return enter_mage_arena_bank()

    if GET_FREE_STAFF:
       return get_free_staff()

    if not INV_READY:
        return bank()
        
    if get_combat_style() != COMBAT_STYLE:
        set_combat_style(COMBAT_STYLE)

    staff = get_inventory_item_by_id(ids=[FIRE_STAFF, AIR_STAFF])
    if staff != None and not is_inventory_item_equipped(staff.id):
        equip_item(staff)
        TALKING = 0
        return 2000

    if get_z() > 1000 and not KILLED_KOLODION:
        return enter_arena()

    god_staff = get_inventory_item_by_id(GOD_STAFF)
    if god_staff != None and not is_inventory_item_equipped(GOD_STAFF):
        equip_item(god_staff)
        TALKING = 0
        return 1280

    god_cape = get_inventory_item_by_id(GOD_CAPE)
    if god_staff != None and not is_inventory_item_equipped(GOD_CAPE):
        equip_item(god_cape)
        TALKING = 0
        return 1280

    if CAST_COUNT < 100:
        CHARGE_STAFF = True
        log("Heading to charge staff")
        return 17
    else:
        log("Mage Arena Completed & Staff is Charged")
        stop_account()
        return 1280

    return 640

def walk_to_arena():
    global LADDER, DOOR, WEB, GATE

    if in_rect(458,3364,23,17): # rect of main mage bank area
        ladder = get_object_from_coords(446,3367)
        if ladder != None and timeout(LADDER):
            at_object(ladder)
            LADDER = time.time() + 10
            return 17

    if in_rect(225,108,6,4): # above ladder, before door
        door = get_wall_object_from_coords(226,110)
        if door != None and door.id == 2 and timeout(DOOR):
            at_wall_object(door)
            DOOR = time.time() + 2
            return 17

    if in_rect(225,108,6,4) or in_rect(227,109,2,2):
        if not at(227,109):
            walk_to(227,109)
            return 640

    if at(227,109):
        web = get_wall_object_from_coords(227,109)
        knife = get_inventory_item_by_id(KNIFE)
        if web != None and knife != None and web.id == 24 and timeout(WEB):
            use_item_on_wall_object(knife, web)
            WEB = time.time() + 1.28
            return 17
        elif web != None and web.id == 16:
            walk_to(227, 107)
            return 640

    if in_rect(227,107,1,2):
        web = get_wall_object_from_coords(227,107)
        knife = get_inventory_item_by_id(KNIFE)
        if web != None and knife != None and web.id == 24 and timeout(WEB):
            use_item_on_wall_object(knife, web)
            WEB = time.time() + 1.28
            return 17
        elif web != None and web.id == 16:
            walk_to(227, 106)
            return 640

    if is_reachable(228,118) and not at(228,118):
        walk_to(228,118)
        return 1280

    if at(228,118):
        main_entrance = get_object_from_coords(228, 119)
        if main_entrance != None:
            at_object(main_entrance)
            return 1280

    if is_reachable(237,129) and not at(237,129):
        walk_to(237,129)
        return 640

    if at(237,129):
        gate = get_object_from_coords(237,129)
        if gate != None and timeout(GATE):
            at_object(gate)
            GATE = time.time() + 10
            return 17

    return 17

def get_free_staff():
    global GET_FREE_STAFF, GOD_CAPE, GOD_STAFF, POOL, CHAMBER, GOD_STONE

    if in_rect(458,3364,23,17): # main bank area:
        if has_inventory_item(GOD_STAFF) and has_inventory_item(GOD_CAPE):
            GET_FREE_STAFF = False
            return 17
        else:
            magic_pool = get_object_from_coords(446,3374)
            if magic_pool != None and timeout(POOL):
                at_object(magic_pool)
                POOL = time.time() + 10
                return 640
            else:
                return 17
    elif in_rect(476,3379,15,26): # get free staff area
        if has_inventory_item(GOD_STAFF) and has_inventory_item(GOD_CAPE):
            magic_pool = get_object_from_coords(471,3383)
            if magic_pool != None and timeout(POOL):
                at_object(magic_pool)
                POOL = time.time() + 10
                return 640
            elif magic_pool == None:
                walk_path_to(471,3385)
                return 2000
            else:
                return 17
        elif has_inventory_item(GOD_CAPE) and not has_inventory_item(GOD_STAFF):
            chamber_guardian = get_nearest_npc_by_id(784, talking=False)
            if distance_to(471,3385) > 10:
                walk_path_to(471,3385)
                return 3000
            if chamber_guardian != None and timeout(CHAMBER):
                talk_to_npc(chamber_guardian)
                CHAMBER = time.time() + 30
                return 17
            else:
                return 17
        elif not has_inventory_item(GOD_CAPE):
            ground_cape = get_nearest_ground_item_by_id(GOD_CAPE, reachable=True)
            if ground_cape != None:
                pickup_item(ground_cape)
                return 1280
            else:
                if distance_to(GOD_STONE[0], GOD_STONE[1]) > 10:
                    walk_path_to(469, 3399)
                    return 1280
                god_stone = get_object_from_coords(GOD_STONE[0], GOD_STONE[1])
                if god_stone != None:
                    at_object(god_stone)
                    return 1280
                else:
                    return 17
    else:
        log("Error: not sure where I am?")
        return 3000

def enter_arena():
    global TALKING, LAST_CHAT
    if is_option_menu() and get_option_menu_option(1) == ("how can i use my new spells outside of the arena?"):
        LAST_CHAT = get_option_menu_option(1)
        answer(1)
    elif is_option_menu() and get_option_menu_option(0) != LAST_CHAT:
        LAST_CHAT = get_option_menu_option(0)
        answer(0)

    LAST_CHAT = ""

    kolodion = get_nearest_npc_by_id(712, talking=False, reachable=True)
    if kolodion != None and timeout(TALKING):
        talk_to_npc(kolodion)
        TALKING = time.time() + 30
        return 17

    kolodion = get_nearest_npc_by_id(712)
    if kolodion == None:
        log("Error: Can't find kolodion. probably in wrong location somehow")
        stop_account()
        return 1280

    return 640

def handle_arena():
    global TALKING, LAST_CHAT, CURRENT_KOLODION, SPELL_IDX, CAST, INV_READY, EAT, CHARGE_STAFF, CAST_COUNT

    if hp_lost() >= EAT_AT_LOST_HP:
        food = get_inventory_item_by_id(FOOD)
        if food != None and timeout(EAT):
            use_item(food)
            EAT = time.time() + 1.28
            return 640

    if get_inventory_count_by_id(FOOD) == 0 and get_hp_percent() <= 50:
        # logging out & back in is the quickest route to the bank!
        log("Low HP and out of food. Banking")
        INV_READY = False
        logout()
        return 640

    if CHARGE_STAFF and CAST_COUNT < 100:
        if SPELL_IDX == -1:
            SPELL_IDX = god_spell_required()
        battle_mage = get_nearest_npc_by_id(ids=BATTLE_MAGE)
        if battle_mage != None and timeout(CAST):
            cast_on_npc(SPELL_IDX, battle_mage)
            CAST = time.time() + 2
            return 17
        else:
            return 17
        return 1280

    elif CHARGE_STAFF and CAST_COUNT >= 100:
        log("Charging staff complete, logging out as quickest route to the bank!")
        logout()
        return 640

    kolodion = get_nearest_npc_by_id(CURRENT_KOLODION)
    if kolodion != None and timeout(CAST):
        cast_on_npc(SPELL_IDX, kolodion)
        CAST = time.time() + 2
        return 17

    if kolodion == None:
        # may have desynced current kolodion. trying to cast on any we can find
        kolodion = get_nearest_npc_by_id(ids=KOLODION)
        if kolodion != None and timeout(CAST):
            cast_on_npc(SPELL_IDX, kolodion)
            CAST = time.time() + 2
            return 17

    return 17

def timeout(timeout_time):
    if time.time() > timeout_time:
        return True
    else:
        return False

def enter_mage_arena_bank():
    global DOOR, LADDER, WEB

    if in_rect(227,109,2,2):
        door = get_wall_object_from_coords(226,110)
        if door != None and door.id == 2 and timeout(DOOR):
            at_wall_object(door)
            DOOR = time.time() + 2
            return 17
    if in_rect(225,108,6,4) or in_rect(227,109,2,2):
        ladder = get_object_from_coords(223, 110)
        if ladder != None and timeout(LADDER):
            at_object(ladder)
            LADDER = time.time() + 15
            return 17
        else:
            return 17
    elif in_rect(227,107,1,2):
        web = get_wall_object_from_coords(227,109)
        knife = get_inventory_item_by_id(KNIFE)
        if web != None and knife != None and web.id == 24 and timeout(WEB):
            use_item_on_wall_object(knife, web)
            WEB = time.time() + 1.28
            return 17
        elif web != None and web.id == 16:
            walk_to(227, 109)
            return 640
    elif not at(227,106):
        walk_to(227,106)
        return 640
    elif at(227,106):
        web = get_wall_object_from_coords(227,107)
        knife = get_inventory_item_by_id(KNIFE)
        if web != None and knife != None and web.id == 24 and timeout(WEB):
            use_item_on_wall_object(knife, web)
            WEB = time.time() + 1.28
            return 17
        elif web != None and web.id == 16:
            walk_to(227, 107)
            return 640
    return 640        

def deposit_all():
    inventory_items = get_inventory_items()
    deposited_ids = set()

    for inv_item in inventory_items:
        if inv_item.id not in deposited_ids:
            deposit(inv_item.id, get_inventory_count_by_id(inv_item.id))
            deposited_ids.add(inv_item.id)

def god_runes_required(id):
    global CAST_COUNT

    casts_left = 100 - CAST_COUNT

    if GOD == 1:
        if id == FIRE_RUNE:
            return 2 * casts_left
        elif id == AIR_RUNE:
            return 4 * casts_left
        elif id == BLOOD_RUNE:
            return 2 * casts_left
    elif GOD == 2:
        if id == FIRE_RUNE:
            return 1 * casts_left
        elif id == AIR_RUNE:
            return 4 * casts_left
        elif id == BLOOD_RUNE:
            return 2 * casts_left
    elif GOD == 3:
        if id == FIRE_RUNE:
            return 4 * casts_left
        elif id == AIR_RUNE:
            return 1 * casts_left
        elif id == BLOOD_RUNE:
            return 2 * casts_left

def god_spell_required():
    if GOD == 1:
        return 34
    elif GOD == 2:
        return 33
    elif GOD == 3:
        return 35

def set_god_staff_and_cape():
    global GOD_STAFF, GOD_CAPE, GOD_STONE

    if GOD == 1:
        GOD_STAFF = 1218
        GOD_CAPE = 1214
        GOD_STONE = [465,3400]
    elif GOD == 2:
        GOD_STAFF = 1217
        GOD_CAPE = 1215
        GOD_STONE = [468,3401]
    elif GOD == 3:
        GOD_STAFF = 1216
        GOD_CAPE = 1213
        GOD_STONE = [473,3400]


def bank_prep_inv():
    global INV_READY, SPELL_IDX, KILLED_KOLODION, GET_FREE_STAFF, GOD_CAPE, GOD_STAFF

    # 1. checks for knife in inv/bank. if not found it will fail
    if get_inventory_count_by_id(KNIFE) == 0 and get_bank_count(KNIFE) == 0:
        log("Error: Knife not found in inventory or bank")
        stop_account()
        return 640

    # 2. deposits everything
    if get_total_inventory_count() > 0:
        deposit_all()
        return 1280

    # 3. withdraw knife & food
    withdraw(KNIFE, 1)
    withdraw(FOOD, FOOD_REQUIRED)

    # 4. withdrawing enough runes to handle the fight & setting spell idx
    # 4.0 grabbing rune counts from bank
    chaos_runes = get_bank_count(CHAOS_RUNE)
    death_runes = get_bank_count(DEATH_RUNE)
    fire_runes = get_bank_count(FIRE_RUNE)
    air_runes = get_bank_count(AIR_RUNE)
    fire_staff = get_bank_count(FIRE_STAFF)
    air_staff = get_bank_count(AIR_STAFF)

    # 4.1. check if we are at the stage of charging staff
    if KILLED_KOLODION:
        set_god_staff_and_cape()
        if get_bank_count(GOD_STAFF) < 1 or get_bank_count(GOD_CAPE) < 1:
            withdraw(GOD_STAFF, 1)
            withdraw(GOD_CAPE, 1)
            log("Going to get the free god staff")
            GET_FREE_STAFF = True
            close_bank()
            return 2000
        air_required = god_runes_required(AIR_RUNE)
        fire_required = god_runes_required(FIRE_RUNE)
        blood_required = god_runes_required(BLOOD_RUNE)
        if get_bank_count(AIR_RUNE) < air_required or get_bank_count(FIRE_RUNE) < fire_required or get_bank_count(BLOOD_RUNE) < blood_required:
            log("Error: not enough runes to charge staff")
            stop_account()
            return 1280
        else:
            withdraw(AIR_RUNE, air_required)
            withdraw(FIRE_RUNE, fire_required)
            withdraw(BLOOD_RUNE, blood_required)
            withdraw(GOD_STAFF, 1)
            withdraw(GOD_CAPE, 1)
            SPELL_IDX = god_spell_required()
            INV_READY = True
            close_bank()
            return 2000

    # 4.2 determine if there are enough runes for NUMBER_OF_SPELLS fire blast using fire staff
    air_required = 4 * NUMBER_OF_SPELLS
    fire_required = 5 * NUMBER_OF_SPELLS
    if death_runes >= NUMBER_OF_SPELLS and fire_staff >= 1 and air_runes >= air_required:
        SPELL_IDX = 32
        withdraw(DEATH_RUNE, NUMBER_OF_SPELLS)
        withdraw(FIRE_STAFF, 1)
        withdraw(AIR_RUNE, air_required)
        INV_READY = True
        close_bank()
        return 2000

    # 4.3 determine if there are enough runes for NUMBER_OF_SPELLS fire blast using air staff
    air_required = 4 * NUMBER_OF_SPELLS
    fire_required = 5 * NUMBER_OF_SPELLS
    if death_runes >= NUMBER_OF_SPELLS and fire_runes >= fire_required and air_staff >= 1:
        SPELL_IDX = 32
        withdraw(DEATH_RUNE, NUMBER_OF_SPELLS)
        withdraw(AIR_STAFF, 1)
        withdraw(FIRE_RUNE, fire_required)
        INV_READY = True
        close_bank()
        return 2000

    # 4.4 determine if there are enough runes for NUMBER_OF_SPELLS fire blast only using runes
    air_required = 4 * NUMBER_OF_SPELLS
    fire_required = 5 * NUMBER_OF_SPELLS
    if death_runes >= NUMBER_OF_SPELLS and fire_runes >= fire_required and air_runes >= air_required:
        SPELL_IDX = 32
        withdraw(DEATH_RUNE, NUMBER_OF_SPELLS)
        withdraw(AIR_RUNE, air_required)
        withdraw(FIRE_RUNE, fire_required)
        INV_READY = True
        close_bank()
        return 2000

    # 4.5 determine if there are enough runes for NUMBER_OF_SPELLS wind blast using air staff
    air_required = 3 * NUMBER_OF_SPELLS
    if death_runes >= NUMBER_OF_SPELLS and air_staff >= 1:
        SPELL_IDX = 20
        withdraw(DEATH_RUNE, NUMBER_OF_SPELLS)
        withdraw(AIR_STAFF, 1)
        INV_READY = True
        close_bank()
        return 2000

    # 4.6 determine if there are enough runes for NUMBER_OF_SPELLS wind blast only using runes
    air_required = 3 * NUMBER_OF_SPELLS
    if death_runes >= NUMBER_OF_SPELLS and air_runes >= air_required:
        SPELL_IDX = 20
        withdraw(DEATH_RUNE, NUMBER_OF_SPELLS)
        withdraw(AIR_RUNE, air_required)
        INV_READY = True
        close_bank()
        return 2000

    # 4.7 determine if there are enough runes for NUMBER_OF_SPELLS fire bolt using fire staff
    air_required = 3 * NUMBER_OF_SPELLS
    fire_required = 4 * NUMBER_OF_SPELLS
    if chaos_runes >= NUMBER_OF_SPELLS and fire_staff >= 1 and air_runes >= air_required:
        SPELL_IDX = 17
        withdraw(CHAOS_RUNE, NUMBER_OF_SPELLS)
        withdraw(FIRE_STAFF, 1)
        withdraw(AIR_RUNE, air_required)
        INV_READY = True
        close_bank()
        return 2000

    # 4.8 determine if there are enough runes for NUMBER_OF_SPELLS fire bolt using air staff
    air_required = 3 * NUMBER_OF_SPELLS
    fire_required = 4 * NUMBER_OF_SPELLS
    if chaos_runes >= NUMBER_OF_SPELLS and fire_runes >= fire_required and air_staff >= 1:
        SPELL_IDX = 17
        withdraw(CHAOS_RUNE, NUMBER_OF_SPELLS)
        withdraw(AIR_STAFF, 1)
        withdraw(FIRE_RUNE, fire_required)
        INV_READY = True
        close_bank()
        return 2000

    # 4.9 determine if there are enough runes for NUMBER_OF_SPELLS fire bolt only using runes
    air_required = 3 * NUMBER_OF_SPELLS
    fire_required = 4 * NUMBER_OF_SPELLS
    if chaos_runes >= NUMBER_OF_SPELLS and fire_runes >= fire_required and air_runes >= air_required:
        SPELL_IDX = 17
        withdraw(CHAOS_RUNE, NUMBER_OF_SPELLS)
        withdraw(AIR_RUNE, air_required)
        withdraw(FIRE_RUNE, fire_required)
        INV_READY = True
        close_bank()
        return 2000

    # 4.10 determine if there are enough runes for NUMBER_OF_SPELLS wind bolt using air staff
    air_required = 2 * NUMBER_OF_SPELLS
    if chaos_runes >= NUMBER_OF_SPELLS and air_staff >= 1:
        SPELL_IDX = 8
        withdraw(CHAOS_RUNE, NUMBER_OF_SPELLS)
        withdraw(AIR_STAFF, 1)
        INV_READY = True
        close_bank()
        return 2000

    # 4.11 determine if there are enough runes for NUMBER_OF_SPELLS wind bolt only using runes
    air_required = 2 * NUMBER_OF_SPELLS
    if chaos_runes >= NUMBER_OF_SPELLS and air_runes >= air_required:
        SPELL_IDX = 8
        withdraw(CHAOS_RUNE, NUMBER_OF_SPELLS)
        withdraw(AIR_RUNE, air_required)
        INV_READY = True
        close_bank()
        return 2000

    log("Error: Not enough runes in the bank")
    stop_account()
    return 1280

def bank():
    if is_bank_open():
        return bank_prep_inv()

    if is_option_menu():
        answer(0)
        return 3000

    Gundai = get_nearest_npc_by_id(792)
    if Gundai == None:
        log("Error: Gundai not found, start at mage arena bank")
        return 640
    Gundai = get_nearest_npc_by_id(792, talking=False, reachable=True)
    if Gundai != None:
        talk_to_npc(Gundai)
        return 3000
    return 640

def at(x,z):
    return get_x() == x and get_z() == z

def hp_lost():
    return get_max_stat(3) - get_current_stat(3)

def on_server_message(msg):
    global WEB, CURRENT_KOLODION, KILLED_KOLODION, CAST, EAT, INV_READY, CHARGE_STAFF, CAST_COUNT, TALKING

    if msg.startswith("You fail to cut through it"):
        WEB = 0
    elif msg.startswith("You slice through the web"):
        WEB = time.time() + 0.64
    elif msg.startswith("You try to destroy the web"):
        WEB = time.time() + 2
    elif msg.startswith("He becomes an intimidating ogre"):
        CURRENT_KOLODION = 757
        log("Killed Kolodion's first form")
    elif msg.startswith("He becomes an enormous spider"):
        CURRENT_KOLODION = 758
        log("Killed Kolodion's second form")
    elif msg.startswith("He becomes an ethereal being"):
        CURRENT_KOLODION = 759
        log("Killed Kolodion's third form")
    elif msg.startswith("He becomes a vicious demon"):
        CURRENT_KOLODION = 760
        log("Killed Kolodion's fourth form")
    elif msg.startswith("..he slowly rises to his feet in his true form"):
        KILLED_KOLODION = True
        INV_READY = False
        log("Killed Kolodion's final form")
    elif msg.startswith("Cast spell successfully"):
        CAST = time.time() + 1.92
        if CHARGE_STAFF:
            CAST_COUNT += 1
        if CAST_COUNT == 10:
            log("Charging Staff 10/100")
        elif CAST_COUNT == 20:
            log("Charging Staff 20/100")
        elif CAST_COUNT == 30:
            log("Charging Staff 30/100")
        elif CAST_COUNT == 40:
            log("Charging Staff 40/100")
        elif CAST_COUNT == 50:
            log("Charging Staff 50/100")
        elif CAST_COUNT == 60:
            log("Charging Staff 60/100")
        elif CAST_COUNT == 70:
            log("Charging Staff 70/100")
        elif CAST_COUNT == 80:
            log("Charging Staff 80/100")
        elif CAST_COUNT == 90:
            log("Charging Staff 90/100")
        elif CAST_COUNT == 100:
            log("Finished Charging Staff 100/100. Time to bank")
    elif msg.startswith("You need to wait"):
        CAST = time.time() + 0.064
    elif msg.startswith("You eat"):
        EAT = time.time() + 1.28
    elif msg.startswith("It heals"):
        EAT = 0
    elif msg.startswith("You don't have all the reagents you need for this spell"):
        INV_READY = False
        log("Out of runes, banking. Logging out as the quickest route to bank")
        logout()
        return 1280
    elif msg.startswith("Well done .. you can now use the"):
        CAST_COUNT = 100
    elif msg.startswith("kolodion is busy at the moment"):
        TALKING = 0

def on_npc_message(msg, npc, player):
    global INV_READY, KILLED_KOLODION

    if player.username == get_my_player().username and msg.startswith("hello  young mage.. you're a tough one you"):
        KILLED_KOLODION = True
        INV_READY = False
    elif player.username == get_my_player().username and msg.startswith("hey there, how are you?, enjoying the bloodshed?"):
        KILLED_KOLODION = True
        INV_READY = False
