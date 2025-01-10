# Ardougne cake thieving script by Space

# Start with sleeping bag near stall in ardy.
# Optionally bring armour.

CAKE_IDS = [335,333,330]

cake_stolen = 0

def bank():
    global cake_stolen

    if not in_rect(554, 609, 4, 8):
        if in_radius_of(550, 612, 15):
            door = get_object_from_coords(550, 612)
            if door != None and door.id == 64:
                at_object(door)
                return 1300

        walk_path_to(551, 612)
        return 5000
    
    if not is_bank_open():
        return open_bank()
    
    cake_count = get_inventory_count_by_id(330)
    cake_stolen = get_bank_count(330) + cake_count

    if cake_count > 1:
        deposit(330, cake_count-1)
        return 2000
    else:
        close_bank()
        return 2000
    
    return 5000

def thieve():
    if get_combat_style() != 3:
        set_combat_style(3)
        return 2000
    
    if in_combat():
        walk_to(get_x(), get_z())
        return 650

    if get_fatigue() > 99:
        use_sleeping_bag()
        return 2000
    
    if get_current_stat(3) <= 9:
        cake = get_inventory_item_by_id(ids=CAKE_IDS)
        if cake != None:
            use_item(cake)
            return 1350
        else:
            return 1000
    
    if get_x() != 543 or get_z() != 600:
        if in_radius_of(550, 612, 15):
            door = get_object_from_coords(550, 612)
            if door != None and door.id == 64:
                at_object(door)
                return 1300

        walk_path_to(543, 600)
        return 5000

    stall = get_nearest_object_by_id(322)
    if stall != None:
        at_object2(stall)
        return 750
    
    return 1000

def loop():
    if is_bank_open() or get_total_inventory_count() == 30:
        return bank()
    
    return thieve()

def on_progress_report():
    return {"Thieving Level": get_max_stat(17),
            "Cake Stolen":    cake_stolen}