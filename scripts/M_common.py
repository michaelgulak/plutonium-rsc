# common lib, some parts are taken from other scripts
import time

# global constants
talk_timeout = 0

class Herb:
    def __init__(self, unidentified_id, identified_id, level, name):
        """
        Initialize a Herb object.
        :param unidentified_id: The item ID of the unidentified herb.
        :param identified_id: The item ID of the identified herb.
        :param level: The required Herblaw level to identify it.
        :param name: Name of the herb.
        """
        self.unidentified_id = unidentified_id
        self.identified_id = identified_id
        self.level = level
        self.name = name

    def can_identify(self, herblaw_level):
        """
        Check if the given Herblaw level is enough to identify this herb.
        :param herblaw_level: Current Herblaw level (int).
        :return: True if the level is sufficient.
        """
        return herblaw_level >= self.level


class HerbBook:
    def __init__(self):
        """
        Initialize the HerbBook and populate it with herbs.
        """
        self.herbs_by_unid = {}   # Keys must be strings
        self.herbs_by_name = {}
        self.initialize_herbs()

    def initialize_herbs(self):
        """
        Add predefined herbs to the herb book.
        """
        self.add_herb(165, 444, 3, "Guam leaf")
        self.add_herb(435, 445, 5, "Marrentill")
        self.add_herb(436, 446, 12, "Tarromin")
        self.add_herb(437, 447, 22, "Harralander")
        self.add_herb(438, 448, 30, "Ranarr weed")
        self.add_herb(439, 449, 45, "Irit leaf")
        self.add_herb(440, 450, 50, "Avantoe")
        self.add_herb(441, 451, 55, "Kwuarm")
        self.add_herb(442, 452, 66, "Cadantine")
        self.add_herb(443, 453, 72, "Dwarf weed")

    def add_herb(self, unidentified_id, identified_id, level, name):
        herb = Herb(unidentified_id, identified_id, level, name)
        self.herbs_by_unid[str(unidentified_id)] = herb  # Use string key
        self.herbs_by_name[name.lower()] = herb

    def get_herb_by_unidentified_id(self, item_id):
        """
        Get a Herb by its unidentified item ID.
        :param item_id: int or str
        :return: Herb or None
        """
        return self.herbs_by_unid.get(str(item_id))  # Convert key to string

    def get_herb_by_name(self, name):
        return self.herbs_by_name.get(name.lower())

    def get_all_unidentified_ids(self):
        """
        Get a list of all unidentified herb item IDs as strings.
        :return: List[str]
        """
        return list(self.herbs_by_unid.keys())

    def get_all_identified_ids(self):
        """
        Get a list of all identified herb item IDs as strings.
        :return: List[str]
        """
        return [str(herb.identified_id) for herb in self.herbs_by_unid.values()]

    def get_identifiable_herbs(self, herblaw_level):
        """
        Get a list of Herb objects that can be identified at the given Herblaw level.
        :param herblaw_level: The player's current Herblaw level (int).
        :return: List[Herb]
        """
        return [herb.unidentified_id for herb in self.herbs_by_unid.values() if herb.can_identify(herblaw_level)]

    def get_low_level_herb_ids(self, max_level=29):
        """
        Get a list of identified herb item IDs for herbs under a certain level.
        :param max_level: Max herblaw level (exclusive).
        :return: List[int]
        """
        return [herb.identified_id for herb in self.herbs_by_unid.values() if herb.level <= max_level]

        

class Coords:
    def __init__(self, id, x, z):
        self.id = id
        self.x = x
        self.z = z
#Area(north_east_x, north_east_z, south_west_x, south_west_z)   
class Area:
    def __init__(self, start_x, start_z, end_x, end_z, custom_contains=None):
        self.start_x = start_x
        self.start_z = start_z
        self.end_x = end_x
        self.end_z = end_z
        self.custom_contains = custom_contains  # optional function override

    def contains(self, x, z):
        # if custom override is provided, use it
        if self.custom_contains:
            return self.custom_contains(x, z)

        # default behavior
        return (min(self.start_x, self.end_x) <= x <= max(self.start_x, self.end_x) and
                min(self.start_z, self.end_z) <= z <= max(self.start_z, self.end_z))
                
class Areas:
    HERO_QUEST_RANGE_ROOM = Area(459, 672, 460, 673)
    TUTORIAL_ISLAND = Area(190, 720, 240, 770)
    BLACK_HOLE = Area(303, 3298, 307, 3302)
    TUTORIAL_LANDING = Area(214, 739, 221, 747)
    TUTORIAL_RAT_ZONE = Area(226, 728, 234, 738)
    MOD_ROOM = Area(64, 1639, 80, 1643)
    VARROCK = [Area(78, 490, 175, 537), Area(92, 444, 150, 490)]
    EDGEVILLE = [Area(198, 427, 229, 450), Area(208, 451, 227, 472)]
    BARB_VILLAGE = Area(209, 491, 247, 529)
    DRAYNOR = Area(210, 608, 233, 659)
    LUMBRIDGE = Area(108, 620, 147, 670)
    AL_KHARID = Area(48, 659, 96, 703)
    FALADOR = Area(245, 531, 341, 583)
    PORT_SARIM = Area(246, 621, 286, 670)
    TAVERLY = Area(343, 454, 389, 512)
    ENTRANA = Area(395, 525, 441, 573)
    CATHERBY = Area(415, 475, 456, 508)
    SEERS = Area(486, 438, 534, 482)
    GNOME_STRONGHOLD = Area(673, 432, 751, 537)
    ARDOUGNE = Area(500, 537, 600, 708)
    YANILLE = Area(577, 741, 647, 767)
    BRIMHAVEN = Area(435, 644, 477, 709)
    KARAMJA = Area(323, 644, 679, 908)
    SHILO_VILLAGE = Area(384, 817, 430, 860)
    SEERS_PARTY_HALL_UP = Area(490, 1408, 500, 1415)
    SEERS_PARTY_HALL_DOWN = Area(490, 464, 500, 471)
    
    def in_hero_quest_range_room(self, x, z):
        return Areas.HERO_QUEST_RANGE_ROOM.contains(x, z)
    
    def on_tutorial_island(self, x, z):
        return Areas.TUTORIAL_ISLAND.contains(x, z)
    
    def on_black_hole(self, x, z):
        return Areas.BLACK_HOLE.contains(x, z)
    
    def in_tutorial_landing(self, x, z):
        return Areas.TUTORIAL_LANDING.contains(x, z)
    
    def around_tutorial_rat_zone(self, x, z):
        return Areas.TUTORIAL_RAT_ZONE.contains(x, z)
    
    def in_mod_room(self, x, z):
        return Areas.MOD_ROOM.contains(x, z)
    
    def in_varrock(self, x, z):
        return any(area.contains(x, z) for area in Areas.VARROCK)
    
    def in_edgeville(self, x, z):
        return any(area.contains(x, z) for area in Areas.EDGEVILLE)
    
    def in_barb_village(self, x, z):
        return Areas.BARB_VILLAGE.contains(x, z)
    
    def in_draynor(self, x, z):
        return Areas.DRAYNOR.contains(x, z)
    
    def in_lumbridge(self, x, z):
        return Areas.LUMBRIDGE.contains(x, z)
    
    def in_al_kharid(self, x, z):
        return Areas.AL_KHARID.contains(x, z)
    
    def in_falador(self, x, z):
        return Areas.FALADOR.contains(x, z)
    
    def in_port_sarim(self, x, z):
        return Areas.PORT_SARIM.contains(x, z)
    
    def in_taverly(self, x, z):
        return Areas.TAVERLY.contains(x, z)
    
    def in_entrana(self, x, z):
        return Areas.ENTRANA.contains(x, z)
    
    def in_catherby(self, x, z):
        return Areas.CATHERBY.contains(x, z)
    
    def in_seers(self, x, z):
        return Areas.SEERS.contains(x, z)
    
    def in_gnome_stronghold(self, x, z):
        return Areas.GNOME_STRONGHOLD.contains(x, z)
    
    def in_ardougne(self, x, z):
        return Areas.ARDOUGNE.contains(x, z)
    
    def in_yanille(self, x, z):
        return Areas.YANILLE.contains(x, z)
    
    def in_brimhaven(self, x, z):
        return Areas.BRIMHAVEN.contains(x, z)
    
    def in_karamja(self, x, z):
        return Areas.KARAMJA.contains(x, z)
    
    def in_shilo_village(self, x, z):
        return Areas.SHILO_VILLAGE.contains(x, z)
    
    def is_in_seers_party_hall(self, x, z):
        return Areas.SEERS_PARTY_HALL_UP.contains(x, z) or Areas.SEERS_PARTY_HALL_DOWN.contains(x, z)

class Coordinate:
    def __init__(self, x, z):
        self.x = x
        self.z = z

class ItemIDS:
    FIRE_RUNE = 31
    WATER_RUNE = 32
    AIR_RUNE = 33
    EARTH_RUNE = 34
    LAW_RUNE = 42
    DEATH_RUNE = 38
    AIR_STAFF = 101
    WATER_STAFF = 102
    EARTH_STAFF = 103
    FIRE_STAFF = 197
    
    # just an example
    def get_rune_ids():
        return [ItemIDs.FIRE_RUNE, ItemIDs.WATER_RUNE, ItemIDs.AIR_RUNE, ItemIDs.EARTH_RUNE, ItemIDs.LAW_RUNE]


STAFF_RUNE_MAP = {
    str(ItemIDS.AIR_STAFF): ItemIDS.AIR_RUNE,
    str(ItemIDS.FIRE_STAFF): ItemIDS.FIRE_RUNE,
    str(ItemIDS.WATER_STAFF): ItemIDS.WATER_RUNE,
    str(ItemIDS.EARTH_STAFF): ItemIDS.EARTH_RUNE
}

class Spell:
    def __init__(self, name, index, runes, min_magic_level):
        """
        Initialize a Spell object.
        :param name: The name of the spell (str).
        :param index: The predefined index for the spell (int).
        :param runes: A dictionary of rune item IDs and their counts {item_id: count}.
        """
        self.name = name
        self.index = index
        self.runes = runes
        self.min_magic_level = min_magic_level

    def get_runes(self):
        """
        Get an array of rune item IDs required for this spell.
        :return: A list of rune item IDs (ints).
        """
        return list(self.runes.keys())
        
    def get_runes_with_quantities(self):
        """
        Get a dictionary of rune item IDs and their quantities required for this spell.
        :return: A dictionary {item_id: count}.
        """
        return self.runes
        

    def has_all_runes(self):
        """
        Check if the inventory has all required runes for this spell,
        taking into account any equipped staff that provides unlimited runes.
        """
        for rune_id_str, required_quantity in self.runes.items():
            rune_id = int(rune_id_str)

            # skip inventory check if we have a staff that provides this rune equipped
            skip_check = False
            for staff_id_str, provided_rune in STAFF_RUNE_MAP.items():
                if is_inventory_item_equipped(int(staff_id_str)) and rune_id == provided_rune:
                    skip_check = True
                    break

            if not skip_check and get_inventory_count_by_id(rune_id) < required_quantity:
                return False

        return True

        
    def can_cast(self):
        """
        Check if the spell can be cast based on the player's magic level.
        :return: True if the player has the required magic level, False otherwise.
        """
        return get_max_stat(6) >= self.min_magic_level
        
    def stop_if_cannot_cast(self):
        if get_max_stat(6) < self.min_magic_level:
            log("cannot cast: " + name)
            stop_account()
            stop_script()

    def _cast_on_self(self):
        """
        Attempt to cast this spell on the player themselves.

        This method checks whether the spell can be cast by verifying:
        - The player's Magic level is high enough.
        - All required runes are present in the inventory.

        If both conditions are met, it calls the global `cast_on_self(index)` function with the spell's index.

        :return: time to wait
        """
        if self.can_cast() and self.has_all_runes():
            cast_on_self(self.index)

        return 3000
        
class SpellBook:
    def __init__(self):
        """
        Initialize the SpellBook and populate it with spells.
        """
        self.spells = {}
        self.initialize_spells()

    def initialize_spells(self):
        """
        Add predefined spells to the spellbook.
        """
        self.add_spell("Varrock teleport", 12, {str(ItemIDS.FIRE_RUNE): 1, str(ItemIDS.AIR_RUNE): 1, str(ItemIDS.LAW_RUNE): 1}, 25)
        self.add_spell("Lumbridge teleport", 15, {str(ItemIDS.EARTH_RUNE): 1, str(ItemIDS.AIR_RUNE): 1, str(ItemIDS.LAW_RUNE): 1}, 31)
        self.add_spell("Camelot teleport", 22, {str(ItemIDS.AIR_RUNE): 5, str(ItemIDS.LAW_RUNE): 1}, 45)
        self.add_spell("Ardougne teleport", 26, {str(ItemIDS.WATER_RUNE): 2, str(ItemIDS.LAW_RUNE): 2}, 51)
        self.add_spell("Watchtower teleport", 31, {str(ItemIDS.EARTH_RUNE): 2, str(ItemIDS.LAW_RUNE): 2}, 58)
        self.add_spell("Falador teleport", 18, {str(ItemIDS.WATER_RUNE): 1, str(ItemIDS.AIR_RUNE): 3, str(ItemIDS.LAW_RUNE): 1}, 37)

    def add_spell(self, name, index, rune_counts, min_magic_level):
        """
        Add a spell to the spellbook.
        :param name: The name of the spell (str).
        :param index: The predefined index for the spell (int).
        :param rune_counts: A dictionary of rune item IDs and their counts {item_id: count}.
        :param min_magic_level: The minimum magic level required to cast the spell (int).
        """
        rune_counts_str = {str(item_id): count for item_id, count in rune_counts.items()}
        
        spell = Spell(name, index, rune_counts_str, min_magic_level)
        self.spells[name.lower()] = spell

    def get_spell(self, name):
        """
        Retrieve a spell by name.
        :param name: The name of the spell (str).
        :return: A Spell object, or None if the spell does not exist.
        """
        return self.spells.get(name.lower())

    def merge_item_requirements(self, *item_dicts):
        """
        Merge multiple dictionaries of item requirements into one.
        This is useful for combining rune requirements from multiple spells,
        or mixing in other required items like quest items.

        Each dictionary must be of the form {item_id (str): quantity (int)}.

        :param item_dicts: One or more dictionaries to merge.
        :return: A single dictionary with combined quantities.
        """
        combined = {}
        for item_dict in item_dicts:
            for item_id in item_dict:
                quantity = item_dict[item_id]
                if item_id in combined:
                    combined[item_id] += quantity
                else:
                    combined[item_id] = quantity
        return combined


class CommonOperations:
    """
    A utility class for common RuneScape Classic bot operations.
    
    Provides methods to handle various in-game interactions like picking up items,
    using items on objects or NPCs, and interacting with the environment.
    """
    def __init__(self, walker):
        """
        Initializes the CommonOperations instance.

        Parameters:
            walker (PathWalker): A PathWalker instance for handling pathfinding and movement.
        """
        self.walker = walker
        self.talk_timeout = 0
        
    def stop(self, msg):
        log(msg)
        stop_account()
        stop_script()
        logout()
        
        return 640
    
    def minimum_slots(self, slots):
        if get_empty_slots() < slots:
            log("we need : " + str(slots) + " empty slots available in our inventory")
            stop_script()
            stop_account()
    
    def _unequip_item(self, item_id):
        obj = get_inventory_item_by_id(item_id)
        if obj is not None:
            unequip_item(obj)
        else:
            log("tried to unequip an non existing item")
        
        return 1000
    
    def _equip_item(self, item_id):
        obj = get_inventory_item_by_id(item_id)
        if obj is not None:
            if obj.equipped is False:
                equip_item(obj)
        else:
            log("tried to equip an non existing item")
        
        return 1000
        
    def _wall_object_is_reachable(self, wall_object_id):
        obj = get_nearest_wall_object_by_id(wall_object_id, reachable=True)
        return obj is not None
        
    def _object_is_reachable(self, object_id):
        obj = get_nearest_object_by_id(object_id, reachable=True)
        return obj is not None
        
    def _object_is_reachable_coords(self, x, z):
        obj = get_object_from_coords(x, z, reachable=True)
        return obj is not None
    
    def _npc_is_reachable(self, npc_id):
        npc = get_nearest_npc_by_id(npc_id, reachable=True)
        return npc is not None
        
    def _is_npc_nearby(self, npc_id):
        npc = get_nearest_npc_by_id(npc_id)
        return npc is not None
        
    def _pickup_item(self, item_id, fallback_x=None, fallback_z=None):
        """
        Attempts to pick up a ground item with the given ID. If the item is not nearby,
        the bot moves to the fallback coordinates, if provided.

        Parameters:
            item_id (int): The ID of the item to pick up.
            fallback_x (int, optional): The fallback X coordinate if the item isn't nearby.
            fallback_z (int, optional): The fallback Z coordinate if the item isn't nearby.

        Returns:
            int: Delay in milliseconds after the action.
        """
        log("pickup item: " + str(item_id))
        item_obj = get_nearest_ground_item_by_id(item_id)
        if item_obj == None and fallback_x is not None and fallback_z is not None:
            if at(fallback_x, fallback_z):
                return 1000
            else:
                self.walker.set_path(fallback_x, fallback_z)
        else:
            if item_obj != None:
                pickup_item(item_obj)
            
        return 1000
        
    def _use_item_with_item(self, item_id_1, item_id_2):
        """
        Uses one inventory item with another.

        Parameters:
            item_id_1 (int): The ID of the first item.
            item_id_2 (int): The ID of the second item.

        Returns:
            int: Delay in milliseconds after the action.
        """
        log("use item with item: " + str(item_id_1) + " and " + str(item_id_2))
        inventory_item_obj_1 = get_inventory_item_by_id(item_id_1)
        inventory_item_obj_2 = get_inventory_item_by_id(item_id_2)
        
        if inventory_item_obj_1 is not None and inventory_item_obj_2 is not None:
            use_item_with_item(inventory_item_obj_1, inventory_item_obj_2)
        else:
            log("trying to use item_id: " + str(item_id_1) + " with: " + str(item_id_2) + " but did not succeed.")
        
        return 2500
        
    def _at_wall_object(self, x, z, fallback_x=None, fallback_z=None, menu_options=None):
        """
        Interacts with a wall object at the given coordinates and optionally answers dialogue options.

        Parameters:
            x (int): The X coordinate of the wall object.
            z (int): The Z coordinate of the wall object.
            fallback_x (int, optional): The fallback X coordinate if the object isn't found.
            fallback_z (int, optional): The fallback Z coordinate if the object isn't found.
            menu_options (list, optional): Dialogue options to handle if a menu appears.

        Returns:
            int: Delay in milliseconds after the action.
        """
        log("at wall object at: (" + str(x) + ", " + str(z) + ")")

        # Handle dialogue timeout
        if menu_options:
            if self.talk_timeout != 0 and time.time() < self.talk_timeout:
                return 250

            if is_option_menu():
                for opt in menu_options:
                    idx = get_option_menu_index(opt)
                    if idx != -1:
                        answer(idx)
                        self.talk_timeout = time.time() + 5
                        return 2000

        # Interact with wall object
        obj = get_wall_object_from_coords(x, z)
        if obj is not None:
            at_wall_object(obj)
            return 5000
        else:
            if fallback_x is not None and fallback_z is not None:
                self.walker.set_path(fallback_x, fallback_z)
            return 1000

    def _at_wall_object2(self, x, z, fallback_x=None, fallback_z=None):
        """
        Interacts with a wall object at the given coordinates. If the object is not found,
        the bot moves to the fallback coordinates, if provided.

        Parameters:
            x (int): The X coordinate of the wall object.
            z (int): The Z coordinate of the wall object.
            fallback_x (int, optional): The fallback X coordinate if the object isn't found.
            fallback_z (int, optional): The fallback Z coordinate if the object isn't found.

        Returns:
            int: Delay in milliseconds after the action.
        """
        log("at wall object2 with id: at: (" + str(x) + ", " + str(z) + ")")
        obj = get_wall_object_from_coords(x, z)          
        if obj is not None:
            at_wall_object2(obj)
            return 5000
        else:
            if fallback_x is not None and fallback_z is not None:
                self.walker.set_path(fallback_x, fallback_z)
            return 1000
            
    def _at_object(self, x, z, fallback_x=None, fallback_z=None, menu_options=None):
        """
        Interacts with a static object at the given coordinates. If the object is not found,
        the bot moves to the fallback coordinates, if provided.

        Parameters:
            x (int): The X coordinate of the object.
            z (int): The Z coordinate of the object.
            fallback_x (int, optional): The fallback X coordinate if the object isn't found.
            fallback_z (int, optional): The fallback Z coordinate if the object isn't found.

        Returns:
            int: Delay in milliseconds after the action.
        """
        log("at object at: (" + str(x) + ", " + str(z) + ")")
        
        # Handle timeout
        if menu_options:
            if self.talk_timeout is not 0 and time.time() < self.talk_timeout:
                return 250

            if menu_options and is_option_menu():
                for opt in menu_options:
                    idx = get_option_menu_index(opt)
                    if idx != -1:
                        answer(idx)
                        self.talk_timeout = time.time() + 5
                        return 2000
        
        obj = get_object_from_coords(x, z)
        if obj is not None:
            at_object(obj)
            return 3500
        else:
            if fallback_x is not None and fallback_z is not None:
                self.walker.set_path(fallback_x, fallback_z)
            return 1000
            
    def _at_object2(self, x, z, fallback_x=None, fallback_z=None):
        """
        Interacts with a static object at the given coordinates. If the object is not found,
        the bot moves to the fallback coordinates, if provided.

        Parameters:
            x (int): The X coordinate of the object.
            z (int): The Z coordinate of the object.
            fallback_x (int, optional): The fallback X coordinate if the object isn't found.
            fallback_z (int, optional): The fallback Z coordinate if the object isn't found.

        Returns:
            int: Delay in milliseconds after the action.
        """
        log("at object at: (" + str(x) + ", " + str(z) + ")")
        obj = get_object_from_coords(x, z)
        if obj is not None:
            at_object2(obj)
            return 3500
        else:
            if fallback_x is not None and fallback_z is not None:
                self.walker.set_path(fallback_x, fallback_z)
            return 1000

    def _drop_item(self, inventory_item_id):
        inventory_item_obj = get_inventory_item_by_id(inventory_item_id)
        if inventory_item_obj is not None:
            drop_item(inventory_item_obj)
        else:
            log("tried to drop nonexisting item")

        return 2000

    def _use_item(self, inventory_item_id):
        inventory_item_obj = get_inventory_item_by_id(inventory_item_id)
        if inventory_item_obj is not None:
            use_item(inventory_item_obj)
        else:
            log("tried to use non existing item")
            
        return 1000

    def _use_item_on_npc(self, inventory_item_id, npc_id, fallback_x=None, fallback_z=None, should_be_reachable=True):
        """
        Uses an inventory item on a nearby NPC. If the NPC is not nearby, the bot
        moves to the fallback coordinates, if provided.

        Parameters:
            inventory_item_id (int): The ID of the inventory item to use.
            npc_id (int): The ID of the target NPC.
            fallback_x (int, optional): The fallback X coordinate if the NPC isn't found.
            fallback_z (int, optional): The fallback Z coordinate if the NPC isn't found.

        Returns:
            int: Delay in milliseconds after the action.
        """
        log("use item: " + str(inventory_item_id) + " on npc with id: " + str(npc_id))
        inventory_item_obj = get_inventory_item_by_id(inventory_item_id)
        if inventory_item_obj is not None:
            npc_obj = get_nearest_npc_by_id(npc_id, reachable=should_be_reachable)
            if npc_obj == None and fallback_x is not None and fallback_z is not None:
                self.walker.set_path(fallback_x, fallback_z)
            elif npc_obj == None:
                log("npc is not nearby.. waiting..")
            elif npc_obj is not None:
                use_item_on_npc(inventory_item_obj,npc_obj)
                return 1500
                
        return 1000
                
    def _use_item_on_object(self, inventory_item_id, obj_x, obj_z, fallback_x=None, fallback_z=None):
        """
        Uses an inventory item on a static object at the given coordinates. If the
        object is not found, the bot moves to the fallback coordinates, if provided.

        Parameters:
            inventory_item_id (int): The ID of the inventory item to use.
            obj_x (int): The X coordinate of the object.
            obj_z (int): The Z coordinate of the object.
            fallback_x (int, optional): The fallback X coordinate if the object isn't found.
            fallback_z (int, optional): The fallback Z coordinate if the object isn't found.

        Returns:
            int: Delay in milliseconds after the action.
        """
        log("use item: " + str(inventory_item_id) + " on object at: (" + str(obj_x) + ", " + str(obj_z) + ")")
        obj = get_object_from_coords(obj_x, obj_z)
        
        if obj is not None:
            inventory_item_obj = get_inventory_item_by_id(inventory_item_id)
            if inventory_item_obj is not None:
                use_item_on_object(inventory_item_obj, obj)
        else:
            if fallback_x is not None and fallback_z is not None:
                self.walker.set_path(fallback_x, fallback_z)
            
        return 2000
    def _use_item_on_wall_object(self, inventory_item_id, wall_x, wall_z, fallback_x=None, fallback_z=None):
        """
        Uses an inventory item on a wall object at the given coordinates.

        Parameters:
            inventory_item_id (int): The ID of the inventory item to use.
            wall_x (int): The X coordinate of the wall object.
            wall_z (int): The Z coordinate of the wall object.

        Returns:
            int: Delay in milliseconds after the action.
        """
        log("use item: " + str(inventory_item_id) + " on wall object at: (" + str(wall_x) + ", " + str(wall_z) + ")")
        wall_obj = get_wall_object_from_coords(wall_x, wall_z)
        if wall_obj is not None:
            inventory_item_obj = get_inventory_item_by_id(inventory_item_id)
            use_item_on_wall_object(inventory_item_obj, wall_obj)
        else:
            if fallback_x is not None and fallback_z is not None:
                self.walker.set_path(fallback_x, fallback_z)
            
        return 1500

    def _use_item_on_ground_item(self, inventory_item_id, ground_item_id):
        """
        Uses an inventory item on a ground item at the given coordinates.

        Parameters:
            inventory_item_id (int): The ID of the inventory item to use.
            ground_item_id (int): The ground item id

        Returns:
            int: Delay in milliseconds after the action.
        """
        log("use item: " + str(inventory_item_id) + " on ground item "+ str(ground_item_id))
        ground_item = get_nearest_ground_item_by_id(ground_item_id)
        if ground_item is not None:
            inventory_item_obj = get_inventory_item_by_id(inventory_item_id)
            use_item_on_ground_item(inventory_item_obj, ground_item)
        else:
            log("tried to use item: " + str(inventory_item_id ) + " on ground item: " + str(ground_item_id) + ". But did not succeed")
            
        return 1500
        
    def _use_item_on_ground_object(self, inventory_item_id, ground_item_x, ground_item_z):
        """
        Uses an inventory item on a ground object at the given coordinates.

        Parameters:
            inventory_item_id (int): The ID of the inventory item to use.
            ground_item_x (int): The X coordinate of the ground object.
            ground_item_z (int): The Z coordinate of the ground object.

        Returns:
            int: Delay in milliseconds after the action.
        """
        log("use item: " + str(inventory_item_id) + " on ground object at: (" + str(ground_item_x) + ", " + str(ground_item_z) + ")")
        ground_obj = get_object_from_coords(ground_item_x,ground_item_z)
        if ground_obj is not None:
            inventory_item_obj = get_inventory_item_by_id(inventory_item_id)
            use_item_on_object(inventory_item_obj, ground_obj)
        else:
            log("tried to use item: " + str(inventory_item_id ) + " on ground object at coords: " + str(ground_item_x) + "," + str(ground_item_z) + ". But did not succeed")
            
        return 1500
        
    def _is_object_closed(self, obj):
        o = get_object_from_coords(obj.x, obj.z)
        if o != None and o.id == obj.id:
            return True

        return False
        
    def _is_within_reach(self, obj):
        return distance_to(obj.x, obj.z) <= 1
    
    def get_lumbridge_bucket_of_milk(self):
        log("getting lumbridge bucket of milk")
        if not has_inventory_item(21):
            return self.get_lumbridge_bucket()
            
        return self._use_item_on_npc(21,6,103,613)
    
    def get_lumbridge_bucket(self):
        log("getting lumbridge bucket")
        if not in_rect(122,607, 3, 5):
            self.walker.set_path(120,609)
            return 1000

        obj = get_nearest_ground_item_by_id(21)
        if obj is not None:
            pickup_item(obj)
            
        return 1500
        
    def get_catherby_bucket(self, amount = 1):
        log("getting catherby bucket")
        return buy(self.walker, 439, 504, None, None, None, None, 280, 0, 21, amount)
        
    def get_catherby_rope(self, amount = 1):
        log("getting rope")
        return buy(self.walker, 439, 504, None, None, None, None, 280, 0, 237, amount)

    def get_taverly_pestle_and_mortar(self, amount = 1):
        log("getting pestle and mortar")
        return buy(self.walker, 369, 507, None, None, None, None, 230, 0, 468, amount)

    def get_falador_snape_grass(self):
        log("getting snape grass")
        return self._pickup_item(469, 358, 615)

    def get_varrock_longbow(self, amount = 1):
        log("getting varrock longbow")
        
        return buy(self.walker, 114,512, None, None, None, None, 58, 0, 188, amount)
    
    def get_varrock_arrows(self, amount = 1):
        log("getting varrock arrows")
        
        return buy(self.walker, 114,512, None, None, None, None, 58, 0, 11, amount)
        
        
    #_pickup_item(self, item_id, fallback_x=None, fallback_z=None)
    def get_barbarian_village_pickaxe(self):
        log("getting pickaxe")
        
        return self._pickup_item(156,230,509)

    def get_port_sarim_chocolate_bar(self, amount = 1):
        log("getting chocolate bar")
        # if we have done pirates treasure the answer is different
        answer_idx = 0
        
        if is_quest_complete(8):
            answer_idx = 3
        
        return buy(self.walker, 274, 657, 276, 655, 3, 4, 129, answer_idx, 337, amount)

    def get_draynor_garlic(self):
        log("get garlic in draynor")
        # If not at the garlic location and not upstairs
        if not at(216, 619) and get_z() < 1000:
            self.walker.set_path(216, 619)
            return 1000

        # Use the ladder to go upstairs
        ladder_up = get_object_from_coords(216, 620)
        if ladder_up is not None and ladder_up.id == 5 and get_z() < 1500:
            at_object(ladder_up)
            return 2500

        # Interact with the cupboard upstairs
        if get_z() > 1500:
            cupboard = get_object_from_coords(216, 1562)
            if cupboard is not None:
                at_object(cupboard)
                return 1500

        return 1000
        
    def get_ardougne_spade(self):
        log("get spade in ardougne")
        if not in_rect(616, 577, 2, 4):
            self.walker.set_path(615, 579)
            return 1000

        # Pick up the spade from the ground
        spade = get_nearest_ground_item_by_id(211)
        if spade is not None:
            pickup_item(spade)
            return 1500

        return 1000

    def get_falador_spade(self):
        log("get spade in falador")
        """
        Fetch the spade located in Falador's garden.

        Returns:
            int: The delay time (in ms) for the next action.
        """
        if not in_rect(291, 521, 3, 3):
            self.walker.set_path(289, 522)
            return 1000

        # Pick up the spade from the ground
        spade = get_nearest_ground_item_by_id(211)
        if spade is not None:
            pickup_item(spade)
            return 1500

        return 1000

    # Add more operations as methods of this class


class QuestHandler:
    def __init__(self):
        self.talk_timeout = 0

    def talk_to_npc_with_dialogue(self, walker, npc_id, menu_options, fallback_path=None, inventory_check_item=None, next_stage=None):
        """
        Handles interaction with an NPC, including answering dialogue options.

        Parameters:
            npc_id (int): The ID of the NPC to interact with.
            menu_options (list): List of possible dialogue options.
            fallback_path (tuple): Coordinates (x, y) for the fallback path if the NPC is not reachable.
            inventory_check_item (int, optional): Item ID to check in inventory before proceeding.
            next_stage DEPRECATED
        """
        # Check for inventory item
        if inventory_check_item and has_inventory_item(inventory_check_item):
            self.talk_timeout = 0
            return 1000

        # Handle timeout
        if self.talk_timeout is not 0 and time.time() < self.talk_timeout:
            return 250

        # Handle dialogue menu
        if is_option_menu():
            for opt in menu_options:
                idx = get_option_menu_index(opt)
                if idx is not -1:
                    answer(idx)
                    self.talk_timeout = time.time() + 5
                    return 2000

        # Find and talk to the NPC
        npc = get_nearest_npc_by_id(npc_id, reachable=True)
        if npc is not None and npc.is_talking():
            return 500
        elif npc is not None:
            talk_to_npc(npc)
            return 5000
        elif npc is None and fallback_path is not None:
            walker.set_path(*fallback_path)
            return 1000
            
        return 1000
        
    def check_items_in_bank(self, walker, bank, required_items, set_substage, next_substage):
        """
        Withdraws needed items from bank.
        """
        if is_in_bank():
            if not is_bank_open():
                return open_bank()

            if is_bank_open():
                for item_id, quantity in required_items.items():
                    item_id_int = int(item_id)
                    if has_bank_item(item_id_int):
                        bank_count = get_bank_count(item_id_int)
                        inv_count = get_inventory_count_by_id(item_id_int)

                        to_withdraw = min(quantity - inv_count, bank_count)
                        if to_withdraw > 0:
                            withdraw(item_id_int, to_withdraw)
                            time.sleep(1)

                close_bank()
                set_substage(next_substage)
                return 1000

        walker.set_path(bank.x, bank.z)
        return 1000

    def check_inventory_items(self, required_items, set_substage, next_substage_if_missing, next_substage_if_complete):
        """
        Removes found items from required list and advances substage.
        """
        items_to_remove = []
        for item_id, quantity in required_items.items():
            item_id_int = int(item_id)
            if get_inventory_count_by_id(item_id_int) >= quantity:
                items_to_remove.append(item_id)

        for item_id in items_to_remove:
            del required_items[item_id]

        if required_items:
            log("Missing items after bank/inventory check: " + str(required_items))
            set_substage(next_substage_if_missing)
        else:
            set_substage(next_substage_if_complete)

        return 1000

## start shop handler
def buy(walker, walk_to_x, walk_to_z, rect_x, rect_z, rect_width, rect_height, npc_id, answer_idx, item_id, item_count):
    """
    Args:
        walker (PathWalker): PathWalker instance for navigation.
        walk_to_x, walk_to_z (int): Coordinates to walk to within the shop.
        rect_x, rect_z (int): Top-left corner of the shop's rectangle.
        rect_width, rect_height (int): Dimensions of the shop rectangle.
        npc_id (int): NPC ID of the shopkeeper.
        answer_idx (int): Dialogue option index to open the shop.
        item_id (int): Item ID to purchase.
        item_count (int): Quantity of the item to buy.

    Returns:
        int: Delay in milliseconds before the next action.
    """
    if not is_shop_open():
        if is_option_menu():
            answer(answer_idx)
            return 5000
        else:
            npc = get_nearest_npc_by_id(npc_id, reachable=True)
            if npc is not None and npc.is_talking():
                return 500
            elif npc is not None:
                talk_to_npc(npc)
                return 3000
            else:
                walker.set_path(walk_to_x, walk_to_z)
                return 3000
    elif is_shop_open():
        shop_item_id = get_shop_item_by_id(item_id)
        if shop_item_id is not None:
            buy_shop_item(item_id, item_count)
        else:
            close_shop()
            
        return 3000
            
    return 3000
## stop shop handler
    
## start banks
class Bank:
    def __init__(self, name, x, z, north_x, west_z, width, height):
        self.name = name
        self.x = x
        self.z = z
        self.north_x = north_x
        self.west_z = west_z
        self.width = width
        self.height = height
        
        
BANKS = [
    Bank("Al Kharid", 87, 695, 93, 689, 7, 12),
    Bank("Ardougne North", 580, 573, 585, 572, 8, 4),
    Bank("Ardougne South", 550, 612, 554, 609, 4, 8),
    Bank("Catherby", 440, 496, 443, 491, 7, 6),
    Bank("Draynor", 220, 635, 223, 634, 8, 5),
    Bank("Edgeville", 215, 450, 220, 448, 9, 6),
    Bank("Falador East", 285, 570, 286, 564, 7, 10),
    Bank("Falador West", 330, 555, 334, 549, 7, 9),
    Bank("Seers Village", 500, 453, 504, 447, 7, 7),
    Bank("Varrock East", 102, 511, 106, 510, 9, 6),
    Bank("Varrock West", 150, 505, 153, 498, 7, 10),
    Bank("Yanille", 587, 752, 590, 750, 6, 7)
]

def is_in_bank():
    for bank in BANKS:
        if in_rect(bank.north_x, bank.west_z, bank.width, bank.height):
            return True
    
    return False


def get_nearest_bank(from_x, from_z):
    min_dist = float('inf')
    nearest = None

    for bank in BANKS:
        path = calculate_path(from_x, from_z, bank.x, bank.z)
        if path is not None:
            length = path.length()
            if length < min_dist:
                min_dist = length
                nearest = bank
    
    if nearest is None:
        log("No Bank Found!")
        stop_account()

    return nearest

def get_bank_by_name(bank_name):
    selected_bank = next((bank for bank in BANKS if bank.name == bank_name), None)
    
    if selected_bank is None:
        log("Unable to find bank: " + bank_name)
        stop_script()
        return False
        
    return selected_bank


## end banks
    
## start skills
class Skills:
    def __init__(self):
        self._skills = {
            '0': 'Attack', '1': 'Defence', '2': 'Strength', '3': 'Hits',
            '4': 'Ranged', '5': 'Prayer', '6': 'Magic', '7': 'Cooking',
            '8': 'Woodcutting', '9': 'Fletching', '10': 'Fishing', '11': 'Firemaking',
            '12': 'Crafting', '13': 'Smithing', '14': 'Mining', '15': 'Herblaw',
            '16': 'Agility', '17': 'Thieving'
        }

    def get_skill_name(self, skill_id):
        """Return the name of the skill for the given ID."""
        return self._skills.get(skill_id, "unknown skill")

    def get_skill_id(self, skill_name):
        """Return the ID of the skill for the given name."""
        for id_, name in self._skills.items():
            if name.lower() == skill_name.lower():
                return int(id_)
        return -1
        
    def has_skill_requirement(self, skill_name, minimum_level):
        """Check if the player has the minimum required level for a given skill."""
        skill_id = self.get_skill_id(skill_name)
        
        if skill_id == -1:
            log("error: skill " + skill_name + " not found.")
            stop_account()
            return False
        
        skill_level = get_max_stat(skill_id)
        log("skill:" + skill_name + " with ID: " + str(skill_id) + " current Level: " + str(skill_level))
        
        if skill_level < minimum_level:
            log("skill:" + skill_name + " not high enough, minimum: " + str(minimum_level))
            stop_account()
            stop_script()
        
## end skills

## start walker
class Gate:
    def __init__(self, id, x, z, is_x, thresh_to, thresh_from):
        self.id = id
        self.x = x
        self.z = z
        self.is_x = is_x
        self.thresh_to = thresh_to
        self.thresh_from = thresh_from

class Door:
    def __init__(self, id, x, z):
        self.id = id
        self.x = x
        self.z = z

GATES = [
    Gate(137, 341, 487, True, 341, 342),   
    Gate(254, 434, 682, True, 434, 435),  
    Gate(138, 343, 581, False, 580, 581),
    Gate(346, 331, 142, False, 141, 142),
    Gate(346, 111, 142, False, 141, 142),
    #Gate(358, 564, 492, True, 564, 565),
    Gate(60, 535, 555, True, 535, 536),
    Gate(57, 467, 463, False, 462, 463),
]

#ID: 358 (the unique identifier of the gate object).
#X and Z: The coordinates of the gate are (x=564, z=492).
#is_x: Set to True because the player moves along the X-axis from 564 to 565.
#thresh_from and thresh_to: The gate is considered open when the player moves from X=564 to X=565. Thus, thresh_from=564 and thresh_to=565.

DOORS = [
    Door(2, 215, 618),
    Door(2, 292, 522),
    Door(2, 560, 487),
    Door(2, 215, 613),
    Door(2, 582, 603),
    Door(2, 275, 655),
    Door(2, 371, 506),
    Door(2, 616, 577),
    Door(2, 586, 586),
    Door(2, 581, 586),
    Door(2, 581, 586),
    Door(2, 578, 585),
    Door(2, 646, 603),
    Door(2, 230, 511),
    Door(2, 274, 563),
    Door(2, 252, 541),
    Door(2, 274, 563),
    Door(64, 102, 509),
    Door(64, 287, 571),
    Door(60, 105, 619)
]


class PathWalker:
    def __init__(self):
        self.target_x = None
        self.target_z = None
        self.path = None

    def set_path_through_points(self, to_path):
        self.path = calculate_path_through(points=to_path)
        self.path.set_nearest()
        
    def set_path(self, x, z):
        """Sets a new path to the given coordinates."""
        self.target_x = x
        self.target_z = z
        self.path = calculate_path_to(x, z, skip_local=False)

        if self.path == None:
            log("cannot create path, lets first try to open doors and gates")
            if self._process_doors() or self._process_gates():
                return 2000
           

    def walk_path(self):
        """Processes the current path and manages gates and doors."""
        if self.path:
            if at(self.target_x, self.target_z):
                self.path = None
                return 1000
                
            self.path.process()
            if self._process_doors() or self._process_gates():
                return 3000
            
            if not self.path.complete():
                if not self.path.walk():
                    self.path = None
                    return 100
                return 1000
            else:
                self.path = None
                return 1000

    def is_walking(self):
        """Returns True if there is an active path."""
        return self.path is not None

    def _process_doors(self):
        for door in DOORS:
            if in_radius_of(door.x, door.z, 15):
                wall_obj = get_wall_object_from_coords(door.x, door.z)                
                if wall_obj and wall_obj.id == door.id:
                    at_wall_object(wall_obj)
                    return True
                    
                door_obj = get_object_from_coords(door.x, door.z)
                if door_obj and door_obj.id == door.id:
                    at_object(door_obj)
                    return True
                    
        return False

    def _process_gates(self):
        for gate in GATES:
            if in_radius_of(gate.x, gate.z, 15):
                player_next_coord = self.path.next_z() if not gate.is_x else self.path.next_x()
                gate_coord = get_z() if not gate.is_x else get_x()
                if (player_next_coord <= gate.thresh_to and gate_coord >= gate.thresh_from) or (player_next_coord >= gate.thresh_from and gate_coord <= gate.thresh_to):
                    gate_obj = get_object_from_coords(gate.x, gate.z)
                    if gate_obj and gate_obj.id == gate.id:
                        at_object(gate_obj)
                        return True
        return False
## end walker

class Reporting:
    def to_units_per_hour(self, processed, the_start_time):
        if processed == 0:
            return "0"

        amount = processed * 3600.0
        seconds = time.time() - the_start_time
        units_per_hour = amount / seconds

        return str(units_per_hour)
        
    def time_to_completion(self, processed, remaining, the_start_time):
        if processed == 0:
            return "0:00:00"

        elapsed_seconds = time.time() - the_start_time

        if elapsed_seconds <= 0:
            return "0:00:00"

        seconds_per_item = elapsed_seconds / processed
        ttl = int(seconds_per_item * remaining)

        hours = ttl // 3600
        minutes = (ttl % 3600) // 60
        seconds = ttl % 60
        
        return str(hours) + ":" + str(minutes) + ":" + str(seconds)
        
    def add_progress_entry(self, name, processed, remaining=0, start_time=None, out_map=None, show_hourly_rate=True):
        """
        Generic method to calculate per-hour and estimated time left for any metric.

        Args:
            name (str): Name of the metric (e.g., "Bought: Bronze Bar")
            processed (int): Amount processed / gathered
            remaining (int): Remaining to reach target (0 = infinite)
            start_time (float): start time in seconds (time.time())
            out_map (dict): Optional. If provided, will add the entry here
            show_hourly_rate (bool): If False, hides the per-hour and estimated time info

        Returns:
            str: Formatted string if out_map is None
        """
        if start_time is None:
            start_time = time.time()

        elapsed = time.time() - start_time
        if elapsed <= 0:
            elapsed = 1

        per_hour = processed * 3600.0 / elapsed

        if not show_hourly_rate:
            formatted = str(processed)
        else:
            if remaining > 0:
                hours_left = remaining / per_hour if per_hour > 0 else 0
                formatted = "%d (%.0f/h, remaining: %d, est: %d:%02d:%02d)" % (
                    processed,
                    per_hour,
                    remaining,
                    int(hours_left),
                    int((hours_left * 60) % 60),
                    int((hours_left * 3600) % 60)
                )
            elif remaining == 0:
                formatted = "%d (%.0f/h, remaining: 0)" % (processed, per_hour)
            else:
                formatted = "%d (%.0f/h, done)" % (processed, per_hour)

        if out_map is not None:
            out_map[name] = formatted
            return None
        else:
            return formatted