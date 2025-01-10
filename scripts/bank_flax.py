# Flax picker & banker by Space

# Start under the gnome stronghold bank with an empty inv.

flax_banked = 0

def pick():
    if get_z() > 1400:
        ladder = get_object_from_coords(714, 1444)
        if ladder != None:
            at_object(ladder)
            return 1000
    else:
        if get_x() != 714 or get_z() != 501:
            walk_to(714, 501)
            return 1000
        
        flax = get_object_from_coords(714, 502)
        if flax != None:
            at_object2(flax)
            return 650
        
    return 1000

def bank():
    global flax_banked

    if get_z() < 1400:
        ladder = get_object_from_coords(714, 500)
        if ladder != None:
            at_object(ladder)
            return 1000
    else:
        if not is_bank_open():
            return open_bank()
        
        if get_total_inventory_count() == 0:
            flax_banked = get_bank_count(675)

            close_bank()
            return 1200        

        item = get_inventory_item_at_index(0)
        if item != None:
            deposit(item.id, get_inventory_count_by_id(item.id))
        
        return 1200

    return 1000

def loop():
    if get_total_inventory_count() == 30 or is_bank_open():
        return bank()
    
    return pick()

def on_progress_report():
    return {"Flax Banked": flax_banked}
