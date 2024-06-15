# Entrana Vials by O N I O N

# Start with GP in inventory somewhere around Draynor

frincos = 297
vial_id = 465
filled_vial_id = 464
vials_in_bank   = -1
this_run_banked = 0

def loop():
    total_inv = get_total_inventory_count()
    if total_inv == 30 and filled():
        return bank()
    elif total_inv == 30 and not filled():
        return fill()
    else:
        return buy()

def filled():
    unfilled = get_inventory_count_by_id(465)
    if unfilled == 0:
        return True
    else:
        return False

def fill():
    fountain = get_object_from_coords(428, 563)
    unfilled_vial = get_inventory_item_at_index(1)
    if fountain != None and unfilled_vial != None:
        use_item_on_object(unfilled_vial, fountain)
        return 600

def buy():
    shop_npc = get_nearest_npc_by_id(frincos)
    if is_option_menu() and shop_npc != None:
        answer(0)
        return 1000
    elif is_shop_open():
        buy_shop_item(vial_id, 29)
        return 400
    elif shop_npc == None:
        if distance_to(428, 565) < 20:
            walk_to(428, 565)
            return 1000
        else:
            return goToEntrana()
    else:
        shop_npc = get_nearest_npc_by_id(frincos, talking=False)
        if shop_npc != None:
            talk_to_npc(shop_npc)
            return 2000
    return 1000

def goToEntrana():
    ship = get_object_from_coords(262, 661)
    if is_option_menu():
        answer(1)
        return 1000
    elif ship != None:
        at_object(ship)
        return 1000
    else:
        walk_path_to(263, 660)
        return 1000

def bank():
    global vials_in_bank
    global this_run_banked

    ship = get_object_from_coords(419, 571)
    if in_rect(223, 634, 8, 5):
        if not is_bank_open():
            return open_bank()
        else:
            if get_total_inventory_count() != 1:
                vials_in_inventory = get_inventory_count_by_id(filled_vial_id)
                vials_in_bank = get_bank_count(filled_vial_id) + vials_in_inventory
                this_run_banked = this_run_banked + 29
                deposit(filled_vial_id, 29)
                return 600
            else:
                close_bank()
                return 1000
    elif ship != None:
        at_object(ship)
        return 1000
    else:
        walk_path_to(220, 634)
        return 1000
    return 600

def on_progress_report():
    return {"This Run": this_run_banked,
            "Total In Bank": vials_in_bank,
            "Coins Left": get_inventory_count_by_id(10)}