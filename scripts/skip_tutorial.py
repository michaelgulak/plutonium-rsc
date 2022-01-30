HEAD_SPRITES = [0,3,5,6,7]
BODY_SPRITES = [1,4]

wait_for_appearance_packet = False

def loop():
    global wait_for_appearance_packet

    if not wait_for_appearance_packet:
        wait_for_appearance_packet = True
        print("waiting for appearance packet")
        return 3000
    
    if is_appearance_screen():
        head_gender = random(0, 1)
        head = HEAD_SPRITES[random(0, len(HEAD_SPRITES)-1)]
        body_type = BODY_SPRITES[random(0, len(BODY_SPRITES)-1)]
        hair_colour = random(0, 9)
        top_colour = random(0, 14)
        pants_colour = random(0, 14)
        skin_colour = random(0, 4)

        print("GENDER=%d" % head_gender)
        print("HEAD=%d" % head)
        print("BODY=%d" % body_type)
        print("HAIR=%d" % hair_colour)
        print("TOP=%d" % top_colour)
        print("PANTS=%d" % pants_colour)
        print("SKIN=%d" % skin_colour)
        
        send_appearance_update(head_gender, \
                               head, \
                               body_type, \
                               hair_colour, \
                               top_colour, \
                               pants_colour, \
                               skin_colour)
        print("sent appearance")
        return 5000

    if in_rect(244, 719, 55, 49):
        skip_tutorial()
    else:
        print("skip tutorial complete")
        logout()
        return 10000
  
    return 5000