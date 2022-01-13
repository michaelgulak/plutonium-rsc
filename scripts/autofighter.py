def loop():
    if get_combat_style() != settings.fight_mode:
        set_combat_style(settings.fight_mode)
        return 1000

    if in_combat():
        return 500

    if get_fatigue() >= 95:
        use_sleeping_bag()
        return 1000

    npc = get_nearest_npc_by_id(ids=settings.npc_ids, in_combat=False, reachable=True)
    if npc != None:
        attack_npc(npc)
        return 500
  
    return 500

def on_npc_despawned(npc):
    print("NPC Despawned: %d" % npc.sid)