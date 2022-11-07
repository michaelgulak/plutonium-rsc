# Goblin Suicide by RLN

# Kills goblins in Lumbridge and walks back if it dies.
# This script only trains defense and stops at 40 defense.

def loop():
    if get_max_stat(1) >= 40:
        set_autologin(False)
        stop_script()
        logout()
        return 5000

    if get_combat_style() != 3:
        set_combat_style(3)
        return 1000
    
    if in_combat():
        return 650
    
    if get_fatigue() > 95:
        use_sleeping_bag()
        return 2000

    if not is_inventory_item_equipped(87):
        axe = get_inventory_item_by_id(87)
        if axe != None:
            equip_item(axe)
            return 2000
    
    if not in_rect(108, 637, 16, 32):
        walk_path_to(99, 653)
        return 5000
    
    goblin = get_nearest_npc_by_id_in_rect(62, in_combat=False, x=108, z=637, width=16, height=32)
    if goblin != None:
        attack_npc(goblin)
        return 650
    
    return 2000

def on_progress_report():
    return {"Defense level": get_max_stat(1)}