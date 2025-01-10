# Skip tutorial script by Space

# Chooses a random look for your character then skips the tutorial.

HEAD_SPRITES = [0,3,5,6,7]
BODY_SPRITES = [1,4]

wait_for_appearance_packet = False

def loop():
    global wait_for_appearance_packet

    if not wait_for_appearance_packet:
        wait_for_appearance_packet = True
        log("waiting for appearance packet")
        return 2000
    
    if is_appearance_screen():
        head_gender = random(0, 1)
        head = HEAD_SPRITES[random(0, len(HEAD_SPRITES)-1)]
        body_type = BODY_SPRITES[random(0, len(BODY_SPRITES)-1)]
        hair_colour = random(0, 9)
        top_colour = random(0, 14)
        pants_colour = random(0, 14)
        skin_colour = random(0, 4)

        log("GENDER=%d" % head_gender)
        log("HEAD=%d" % head)
        log("BODY=%d" % body_type)
        log("HAIR=%d" % hair_colour)
        log("TOP=%d" % top_colour)
        log("PANTS=%d" % pants_colour)
        log("SKIN=%d" % skin_colour)
        
        send_appearance_update(head_gender, \
                               head, \
                               body_type, \
                               hair_colour, \
                               top_colour, \
                               pants_colour, \
                               skin_colour)
        log("sent appearance")
        return 2000

    if in_rect(244, 719, 55, 49):
        skip_tutorial()
        return 2000
    else:
        log("skip tutorial complete")
        stop_account()
        return 10000
  
    return 5000