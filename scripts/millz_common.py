import time

class Food:
    def __init__(self, name, id, heal):
        self.name = name
        self.id = id
        self.heal = heal

    def get_food_ids(foods):
        return [food.id for food in foods]
       
FOODS = [
    Food("Shark", 564, 20),
    Food("Swordfish", 370, 14),
    Food("Lobster", 373, 12),
    Food("Tuna", 367, 10),
    Food("Cake (1 Slice)", 335, 4),
    Food("Cake (2 Slice)", 333, 4),
    Food("Cake (Full)", 330, 4)
]

class Bank:
    def __init__(self, name, x, y):
        self.name = name
        self.x = x
        self.y = y
        
        
BANKS = [
    Bank("Al Kharid", 87, 695),
    Bank("Ardougne North", 580, 573),
    Bank("Ardougne South", 550, 612),
    Bank("Catherby", 440, 496),
    Bank("Draynor", 220, 635),
    Bank("Edgeville", 215, 450),
    Bank("Falador East", 285, 570),
    Bank("Falador West", 330, 555),
    Bank("Seers Village", 500, 453),
    Bank("Varrock East", 102, 511),
    Bank("Varrock West", 150, 505),
    Bank("Yanille", 587, 752)
]


def get_nearest_bank(from_x, from_y):
    
    min_dist = float('inf')
    nearest = None
    
    log("Finding closest bank.")
    
    for bank in BANKS:
        path = calculate_path(from_x, from_y, bank.x, bank.y)
        if path is not None:
            length = path.length()
            # log("Bank: " + bank.name + " - Distance: " + str(length))
            if length < min_dist:
                min_dist = length
                nearest = bank
    
    if nearest is None:
        log("No Bank Found!")
        stop_script()
        
    log("Nearest bank: " + nearest.name)
    return nearest


def get_bank_by_name(bank_name):
    selected_bank = next((bank for bank in BANKS if bank.name == bank_name), None)
    
    if selected_bank is None:
        log("Unable to find bank: " + bank_name)
        stop_script()
        return False
        
    log("Selected bank: " + selected_bank.name)
    return selected_bank


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
        stop_script()
        
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
        return "0"
    elapsed_time = time.time() - start_time_seconds
    gained_per_second = gained_xp / elapsed_time
    return int(gained_per_second * 3600)

