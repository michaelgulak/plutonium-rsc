# start at gnome tree maze with 6 logs in inventory
# must be able to kill khazard warlod lvl 100
# have 2-3 sharks in inventory (OK for stats 51-51-51)
# https://gitlab.com/openrsc/openrsc/-/blob/develop/server/plugins/com/openrsc/server/plugins/authentic/quests/members/TreeGnomeVillage.java
import M_common
import time
import millz_common

walker = M_common.PathWalker()
skills = M_common.Skills()
common_ops = M_common.CommonOperations(walker)
quest_handler = M_common.QuestHandler()

QUEST_ID = 31

required_items = {
    '14': 6
}


init = True

standing_still_time = time.time()
standing_still_previous_x = None
standing_still_previous_z = None

# set to true as we don't really need to the trackers, keep as is
did_talk_to_tracker_1 = True
did_talk_to_tracker_2 = True
did_talk_to_tracker_3 = True

did_click_option_coord_4 = False
did_click_option_coord_3 = False
did_click_option_coord_5 = False

def loop():
    global init, standing_still_previous_x, standing_still_previous_z, standing_still_time
    
    if init is True:
        for item_id_str, quantity in required_items.items():
            item_id_int = int(item_id_str)
            if get_inventory_count_by_id(item_id_int) < quantity:
                log("invalid count for item : " + item_id_str + " should have qty: " + str(quantity))
                stop_account()
                
        if get_empty_slots() < 4:
            log("we need 3 empty slots")
            stop_account()
                
        init = False
    
    if in_combat() and get_hp_percent() < 60:
        walk_to(get_x(), get_z())
        return 400
    
    if millz_common.use_food():
        return 1500
    
    quest = get_quest(QUEST_ID)
    log("server quest stage: " + str(quest.stage))

    if is_quest_complete(QUEST_ID):
        return common_ops.stop("done with quest")
        return common_ops.stop("done with quest")
    
    if walker.is_walking():
        current_x = get_x()
        current_z = get_z()
        
        if quest.stage < 2:
            if standing_still_previous_x is None or standing_still_previous_z is None:
                standing_still_previous_x = current_x
                standing_still_previous_z = current_z
                standing_still_time = time.time()
            else:
                if current_x == standing_still_previous_x and current_z == standing_still_previous_z:
                    if time.time() - standing_still_time >= 15:
                        log("standing still for 15 seconds, attacking nearby npc")
                        
                        nearby_npcs = get_npcs()
                        if nearby_npcs != None:
                            npc_obj = get_nearest_npc_by_id(ids=[npc.id for npc in nearby_npcs])
                            attack_npc(npc_obj)
                            return 1000

                        standing_still_time = time.time()       
                else:
                    standing_still_previous_x = current_x
                    standing_still_previous_z = current_z
                    standing_still_time = time.time()

        return walker.walk_path()
        
        
    # (0 means that we didn't start the quest yet)
    if quest.stage == 0:
        return talk_to_bolren()
    elif quest.stage == 1:
        return talk_to_commander_montai_init()
    elif quest.stage == 2:
        return talk_to_commander_montai_logs()
    elif quest.stage == 3:
        return talk_to_trackers()
    elif quest.stage == 4:
        return fire_balista()
    elif quest.stage == 5:
        return kill_khazard_commander()
    elif quest.stage == 6:
        return kill_khazard_warlord()
        
    return 1000
  
def kill_khazard_warlord():
    if in_combat():
        return 1000
  
    if has_inventory_item(741) and common_ops._object_is_reachable(69):
        return quest_handler.talk_to_npc_with_dialogue(
                walker,
                npc_id=400,
                menu_options=[],
                fallback_path=(656,696)
        )

    if not common_ops._is_npc_nearby(410) and not has_inventory_item(741):
        walker.set_path(664,606)
        return 1000

    if not has_inventory_item(741):
        npc_obj = get_nearest_npc_by_id(410)
        attack_npc(npc_obj)
        return 1000
    
    return quest_handler.talk_to_npc_with_dialogue(
        walker,
        npc_id=396,
        menu_options=["Yes please"],
        fallback_path=(624,675)
    )
    
    return 1000
    
def kill_khazard_commander():
    # back in maze by teleport
    if common_ops._object_is_reachable(69):
        return quest_handler.talk_to_npc_with_dialogue(
                walker,
                npc_id=400,
                menu_options=["I will find the warlord and bring back the orbs"],
                fallback_path=(656,696)
        )

    # not in room with commander
    if not common_ops._object_is_reachable(206) and get_z() < 1000 and not has_inventory_item(740):
        return common_ops._at_object(658,632,659,633)
        
    if not has_inventory_item(740):
        if get_z() < 1000:
            return common_ops._at_object(665,635)
        elif get_z() > 1000:
            return common_ops._at_object(662,1578)
    
    if has_inventory_item(740):
        if get_z() > 1000:
            return common_ops._at_object(665,1579)
        
        if common_ops._object_is_reachable(206):
            return common_ops._at_object(658,632,659,631)
            
        return quest_handler.talk_to_npc_with_dialogue(
            walker,
            npc_id=396,
            menu_options=["Yes please"],
            fallback_path=(624,675)
        )
        
    return 1000
        
def fire_balista():
    global did_click_option_coord_4, did_click_option_coord_3, did_click_option_coord_5
    
    if is_option_menu() is False:
        return common_ops._at_object(657, 663)
    
    if did_click_option_coord_4 is False:
        answer(3)
        did_click_option_coord_4 = True
        return 3000
    
    if did_click_option_coord_3 is False:
        answer(2)
        did_click_option_coord_3 = True
        return 3000
        
    if did_click_option_coord_5 is False:
        answer(4)
        did_click_option_coord_5 = True
        return 3000
    
    
    return 1000
  
def talk_to_trackers():
    if did_talk_to_tracker_1 is False:
        return quest_handler.talk_to_npc_with_dialogue(
                walker,
                npc_id=404,
                menu_options=[],
                fallback_path=(649,626)
        )
        
    if did_talk_to_tracker_3 is False:
        return quest_handler.talk_to_npc_with_dialogue(
                walker,
                npc_id=406,
                menu_options=[],
                fallback_path=(668,633)
        )
        
    if did_talk_to_tracker_2 is False:
        return common_ops._at_wall_object(648,629,650,630)

    return quest_handler.talk_to_npc_with_dialogue(
            walker,
            npc_id=408,
            menu_options=["I'll try my best"],
            fallback_path=(658,660)
    )
    
def talk_to_commander_montai_logs():
    return quest_handler.talk_to_npc_with_dialogue(
            walker,
            npc_id=408,
            menu_options=[],
            fallback_path=(658,660)
    )
  
def talk_to_commander_montai_init():
    return quest_handler.talk_to_npc_with_dialogue(
            walker,
            npc_id=408,
            menu_options=["Ok, i'll gather some wood"],
            fallback_path=(658,660)
    )
  
def talk_to_bolren():
    if not common_ops._object_is_reachable(69):
        if not at(633,705):
            walker.set_path(633,705)
            return 1000
        # if we are at fence location    
        return common_ops._at_wall_object(633,705,633,705)
    
    return quest_handler.talk_to_npc_with_dialogue(
            walker,
            npc_id=400,
            menu_options=["Can i help at all?", "I would be glad to help"],
            fallback_path=(656,696)
    )
    
    return 1000
    
def on_npc_message(msg, npc, player):
    global did_talk_to_tracker_1, did_talk_to_tracker_2, did_talk_to_tracker_3
    
    if get_pid() != player.pid:
        return
    
    if npc.id == 404 and msg == "ok, take care":
        did_talk_to_tracker_1 = True
    
    if npc.id == 406 and msg == "The poor gnome has gone mad":
        did_talk_to_tracker_3 = True    
        
    if npc.id == 405 and msg == "go":
        did_talk_to_tracker_2 = True