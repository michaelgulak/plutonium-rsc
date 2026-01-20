# Quest script template
# Author: You
# Description: Shared base logic to prepare quest requirements before starting
# Requirements: Handles bank/inventory checks, fetch logic, and initialization.

import M_common
import time
import millz_common

# === Game APIs ===
walker = M_common.PathWalker()
common_ops = M_common.CommonOperations(walker)
quest_handler = M_common.QuestHandler()
spell_book = M_common.SpellBook()
skills = M_common.Skills()
areas = M_common.Areas()
npcs = M_npcs.NpcId()

# === Quest Configuration ===
QUEST_ID = 22
REQUIRED_ITEMS = {
    '420': 1,   #Anti-dragon shield
    '566': 1,   #Cure poison potion
    '125': 1,   # test
}

# === Spells ===
falador_teleport = spell_book.get_spell("Falador Teleport")
camelot_teleport = spell_book.get_spell("Camelot teleport")

# Merge spell runes and basic required items into one flat dict
required_items = spell_book.merge_item_requirements(
    falador_teleport.get_runes_with_quantities(),
    camelot_teleport.get_runes_with_quantities(),
    REQUIRED_ITEMS
)

# === Stages ===
STAGE_CHECK_ITEMS_BANK = 1
STAGE_CHECK_INVENTORY_ITEMS = 2
STAGE_FETCH_MISSING_ITEMS = 3
STAGE_START_QUEST = 4

# === Globals ===
current_sub_stage = STAGE_CHECK_ITEMS_BANK
init = True
bank = None

# === Main Loop ===
def loop():
    global init, bank

    if init:
        bank = millz_common.get_bank_by_name("Seers Village")

        common_ops.minimum_slots(10)
        falador_teleport.stop_if_cannot_cast()
        camelot_teleport.stop_if_cannot_cast()
        skills.has_skill_requirement("Crafting", 40)
        
        init = False

    if is_quest_complete(QUEST_ID):
        log("Quest already complete.")
        stop_account()
        return 100

    quest = get_quest(QUEST_ID)
    log("Current server quest stage: " + str(quest.stage))

    if walker.is_walking():
        return walker.walk_path()

    if quest.stage == 0:
        if current_sub_stage == STAGE_CHECK_ITEMS_BANK:
            return quest_handler.check_items_in_bank(walker, bank, required_items, set_substage, STAGE_CHECK_INVENTORY_ITEMS)
        elif current_sub_stage == STAGE_CHECK_INVENTORY_ITEMS:
            return quest_handler.check_inventory_items(required_items, set_substage, STAGE_FETCH_MISSING_ITEMS, STAGE_START_QUEST)
        elif current_sub_stage == STAGE_FETCH_MISSING_ITEMS:
            # at this point we have the required_items map with items we still need to fetch
            return fetch_missing_items()
        elif current_sub_stage == STAGE_START_QUEST:
            return start_quest_logic()
    else:
        log("custom operations here")
       

    return 1000


def start_quest_logic():
    log("All required items gathered. Ready to start the quest.")
    stop_account()
    return 1000

def fetch_missing_items():
    global required_items

    log("Fetching missing items: " + str(required_items))

    fetch_priority = [
        ITEM_ID_FISHING_ROD,
        ITEM_ID_BAIT,
        ITEM_ID_BUCKET_OF_MILK,
        ITEM_ID_RAW_SARDINE
    ]

    for item_id in fetch_priority:
        item_id_str = str(item_id)
        if item_id_str not in required_items:
            continue

        quantity_needed = required_items[item_id_str]

        if get_inventory_count_by_id(item_id) >= quantity_needed:
            del required_items[item_id_str]
            continue

        if item_id == ITEM_ID_FISHING_ROD:
            return M_common.buy(walker, 277, 648, 280, 645, 4, 5, 167, 0, 377, quantity_needed)
        elif item_id == ITEM_ID_BAIT:
            qty_missing = quantity_needed - get_inventory_count_by_id(380)
            return M_common.buy(walker, 277, 648, 280, 645, 4, 5, 167, 0, 380, qty_missing)
        elif item_id == ITEM_ID_RAW_SARDINE:
            if (get_inventory_count_by_id(ITEM_ID_FISHING_ROD) >= 1 and get_inventory_count_by_id(ITEM_ID_BAIT) >= 1):
                return common_ops._at_object2(224, 659, 223, 659)
            else:
                log("Need rod and bait first")
                return 1000
        elif item_id == ITEM_ID_BUCKET_OF_MILK:
            return common_ops.get_lumbridge_bucket_of_milk()
        elif item_id == 1100:
            log("Fetching doggle leaves...")
            return 1000

    log("All missing items fetched. Proceeding to quest.")
    set_substage(STAGE_START_QUEST)
    return 1000

def on_server_message(msg):
    msg_lower = msg.lower()


def set_substage(stage):
    global current_sub_stage
    current_sub_stage = stage