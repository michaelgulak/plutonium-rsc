TREES = [0, 1]
LOGS  = 14
KNIFE = 13

shaft_num = 0

def loop():
    global shaft_num

    if get_fatigue() > 95:
        use_sleeping_bag()
        return 5000
    
    if is_option_menu():
        answer(0)
        return 700

    shaft_num = get_inventory_count_by_id(280)

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
            "Shafts":            shaft_num}