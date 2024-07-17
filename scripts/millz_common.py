import time

class Food:
    def __init__(self, name, id, heal):
        self.name = name
        self.id = id
        self.heal = heal

    def get_food_ids(foods):
        return [food.id for food in foods]
       
FOODS = [
    Food("Shark", 546, 20),
    Food("Swordfish", 370, 14),
    Food("Lobster", 373, 12),
    Food("Tuna", 367, 10),
    Food("Cake (1 Slice)", 335, 4),
    Food("Cake (2 Slice)", 333, 4),
    Food("Cake (Full)", 330, 4)
]

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


def get_nearest_bank(from_x, from_z):
    
    min_dist = float('inf')
    nearest = None
    
    # log("Finding closest bank.")
    
    for bank in BANKS:
        path = calculate_path(from_x, from_z, bank.x, bank.z)
        if path is not None:
            length = path.length()
            # log("Bank: " + bank.name + " - Distance: " + str(length))
            if length < min_dist:
                min_dist = length
                nearest = bank
    
    if nearest is None:
        log("No Bank Found!")
        stop_account()
        
    #log("Nearest bank: " + nearest.name)
    return nearest


def get_bank_by_name(bank_name):
    selected_bank = next((bank for bank in BANKS if bank.name == bank_name), None)
    
    if selected_bank is None:
        log("Unable to find bank: " + bank_name)
        stop_script()
        return False
        
    #log("Selected bank: " + selected_bank.name)
    return selected_bank


def is_in_bank():
    for bank in BANKS:
        if in_rect(bank.north_x, bank.west_z, bank.width, bank.height):
            return True
    
    return False

def get_current_bank():
    for bank in BANKS:
        if in_rect(bank.north_x, bank.west_z, bank.width, bank.height):
            return bank
    
    return None


def food_count():
    return get_inventory_count_by_id(ids=Food.get_food_ids(FOODS))


def withdraw_food(count, stop_when_no_food):
    if not is_bank_open():
        return False
    
    holding_amount = food_count()
    available_space = 29 - get_total_inventory_count() - holding_amount
    count_or_available = min(available_space, count)
    required_amount = count_or_available - holding_amount

    if required_amount <= 0:
        return False
        
    for FOOD in Food.get_food_ids(FOODS):
        if get_bank_count(FOOD) >= required_amount:
            log("Withdrawing " + str(required_amount) + " " + get_item_name(FOOD) + ". Remaining in bank: " + str(get_bank_count(FOOD)))
            withdraw(FOOD, required_amount)
            return True
            
    if stop_when_no_food:
        log("No food, stopping.")
        stop_account()
        
    return False
        
    
def use_food():
    if is_bank_open():
        return False
        
    currentHp = get_current_stat(3)
    maxHp = get_max_stat(3)
    missingHp = maxHp - currentHp
    
    food = get_inventory_item_by_id(ids=Food.get_food_ids(FOODS))
    if food == None:
        return False
        
    current_food_item = next((food_item for food_item in FOODS if food_item.id == food.id), None)
    
    if missingHp > current_food_item.heal or get_hp_percent() <= 40:
        if in_combat():
            walk_to(get_x(), get_z())
            return True
        
        use_item(food)
        log("Eating: " + str(currentHp) + "/" + str(maxHp) + "hp - " + current_food_item.name + " Remaining: " + str(food_count()))
        return True
    
    return False
    

def xp_per_hour(gained_xp, start_time_seconds):
    if gained_xp == 0:
        return 0
    elapsed_time = time.time() - start_time_seconds
    gained_per_second = gained_xp / elapsed_time
    return int(gained_per_second * 3600)
    
    
experience_table = {
    "1": 0,
    "2": 83,
    "3": 174,
    "4": 276,
    "5": 388,
    "6": 512,
    "7": 650,
    "8": 801,
    "9": 969,
    "10": 1154,
    "11": 1358,
    "12": 1584,
    "13": 1833,
    "14": 2107,
    "15": 2411,
    "16": 2746,
    "17": 3115,
    "18": 3523,
    "19": 3973,
    "20": 4470,
    "21": 5018,
    "22": 5624,
    "23": 6291,
    "24": 7028,
    "25": 7842,
    "26": 8740,
    "27": 9730,
    "28": 10824,
    "29": 12031,
    "30": 13363,
    "31": 14833,
    "32": 16456,
    "33": 18247,
    "34": 20224,
    "35": 22406,
    "36": 24815,
    "37": 27473,
    "38": 30408,
    "39": 33648,
    "40": 37224,
    "41": 41171,
    "42": 45529,
    "43": 50339,
    "44": 55649,
    "45": 61512,
    "46": 67983,
    "47": 75127,
    "48": 83014,
    "49": 91721,
    "50": 101333,
    "51": 111945,
    "52": 123660,
    "53": 136594,
    "54": 150872,
    "55": 166636,
    "56": 184040,
    "57": 203254,
    "58": 224466,
    "59": 247886,
    "60": 273742,
    "61": 302288,
    "62": 333804,
    "63": 368599,
    "64": 407015,
    "65": 449428,
    "66": 496254,
    "67": 547953,
    "68": 605032,
    "69": 668051,
    "70": 737627,
    "71": 814445,
    "72": 899257,
    "73": 992895,
    "74": 1096278,
    "75": 1210421,
    "76": 1336443,
    "77": 1475581,
    "78": 1629200,
    "79": 1798808,
    "80": 1986068,
    "81": 2192818,
    "82": 2421087,
    "83": 2673114,
    "84": 2951373,
    "85": 3258594,
    "86": 3597792,
    "87": 3972294,
    "88": 4385776,
    "89": 4842295,
    "90": 5346332,
    "91": 5902831,
    "92": 6517253,
    "93": 7195629,
    "94": 7944614,
    "95": 8771558,
    "96": 9684577,
    "97": 10692629,
    "98": 11805606,
    "99": 13034431,
}

def exp_until_next_level(exp):
    for lvl in sorted(experience_table, key=int):
        exp_required = experience_table[lvl]
        if exp_required > exp:
            return exp_required - exp
    return 0


polygon_fishing_guild = [583, 523, 595, 523, 605, 513, 605, 501, 600, 496, 593, 491, 583, 490]

def in_fishing_guild(x, z):
    return point_in_polygon(x, z, polygon_fishing_guild)


def get_walkable_coordinate():
    '''
    Returns a coordinate next to us that's walkable for the 5 minute timer movement
    '''
    for i in range(-1, 2):
        x = get_x() + i

        for j in range(-1, 2):
            z = get_z() + j

            if i == 0 and j == 0:
                continue

            if is_reachable(x, z):
                return [x, z]
    return [get_x(), get_z()]

