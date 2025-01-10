# Man thieving script by Space

# Thieves men until 5 thieving then stops.
# The Ardougne and various men have different ids than the ones in
# Lumbridge, so you will need to edit the id if it doesn't work.

def loop():
    if get_combat_style() != 3:
        set_combat_style(3)
        return 2000

    if get_max_stat(17) >= 5:
        log("achieved 5 thieving, stopping")
        stop_account()
        return 2000

    if in_combat():
        walk_to(get_x(), get_z())
        return 650
    
    if get_fatigue() > 99:
        use_sleeping_bag()
        return 2000
    
    man = get_nearest_npc_by_id(11, in_combat=False)
    if man != None:
        thieve_npc(man)
        return 650
    
    return 1200

def on_progress_report():
    return {"Thieving Level": get_max_stat(17)}