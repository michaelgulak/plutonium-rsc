# Generic walker that handles members gates by Space

# Script block should look like this:

# [script]
# name = "walker.py"
#
# [script.settings]
# target_x = 329
# target_z = 553

class Gate:
    def __init__(self, id, x, z, is_x, thresh_to, thresh_from):
        self.id = id
        self.x = x
        self.z = z
        self.is_x = is_x
        self.thresh_to = thresh_to
        self.thresh_from = thresh_from

GATES = [
    Gate(137, 341, 487, True, 341, 342),
    Gate(254, 434, 682, True, 434, 435),
    Gate(138, 343, 581, False, 580, 581),
    Gate(346, 331, 142, False, 141, 142),
    Gate(346, 111, 142, False, 141, 142),
]

path = None

def loop():
    global path

    if path != None:
        path.process()
        for gate in GATES:
            if in_radius_of(gate.x, gate.z, 15):
                if gate.is_x:
                    pnc = path.next_x()
                    mnc = get_x()
                else:
                    pnc = path.next_z()
                    mnc = get_z()
                    
                if (pnc <= gate.thresh_to and mnc >= gate.thresh_from) or \
                    (pnc >= gate.thresh_from and mnc <= gate.thresh_to):
                
                    gate_ = get_object_from_coords(gate.x, gate.z)
                    if gate_ != None and gate_.id == gate.id:
                        at_object(gate_)

                    return 800
        
        if not path.complete():
            if not path.walk():
                path = calculate_path_to(settings.target_x, settings.target_z)
                if path == None:
                    log("Failed to repath")
                    stop_account()
                    return 1000

                return 100
            return 800
        else:
            path = None

    if not at(settings.target_x, settings.target_z):
        path = calculate_path_to(settings.target_x, settings.target_z, skip_local=True)
        if path == None:
            log("Failed to path")
            stop_account()
            return 1000
    else:
        path = None
        log("Walking complete!")
        stop_account()
        return 1000

    return 1000
