MAN          = 11
SLEEPING_BAG = 1263
MEAT         = 132
SHOP_NPCS    = [83, 55]
RECT_PT      = (135, 643)
RECT_DIMS    = (20, 6)
SHOP_PT      = (133, 643)

def loop():
    if in_combat():
        walk_to(get_x(), get_z())
        return 650

    meat = get_inventory_item_by_id(MEAT)
    if meat != None:
        use_item(meat)
        return 1300

    if has_inventory_item(SLEEPING_BAG):
        print("finished objective")
        stop_script()
        set_autologin(False)
        logout()
        return 10000
    
    coin_count = get_inventory_count_by_id(10)
    print("# Coins = %d" % coin_count)
    if coin_count >= 39:
        if not in_radius_of(SHOP_PT[0],SHOP_PT[1],8):
            walk_path_to(SHOP_PT[0],SHOP_PT[1])
            return 650
        if not is_shop_open():
            if is_option_menu():
                answer(0)
                return 3000
            else:
                npc = get_nearest_npc_by_id(ids=SHOP_NPCS, talking=False, reachable=True)
                if npc != None:
                    talk_to_npc(npc)
                    return 3000
        else:
            buy_shop_item(SLEEPING_BAG, 1)
            return 3000
    else:
        if not in_rect(RECT_PT[0], RECT_PT[1], RECT_DIMS[0], RECT_DIMS[1]):
            walk_path_to(124, 647)
            return 5000
    
        coins = get_nearest_ground_item_by_id_in_rect(10, x=RECT_PT[0], z=RECT_PT[1], width=RECT_DIMS[0], height=RECT_DIMS[1])
        if coins != None:
            pickup_item(coins)
            return 700
        
        man = get_nearest_npc_by_id_in_rect(MAN, in_combat=False, x=RECT_PT[0], z=RECT_PT[1], width=RECT_DIMS[0], height=RECT_DIMS[1])
        if man != None:
            thieve_npc(man)
            return 700
    
    return 700

