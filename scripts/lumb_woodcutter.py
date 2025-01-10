# Lumbridge woodcutter by Space

# This script cuts trees between lumbridge and draynor.
# Start with axe and sleeping bag.

import time

TREES = [0, 1]

click_timeout = 0

def loop():
    global click_timeout

    if get_fatigue() > 99:
        use_sleeping_bag()
        return 2000

    if not in_rect(214, 622, 54, 58):
        walk_path_to(167, 646)
        return 3000

    if click_timeout != 0 and time.time() <= click_timeout:
        return 250

    tree = get_nearest_object_by_id_in_rect(ids=TREES, x=214, z=622, width=54, height=58)
    if tree != None:
        at_object(tree)
        click_timeout = time.time() + 5
        return 700
    
    return 500

def on_server_message(msg):
    global click_timeout

    if msg.startswith("You slip") or msg.startswith("You get"):
        click_timeout = 0

def on_progress_report():
    return {"Woodcutting Level": get_max_stat(8)}