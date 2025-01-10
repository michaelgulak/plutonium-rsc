# Goblin Suicide by Space

# Kills goblins in Lumbridge and walks back if it dies.
# This script only trains defence and stops at 40 defence.

# This script is very slow. I don't recommend you use it.

GOBLIN_POLYGON = [124, 641, 124, 622, 105, 622, 105, 629, 92, 629,
                  92, 669, 105, 669, 105, 654]

def get_nearest_goblin():
    min_dist = 999
    min_npc = None
    for i in range(get_npc_count()):
        npc = get_npc_at_index(i)
        if npc != None and npc.id == 62:
            dist = distance_to(npc.x, npc.z)
            if dist < min_dist:
                if npc.in_combat() or not point_in_polygon(npc.x, npc.z, GOBLIN_POLYGON):
                    continue
                min_dist = dist
                min_npc = npc
    
    return min_npc

def loop():
    if get_max_stat(1) >= 40:
        log("40 defence reached")
        stop_account()
        return 5000

    if get_combat_style() != 3:
        set_combat_style(3)
        return 1000
    
    if in_combat():
        return 650
    
    if get_fatigue() > 99:
        use_sleeping_bag()
        return 2000

    if not is_inventory_item_equipped(87):
        axe = get_inventory_item_by_id(87)
        if axe != None:
            equip_item(axe)
            return 2000
    
    if not point_in_polygon(get_x(), get_z(), GOBLIN_POLYGON):
        walk_path_to(99, 653)
        return 1200
    
    goblin = get_nearest_goblin()
    if goblin != None:
        attack_npc(goblin)
        return 650
    
    return 800

def on_progress_report():
    return {"Defence level": get_max_stat(1)}