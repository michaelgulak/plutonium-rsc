# Shafts by Space

# This script cuts trees and fletches the logs into arrow shafts.
# Start with axe, knife, and sleeping bag in inventory.

TREES = [0, 1]
LOGS  = 14
KNIFE = 13

def loop():
    if get_fatigue() > 95:
        use_sleeping_bag()
        return 5000
    
    if is_option_menu():
        answer(0)
        return 700

    logs = get_inventory_item_by_id(LOGS)
    if logs != None:
        knife = get_inventory_item_by_id(KNIFE)
        if knife != None:
            use_item_with_item(knife, logs)
            return 700
    
    tree = get_nearest_object_by_id(ids=TREES)
    if tree != None:
        at_object(tree)
        return 700
    
    return 1000

    
def on_progress_report():
    return {"Woodcutting Level": get_max_stat(8), \
            "Fletching Level":   get_max_stat(9), \
            "Shafts":            get_inventory_count_by_id(280)}