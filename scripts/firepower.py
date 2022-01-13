LOGS      = 14
TINDERBOX = 166
TREES     = [0, 1]

def loop():
    if get_fatigue() > 95:
        use_sleeping_bag()
        return 2000

    logs = get_nearest_ground_item_by_id(LOGS)
    if logs != None \
        and logs.x == get_x() \
        and logs.z == get_z() \
        and not is_object_at(logs.x, logs.z):

        box = get_inventory_item_by_id(TINDERBOX)
        if box != None:
            use_item_on_ground_item(box, logs)
            return 700
    
    tree = get_nearest_object_by_id(ids=TREES, reachable=True)
    if tree != None:
        at_object(tree)
        return 700
    
    return 500