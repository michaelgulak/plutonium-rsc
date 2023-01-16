package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	supportBankNotes = false
	skillCount       = 18
	readWriteTimeout = 10
)

var (
	tileDefs   []*tileDef
	doorDefs   []*doorDef
	objectDefs []*objectDef
	itemDefs   []*itemDef
)

type client struct {
	*RSCPacket
	conn                net.Conn
	user                string
	pass                string
	packetChan          chan *RSCPacket
	lastSentPacketTime  int64
	sendLock            *sync.Mutex
	closed              chan struct{}
	closedLock          *sync.Mutex
	sleepChan           chan []byte
	captchaChan         chan []byte
	cancelSleep         chan struct{}
	firstWakeupReceived bool
	world               *world
	lenBuf              [2]byte
	account             *account
	localWorld          *localWorld
	scriptKillSignal    chan struct{}

	script *script

	loading         bool
	serverIndex     int
	worldX          int
	worldZ          int
	regionX         int
	regionZ         int
	localX          int
	localZ          int
	planeIndex      int
	planeMultiplier int
	x               int
	z               int
	sprite          int

	currentStats              [skillCount]int
	baseStats                 [skillCount]int
	experience                [skillCount]int
	questPoints               int
	equipmentStats            [5]int
	inventory                 [30]*item
	inventoryCount            int
	players                   [500]*player
	playerCache               [500]*player
	playersServer             [4000]*player
	playerCacheCount          int
	playerCount               int
	npcs                      [500]*npc
	npcCache                  [500]*npc
	npcsServer                [5000]*npc
	npcCacheCount             int
	npcCount                  int
	objects                   [5000]*object
	objectCount               int
	wallObjects               [5000]*wallObject
	wallObjectCount           int
	groundItems               [5000]*groundItem
	groundItemCount           int
	prayers                   [50]bool
	fatigue                   int
	accurateFatigue           float64
	sleepingFatigue           int64
	sleeping                  bool
	combatStyle               int
	optionMenuVisible         bool
	optionMenuCount           int
	optionMenu                [20]string
	banking                   bool
	bankItemCount             int
	maxBankItemCount          int
	bankItems                 []*bankItem
	tradeScreen1Active        bool
	tradeScreen2Active        bool
	tradeRecipient            int
	tradeAccepted             bool
	tradeRecipientAccepted    bool
	myTradeItems              [14]*tradeItem
	recipientTradeItems       [14]*tradeItem
	recipientTradeCount       int
	myTradeCount              int
	myTradeConfirmItems       [14]*tradeItem
	recipientConfirmItems     [14]*tradeItem
	myTradeConfirmItemCount   int
	recipientConfirmItemCount int
	tradeConfirmAccepted      bool
	shopping                  bool
	shopItemCount             int
	shopItems                 [256]*shopItem
	friendList                [200]*friend
	friendListCount           int
	ignoreList                [200]*ignored
	ignoreListCount           int
	questCount                int
	quests                    [100]*quest
	encodedChatBuffer         [4997]byte
	chatBuffer                [5096]byte
	decodedChatLength         int
	encodedChatLength         int
	appearanceChange          bool
	pathX                     [8000]int
	pathZ                     [8000]int
	pathFindSource            [96][96]int

	blackhole bool
}

type npc struct {
	id                 int
	serverIndex        int
	x                  int
	z                  int
	sprite             int
	currentHP          int
	maxHP              int
	waypointsX         [10]int
	waypointsZ         [10]int
	waypointCurrent    int
	waypointNext       int
	messageTime        time.Time
	lastMessageTimeout time.Duration
}

type player struct {
	serverIndex        int
	x                  int
	z                  int
	sprite             int
	currentHP          int
	maxHP              int
	combatLevel        int
	waypointsX         [10]int
	waypointsZ         [10]int
	waypointCurrent    int
	waypointNext       int
	username           string
	messageTime        time.Time
	lastMessageTimeout time.Duration
	skillTime          time.Time
	skillingTimeout    time.Duration
	clan               string
}

type object struct {
	id  int
	x   int
	z   int
	dir int
}

type wallObject struct {
	id  int
	x   int
	z   int
	dir int
}

type groundItem struct {
	id int
	x  int
	z  int
}

type bankItem struct {
	id     int
	amount int
}

type tradeItem struct {
	id     int
	amount int
}

type item struct {
	id       int
	amount   int
	equipped bool
	slot     int
}

type shopItem struct {
	id     int
	amount int
	price  int
}

type friend struct {
	username string
	online   bool
}

type ignored struct {
	username string
	arg0     string
	arg1     string
	old      string
}

type quest struct {
	id    int
	name  string
	stage int
}

type tileDef struct {
	Colour     int
	TileValue  int
	ObjectType int
}

type doorDef struct {
	Name             string
	Description      string
	Command1         string
	Command2         string
	DoorType         int
	Unknown          int
	WallObjectHeight int
	ModelVar2        int
	ModelVar3        int
	ID               int
}

type objectDef struct {
	Name          string
	Description   string
	Command1      string
	Command2      string
	Typ           int
	Width         int
	Height        int
	GroundItemVar int
	ObjectModel   string
	ID            int
}

type itemDef struct {
	ID               int
	Name             string
	Description      string
	Command          string
	IsFemaleOnly     int
	IsMembersOnly    int
	IsStackable      int
	IsUntradable     int
	IsWearable       int
	AppearanceID     int
	WearableID       int
	WearSlot         int
	RequiredLevel    int
	RequiredSkillID  int
	ArmourBonus      int
	WeaponAimBonus   int
	WeaponPowerBonus int
	MagicBonus       int
	PrayerBonus      int
	BasePrice        int
	IsNoteable       int
}

func main() {
	rand.Seed(time.Now().UnixNano())

	var (
		user       string
		pass       string
		scriptPath string
		regFile    string
		configPath string
		acctFile   string
		autologin  bool
		closing    = make(chan struct{})
		sigs       = make(chan os.Signal, 1)
	)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		fmt.Println("[BOT] Received kill signal")
		close(closing)
	}()

	/*f, err := os.Create("cpu.pprof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	*/
	//defer profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()

	flag.StringVar(&user, "u", "", "OpenRSC username")
	flag.StringVar(&pass, "p", "", "OpenRSC password")
	flag.StringVar(&regFile, "r", "", "Register accounts file")
	flag.StringVar(&scriptPath, "s", "", "Python script file")
	flag.StringVar(&acctFile, "f", "", "Account file")
	flag.StringVar(&configPath, "c", "settings.toml", "Config file path")
	flag.BoolVar(&autologin, "l", false, "Autologin flag")
	flag.Parse()

	parseConfig(configPath)

	if regFile != "" {
		doAccountCreations(closing, regFile)
		return
	}

	if settings.OcrSettings.OcrType == "local" {
		_, err := os.Stat(settings.OcrSettings.OcrPath)
		if os.IsNotExist(err) {
			fmt.Printf("OCR not found: %s\n", err)
			os.Exit(1)
		} else if err != nil {
			fmt.Printf("OCR Error: %s\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("[BOT] Loading definitions...")
	loadDefs()
	fmt.Println("complete")

	fmt.Printf("[BOT] Loading landscape...")
	world := newWorld()
	world.loadWorld()
	fmt.Println("complete")

	var (
		accounts []*account
		err      error
	)

	if user == "" && pass == "" {
		if acctFile == "" {
			accounts, err = loadAccounts()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		} else {
			acct, err := loadAccount(acctFile)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			accounts = append(accounts, acct)
		}
	} else {
		acct := &account{
			username:    user,
			password:    pass,
			autologin:   autologin,
			enabled:     true,
			debug:       true,
			progressDur: time.Duration(math.MaxInt),
		}
		if scriptPath != "" {
			acct.script = &script{
				name: scriptPath,
			}
		}
		accounts = append(accounts, acct)
	}

	wg := &sync.WaitGroup{}
	loadScriptLock := &sync.Mutex{}

	loginWait := make(chan struct{})

	go func() {
		for {
			<-loginWait
			time.Sleep(6 * time.Second)
		}
	}()

	if len(accounts) > 1 {
		fmt.Println("[BOT] NOTE: Plutonium has to stagger logins by 6 seconds so as to not spam the server.")
	}

	for _, acc := range accounts {
		if !acc.enabled {
			continue
		}
		var err error
		acc.progressFile, err = os.Create(settings.LogSettings.Directory + string(os.PathSeparator) + settings.LogSettings.ProgressReports + string(os.PathSeparator) + acc.username + ".log")
		if err != nil {
			fmt.Printf("[%s] Error creating progress report file: %s\n", acc.username, err)
			os.Exit(1)
		}
		defer acc.progressFile.Close()
		acc.lastProgressReport = time.Now()

		wg.Add(1)
		go func(acc *account) {
		clientLoop:
			for {
				select {
				case <-closing:
					break clientLoop
				default:
				}
				c := &client{
					RSCPacket:        &RSCPacket{buf: make([]byte, 5000)},
					user:             acc.username,
					pass:             acc.password,
					world:            world,
					sendLock:         &sync.Mutex{},
					closed:           make(chan struct{}),
					closedLock:       &sync.Mutex{},
					packetChan:       make(chan *RSCPacket, 32),
					sleepChan:        make(chan []byte, 1),
					captchaChan:      make(chan []byte, 1),
					cancelSleep:      make(chan struct{}),
					loading:          true,
					account:          acc,
					localWorld:       newLocalWorld(),
					scriptKillSignal: make(chan struct{}, 1),
				}

				c.init()

				if acc.script != nil {
					c.script = acc.script
					if err := c.script.load(loadScriptLock, c); err != nil {
						fmt.Printf("[%s] Script load error [%s]: %s\n", c.user, c.script.name, err)
						break clientLoop
					}
				}

				select {
				case loginWait <- struct{}{}:
				case <-closing:
					break clientLoop
				}

				loginResponse, err := c.login()
				if err != nil {
					fmt.Printf("[%s] Login error: %s\n", c.user, err)
					time.Sleep(15 * time.Second)
					continue clientLoop
				}
				fmt.Printf("[%s] Login response: %d\n", c.user, loginResponse)
				if loginResponse&0x40 == 0 {
					switch loginResponse {
					case 4: // currently logged in
						fmt.Printf("[%s] Account already logged in\n", c.user)
						time.Sleep(time.Second * 5)
						continue clientLoop
					case 100: // something
						fmt.Printf("[%s] Weird server response\n", c.user)
						time.Sleep(time.Second * 5)
						continue clientLoop
					case 3:
						fmt.Printf("[%s] Invalid username or password\n", c.user)
						break clientLoop
					case 5:
						fmt.Printf("[%s] The client has been updated\n", c.user)
						break clientLoop
					case 6:
						fmt.Printf("[%s] You may only use 1 character at once, waiting 5-7 minutes\n", c.user)
						select {
						case <-time.After((5 * time.Minute) + time.Duration(rand.Int63n(int64(2*time.Minute)))):
							continue clientLoop
						case <-closing:
							break clientLoop
						}
					case 7:
						fmt.Printf("[%s] Login attempts exceeded, waiting 5-7 minutes\n", c.user)
						select {
						case <-time.After((5 * time.Minute) + time.Duration(rand.Int63n(int64(2*time.Minute)))):
							continue clientLoop
						case <-closing:
							break clientLoop
						}
					case 8:
						fmt.Printf("[%s] Server rejected session\n", c.user)
						time.Sleep(5 * time.Second)
						continue clientLoop
					case 9:
						fmt.Printf("[%s] Error unable to login\n", c.user)
						time.Sleep(5 * time.Second)
						continue clientLoop
					case 10:
						fmt.Printf("[%s] Username already in use\n", c.user)
						time.Sleep(10 * time.Second)
						continue clientLoop
					case 11:
						fmt.Printf("[%s] Account temporarily disabled, waiting 5-7 minutes\n", c.user)
						select {
						case <-time.After((5 * time.Minute) + time.Duration(rand.Int63n(int64(2*time.Minute)))):
							continue clientLoop
						case <-closing:
							break clientLoop
						}
					case 12:
						fmt.Printf("[%s] Account permanently disabled, waiting 5-7 minutes\n", c.user)
						select {
						case <-time.After((5 * time.Minute) + time.Duration(rand.Int63n(int64(2*time.Minute)))):
							continue clientLoop
						case <-closing:
							break clientLoop
						}
					case 14:
						fmt.Printf("[%s] World full\n", c.user)
						time.Sleep(5 * time.Second)
						continue clientLoop
					case 15, 16, 17, 18, 20, 21, 22, 23, 24, 25:
						fmt.Printf("[%s] Generic server error, waiting 5-7 minutes\n", c.user)
						select {
						case <-time.After((5 * time.Minute) + time.Duration(rand.Int63n(int64(2*time.Minute)))):
							continue clientLoop
						case <-closing:
							break clientLoop
						}
					default:
						fmt.Printf("[%s] Unhandled login response \n", c.user)
						select {
						case <-time.After(time.Minute):
							continue clientLoop
						case <-closing:
							break clientLoop
						}
					}
				}

				go func() {
				out:
					for {
						select {
						case <-c.closed:
							break out
						default:
							message, err := c.readPacket()
							if err != nil {
								if err != io.EOF {
									fmt.Printf("[%s] Read error: %s\n", c.user, err)
								}
								c.close()
								break out
							}

							c.packetChan <- message
						}
					}
					c.conn.Close()
				}()

				go func() {
					t := time.NewTicker(time.Duration(math.MaxInt))
					t.Stop()
					var buf []byte

					drainTicker := func() {
						t.Stop()
						select {
						case <-t.C:
						default:
						}
					}
				loop:
					for {
						select {
						case <-c.closed:
							drainTicker()
							break loop
						case <-c.cancelSleep:
							drainTicker()
							if c.account.debug {
								fmt.Printf("[%s] Cancelled sleep\n", c.user)
							}
						case buf = <-c.sleepChan:
							drainTicker()
							t = time.NewTicker(500 * time.Millisecond)
						case <-t.C:
							sleepFatigue := atomic.LoadInt64(&c.sleepingFatigue)
							if c.script != nil && c.script.fatigueTricking && sleepFatigue < 100 {
								drainTicker()
								c.captchaChan <- buf
							} else if sleepFatigue == 0 {
								drainTicker()
								c.captchaChan <- buf
							} else {
								if c.account.debug {
									fmt.Printf("[%s] Sleep fatigue at %d\n", c.user, sleepFatigue)
								}
							}
						}
					}
				}()

				go func() {
					var diff int64
					t := time.NewTimer(time.Second * 5)
					for {
						select {
						case <-c.closed:
							if !t.Stop() {
								<-t.C
							}
							return
						case <-t.C:
							diff = time.Now().Unix() - atomic.LoadInt64(&c.lastSentPacketTime)

							if diff >= 5 {
								c.createPacket(67)
								c.sendPacket()

								t.Reset(5 * time.Second)
							} else {
								t.Reset(time.Duration(5-diff) * time.Second)
							}
						}
					}
				}()

				var (
					progressTimer = time.NewTimer(acc.progressDur - time.Since(acc.lastProgressReport))
					scriptTimer   = time.NewTimer(500 * time.Millisecond)
					localClosing  = closing
				)

				if c.account.progressDur == time.Duration(math.MaxInt) {
					c.writeProgressReportf("You have not configured progress reports")
				} else {
					if c.script.onProgressReport != nil {
						c.writeProgressReportf("Next progress report scheduled for %s", time.Now().Add(acc.progressDur-time.Since(acc.lastProgressReport)).Format(time.RFC1123))
					} else {
						c.writeProgressReportf("Your script is not designed to output progress reports")
					}
				}

			loop:
				for {
					select {
					case <-c.closed:
						if !progressTimer.Stop() {
							<-progressTimer.C
						}
						if !scriptTimer.Stop() {
							<-scriptTimer.C
						}
						break loop
					case packet := <-c.packetChan:
						opcode := packet.readByteAsInt()
						c.handlePacket(packet, opcode, len(packet.buf)-1)
					case <-progressTimer.C:
						acc.lastProgressReport = time.Now()
						if c.script != nil && c.script.running && c.account.progressDur != time.Duration(math.MaxInt) {
							c.script.callOnProgressReport()
							c.writeProgressReportf("Next progress report scheduled for %s", time.Now().Add(acc.progressDur).Format(time.RFC1123))
						}
						progressTimer.Reset(acc.progressDur)
					case buf := <-c.captchaChan:
						c.doCaptcha(buf)
					case <-c.scriptKillSignal:
						c.close()
						fmt.Printf("[%s] Script kill signal received, terminating\n", c.user)
						break clientLoop
					case <-scriptTimer.C:
						if c.script != nil && c.script.running {
							if !c.sleeping && !c.loading {
								scriptTimer.Reset(time.Duration(c.script.loop()) * time.Millisecond)
							} else {
								scriptTimer.Reset(500 * time.Millisecond)
							}
						} else {
							scriptTimer.Reset(time.Duration(math.MaxInt))
						}
					case <-localClosing:
						if c.script != nil && c.script.running {
							if c.script.onKillSignal != nil {
								c.script.callOnKillSignal()
								localClosing = nil
								go func() {
									time.Sleep(15 * time.Second)
									c.close()
								}()
							} else {
								c.close()
								fmt.Printf("[%s] No kill signal handler, terminating\n", c.user)
								break clientLoop
							}
						} else {
							c.close()
							fmt.Printf("[%s] Script not running, terminating\n", c.user)
							break clientLoop
						}
					}
				}

				if !acc.autologin {
					break
				}

				time.Sleep(5 * time.Second)
			}
			wg.Done()
		}(acc)
	}
	wg.Wait()
	fmt.Println("[BOT] All accounts are finished, exiting")
}

func (c *client) close() {
	c.closedLock.Lock()
	defer c.closedLock.Unlock()

	select {
	case <-c.closed:
	default:
		close(c.closed)
	}
}

func loadDefs() {
	// tile defs
	f, err := os.Open(settings.DataSettings.Directory + string(os.PathSeparator) + settings.DataSettings.TileDefs)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&tileDefs)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// door defs
	f2, err := os.Open(settings.DataSettings.Directory + string(os.PathSeparator) + settings.DataSettings.DoorDefs)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f2.Close()

	err = json.NewDecoder(f2).Decode(&doorDefs)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// object defs
	f3, err := os.Open(settings.DataSettings.Directory + string(os.PathSeparator) + settings.DataSettings.ObjectDefs)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f3.Close()

	err = json.NewDecoder(f3).Decode(&objectDefs)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// item defs
	f1, err := os.Open(settings.DataSettings.Directory + string(os.PathSeparator) + settings.DataSettings.ItemDefs)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f1.Close()

	err = json.NewDecoder(f1).Decode(&itemDefs)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func doAccountCreations(closing chan struct{}, fileStr string) {
	f, err := os.Open(fileStr)
	if err != nil {
		fmt.Printf("Error opening register account file: %s\n", err)
		os.Exit(1)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	var (
		username string
		password string
		email    string
	)
	buf := make([]byte, 5000)
	p := &RSCPacket{buf: buf}
	for scanner.Scan() {
		select {
		case <-closing:
			return
		default:
		}
		time.Sleep(2 * time.Second)

		line := scanner.Text()
		if line == "" {
			break
		}
		lineSplit := strings.Split(line, ":")
		username = lineSplit[0]
		password = lineSplit[1]
		email = lineSplit[2]
	retry:
		conn, err := net.Dial("tcp4", settings.ServerSettings.Address)
		if err != nil {
			fmt.Printf("Error connecting in account creation: %s\n", err)
			return
		}

		sessionPacket, err := readBytesRaw(conn, 4)
		if err != nil {
			fmt.Printf("Error reading session packet in account creation: %s\n", err)
			return
		}
		_ = sessionPacket

		p.createPacket0(2)
		p.writeBytes([]byte(username))
		p.writeByte(10)
		p.writeBytes([]byte(encodeCredential(password)))
		p.writeByte(10)
		p.writeBytes([]byte(email))
		p.writeByte(10)
		p.finish()

		wrote := 0
		for wrote < p.i {
			conn.SetWriteDeadline(time.Now().Add(time.Second * readWriteTimeout))
			n, err := conn.Write(p.buf[wrote:p.i])
			if err != nil {
				fmt.Printf("Error writing creation packet: %s\n", err)
				conn.Close()
				return
			}
			wrote += n
		}

		p.i = 0

		registerResponse, err := readBytesRaw(conn, 1)
		if err != nil {
			conn.Close()
			fmt.Printf("Error reading register response: %s\n", err)
			continue
		}
		conn.Close()
		switch registerResponse[0] {
		case 255:
			fmt.Printf("[%s] Registration failed\n", username)
		case 0:
			fmt.Printf("[%s] Account created\n", username)
		case 2:
			fmt.Printf("[%s] Username already taken\n", username)
		case 3:
			fmt.Printf("[%s] Email already in use\n", username)
		case 4:
			fmt.Printf("[%s] Registration disabled\n", username)
			return
		case 5:
			fmt.Printf("[%s] You have registered too recently, waiting 5 minutes\n", username)
			select {
			case <-time.After(5 * time.Minute):
			case <-closing:
				return
			}
			goto retry
		case 6:
			fmt.Printf("[%s] Invalid email address\n", username)
		case 7:
			fmt.Printf("[%s] Username must be 2-12 characters long\n", username)
		case 8:
			fmt.Printf("[%s] Invalid username\n", username)
		default:
			fmt.Printf("[%s] Unknown registration response: %d\n", username, registerResponse[0])
		}
	}
}

func (c *client) writeProgressReportf(s0 string, args ...interface{}) error {
	_, err := c.account.progressFile.Write([]byte(fmt.Sprintf(s0, args...)))
	if err != nil {
		fmt.Printf("[%s] Unable to save progress report: %s\n", c.user, err)
		return err
	}
	if runtime.GOOS == "windows" {
		_, err = c.account.progressFile.Write([]byte("\r\n"))
	} else {
		_, err = c.account.progressFile.Write([]byte{'\n'})
	}
	return err
}

func (c *client) init() {
	for i := 0; i < len(c.inventory); i++ {
		c.inventory[i] = &item{}
	}
	for i := 0; i < len(c.groundItems); i++ {
		c.groundItems[i] = &groundItem{}
	}
	for i := 0; i < len(c.objects); i++ {
		c.objects[i] = &object{}
	}
	for i := 0; i < len(c.wallObjects); i++ {
		c.wallObjects[i] = &wallObject{}
	}
	for i := 0; i < len(c.ignoreList); i++ {
		c.ignoreList[i] = &ignored{}
	}
	for i := 0; i < len(c.myTradeItems); i++ {
		c.myTradeItems[i] = &tradeItem{}
	}
}

func (c *client) doCaptcha(buf []byte) {
	bmpData := convertImage(buf)
	if bmpData == nil {
		return
	}

	var sleepWord string

	switch settings.OcrSettings.OcrType {
	case "local":
		num := rand.Intn(math.MaxInt)
		randName := hex.EncodeToString([]byte{byte(num >> 24), byte(num >> 16), byte(num >> 8), byte(num)}) + ".bmp"
		path := settings.OcrSettings.Directory + string(os.PathSeparator) + randName
		f, err := os.Create(path)
		if err != nil {
			fmt.Printf("[%s] Error creating sleep bmp file: %s\n", c.user, err)
			return
		}
		_, err = io.Copy(f, bytes.NewReader(bmpData))
		f.Close()
		if err != nil {
			fmt.Printf("[%s] Error copying sleepword bytes to file: %s\n", c.user, err)
			os.Remove(path)
			return
		}
		cmd := exec.Command(settings.OcrSettings.OcrPath, path)
		bs, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("[%s] Error getting sleepword from external ocr: %s\n", c.user, err)
			os.Remove(path)
			return
		}
		sleepWord = strings.Trim(string(bs), " \r\n\t")
		os.Remove(path)
	case "server":
		body := bytes.NewReader(bmpData)
		resp, err := http.Post(settings.OcrSettings.OcrEndpoint, "application/octet-stream", body)
		if err != nil {
			fmt.Printf("[%s] Sleep server error: %s\n", c.user, err)
			return
		}

		scanner := bufio.NewScanner(resp.Body)

		if scanner.Scan() {
			sleepWord = scanner.Text()
		} else {
			fmt.Printf("[%s] Sleep server did not return a scannable line\n", c.user)
			return
		}
	default:
		fmt.Printf("[%s] Unknown ocr setting\n", c.user)
		return
	}

	c.sendSleepWord(sleepWord)
}

func readBytesRaw(conn net.Conn, length int) ([]byte, error) {
	buf := make([]byte, length)
	readLen := 0
	for readLen < length {
		conn.SetReadDeadline(time.Now().Add(time.Second * readWriteTimeout))
		n, err := conn.Read(buf[readLen:])
		if err != nil {
			return nil, err
		}
		readLen += n
	}
	return buf, nil
}

func encodeCredential(s string) string {
	s1 := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			s1[i] = c
		} else if c >= 'A' && c <= 'Z' {
			s1[i] = c
		} else if c >= '0' && c <= '9' {
			s1[i] = c
		} else {
			s1[i] = '_'
		}
	}

	return string(s1)
}

func (client *client) login() (int, error) {
	conn, err := net.Dial("tcp4", settings.ServerSettings.Address)
	if err != nil {
		return -1, fmt.Errorf("connection error: %s", err)
	}
	client.conn = conn

	sessionPacket, err := readBytesRaw(conn, 4)
	if err != nil {
		return -1, fmt.Errorf("read session id error: %s", err)
	}
	_ = sessionPacket

	client.createPacket(0)
	client.writeByte(0)
	client.writeInt(settings.ClientSettings.ClientVersion)
	client.writeBytes([]byte(encodeCredential(client.user)))
	client.writeByte(10)
	client.writeBytes([]byte(encodeCredential(client.pass)))
	client.writeByte(10)
	client.writeLong(0)

	client.writeShort(228)
	client.writeInt(1289)
	client.writeInt(793)
	client.writeInt(1188)
	client.writeShort(13)
	client.writeShort(47)
	client.writeByte(17)
	client.writeShort(5)
	client.writeShort(54)
	client.writeShort(24)
	client.writeInt(213)
	client.writeByte(1)
	client.writeShort(6)
	client.writeInt(4)
	client.writeInt(9)
	client.writeInt(14)
	client.writeShort(49)

	client.writeInt(37) // sound cache
	client.writeByte(0) // crown count
	client.writeByte(5)
	client.writeInt(48 * 4)
	client.writeBytes([]byte("63"))
	client.writeByte(10)

	client.sendPacket()

	var loginResponse byte
	loginRespBuf, err := readBytesRaw(conn, 1)
	if err != nil {
		return -1, err
	}
	loginResponse = loginRespBuf[0]

	return int(loginResponse), nil
}

func (c *client) sendPacket() error {
	defer c.sendLock.Unlock()

	c.finish()
	wrote := 0
	length := c.RSCPacket.i
	var (
		err error
		n   int
	)
	for wrote < length {
		c.conn.SetWriteDeadline(time.Now().Add(time.Second * readWriteTimeout))
		n, err = c.conn.Write(c.RSCPacket.buf[wrote:length])
		if err != nil {
			c.close()
			fmt.Printf("[%s] Error writing packet to stream: %s\n", c.user, err)
			break
		}
		wrote += n
	}
	c.RSCPacket.i = 0
	atomic.StoreInt64(&c.lastSentPacketTime, time.Now().Unix())
	return err
}

func (c *client) readPacket() (*RSCPacket, error) {
	var length int
	lengthPacketRead := 0
	for lengthPacketRead < 2 {
		c.conn.SetReadDeadline(time.Now().Add(time.Second * readWriteTimeout))
		n, err := c.conn.Read(c.lenBuf[lengthPacketRead:])
		if err != nil {
			return nil, err
		}
		lengthPacketRead += n
	}
	length = ((int(c.lenBuf[0]) << 8) | int(c.lenBuf[1])) - 2
	readLength := 0

	buf := make([]byte, length)
	for readLength < length {
		c.conn.SetReadDeadline(time.Now().Add(time.Second * readWriteTimeout))
		n, err := c.conn.Read(buf[readLength:])
		if err != nil {
			return nil, err
		}
		readLength += n
	}
	return &RSCPacket{buf: buf}, nil
}

func (c *client) createPlayer(serverIndex, x, z, sprite int) {
	var p *player
	if p = c.playersServer[serverIndex]; p == nil {
		p = &player{
			serverIndex: serverIndex,
		}
		c.playersServer[serverIndex] = p
	}

	found := false
	for i := 0; i < c.playerCacheCount; i++ {
		if c.playerCache[i].serverIndex == serverIndex {
			found = true
			break
		}
	}

	p.sprite = sprite
	p.x = x
	p.z = z

	if found {
		if x != p.waypointsX[p.waypointCurrent] || z != p.waypointsZ[p.waypointCurrent] {
			p.waypointCurrent = (p.waypointCurrent + 1) % 10
			p.waypointsX[p.waypointCurrent] = x
			p.waypointsZ[p.waypointCurrent] = z
		}
	} else {
		p.serverIndex = serverIndex
		p.waypointCurrent = 0
		p.waypointNext = 0
		p.waypointsX[p.waypointCurrent] = x
		p.waypointsZ[p.waypointCurrent] = z
	}
	c.players[c.playerCount] = p
	c.playerCount++
}

func (c *client) createNPC(serverIndex, id, x, z, sprite int) *npc {
	var n *npc
	if n = c.npcsServer[serverIndex]; n == nil {
		n = &npc{
			serverIndex: serverIndex,
		}
		c.npcsServer[serverIndex] = n
	}

	found := false
	for i := 0; i < c.npcCacheCount; i++ {
		if c.npcCache[i].serverIndex == serverIndex {
			found = true
			break
		}
	}

	n.sprite = sprite
	n.id = id
	n.x = x
	n.z = z

	if found {
		if x != n.waypointsX[n.waypointCurrent] || z != n.waypointsZ[n.waypointCurrent] {
			n.waypointCurrent = (n.waypointCurrent + 1) % 10
			n.waypointsX[n.waypointCurrent] = x
			n.waypointsZ[n.waypointCurrent] = z
		}
	} else {
		n.serverIndex = serverIndex
		n.waypointCurrent = 0
		n.waypointNext = 0
		n.waypointsX[n.waypointCurrent] = x
		n.waypointsZ[n.waypointCurrent] = z
		if c.script != nil && c.script.onNPCSpawned != nil {
			c.script.callOnNPCSpawned(n)
		}
	}
	c.npcs[c.npcCount] = n
	c.npcCount++

	if found {
		return n
	}

	return nil
}

func calculatePrice(basePrice, shopItemPrice, shopBuyPriceMod, shopPriceMultiplier, count, var5 int, var4 bool) int {
	cost := 0

	for k := 0; var5 > k; k++ {
		k0 := k
		if var4 {
			k0 = -k
		}
		var10 := shopPriceMultiplier * (shopItemPrice + (k0 - count))
		if var10 >= -100 {
			if var10 > 100 {
				var10 = 100
			}
		} else {
			var10 = -100
		}
		scaling := shopBuyPriceMod + var10
		if scaling < 10 {
			scaling = 10
		}

		cost += basePrice * scaling / 100
	}

	return cost
}

func (c *client) handlePacket(p *RSCPacket, opcode, length int) {
	switch opcode {
	case 191: // Player coords
		c.playerCacheCount = c.playerCount
		for i := 0; i < c.playerCacheCount; i++ {
			c.playerCache[i] = c.players[i]
		}
		buf := p.readBytes(length)
		ofs := 0
		x := readBits(buf, ofs, 11)
		ofs += 11
		z := readBits(buf, ofs, 13)
		ofs += 13
		sprite := readBits(buf, ofs, 4)
		ofs += 4

		c.loading = false

		count := readBits(buf, ofs, 8)
		ofs += 8

		c.playerCount = 0

		c.createPlayer(c.serverIndex, x, z, sprite)

		c.x = x
		c.z = z
		wantX := x + c.worldX
		wantZ := z + c.worldZ
		regionX := (wantX + 24) / 48
		regionZ := (wantZ + 24) / 48

		c.regionX = (regionX*48 - 48) - c.worldX
		c.regionZ = (regionZ*48 - 48) - c.worldZ
		c.localX = c.x - c.regionX
		c.localZ = c.z - c.regionZ

		c.sprite = sprite

		for i := 0; count > i; i++ {
			player := c.playerCache[i+1]

			needsUpdate := readBits(buf, ofs, 1)
			ofs++
			if needsUpdate != 0 {
				updateType := readBits(buf, ofs, 1)
				ofs++
				if updateType != 0 {
					needsNextSprite := readBits(buf, ofs, 2)
					ofs += 2
					if needsNextSprite == 3 { // removed player / in combat
						continue
					} else {
						player.sprite = readBits(buf, ofs, 2) + (needsNextSprite << 2)
						ofs += 2
					}
				} else {
					dir := readBits(buf, ofs, 3)
					waypointX := player.waypointsX[player.waypointCurrent]
					waypointZ := player.waypointsZ[player.waypointCurrent]
					if dir == 2 || dir == 1 || dir == 3 {
						waypointX++
					}
					if dir == 6 || dir == 5 || dir == 7 {
						waypointX--
					}
					if dir == 4 || dir == 3 || dir == 5 {
						waypointZ++
					}
					if dir == 0 || dir == 1 || dir == 7 {
						waypointZ--
					}
					player.waypointCurrent = (player.waypointCurrent + 1) % 10
					player.waypointsX[player.waypointCurrent] = waypointX
					player.waypointsZ[player.waypointCurrent] = waypointZ
					player.x = waypointX
					player.z = waypointZ
					player.sprite = dir
					ofs += 3
				}
			}

			c.players[c.playerCount] = player
			c.playerCount++
		}

		for ofs+24 < length*8 {
			serverIndex := readBits(buf, ofs, 11)
			ofs += 11
			offsetX := readBits(buf, ofs, 6)
			ofs += 6
			offsetZ := readBits(buf, ofs, 6)
			ofs += 6
			sprite := readBits(buf, ofs, 4)
			ofs += 4

			if offsetX > 31 {
				offsetX -= 64
			}

			if offsetZ > 31 {
				offsetZ -= 64
			}

			c.createPlayer(serverIndex, c.x+offsetX, c.z+offsetZ, sprite)
		}
	case 79: // npc coords
		buf := p.readBytes(length)
		ofs := 0
		count := readBits(buf, ofs, 8)
		ofs += 8

		c.npcCacheCount = c.npcCount

		c.npcCount = 0

		for i := 0; i < c.npcCacheCount; i++ {
			c.npcCache[i] = c.npcs[i]
		}

		var removedNPCs map[*npc]struct{}
		if c.script != nil && c.script.onNPCDespawned != nil {
			removedNPCs = make(map[*npc]struct{})
		}

		for i := 0; i < count; i++ {
			npc := c.npcCache[i]

			var7 := readBits(buf, ofs, 1)
			ofs++
			if var7 != 0 {
				var12 := readBits(buf, ofs, 1)
				ofs++
				if var12 != 0 {
					nextSpriteOffset := readBits(buf, ofs, 2)
					ofs += 2
					if nextSpriteOffset == 3 { // npc remove / incombat
						if removedNPCs != nil {
							removedNPCs[npc] = struct{}{}
						}
						continue
					}
					npc.sprite = (nextSpriteOffset << 2) + readBits(buf, ofs, 2)
					ofs += 2
				} else { // npc walk
					dir := readBits(buf, ofs, 3)
					waypointX := npc.waypointsX[npc.waypointCurrent]
					waypointZ := npc.waypointsZ[npc.waypointCurrent]
					if dir == 2 || dir == 1 || dir == 3 {
						waypointX++
					}
					if dir == 6 || dir == 5 || dir == 7 {
						waypointX--
					}
					if dir == 4 || dir == 3 || dir == 5 {
						waypointZ++
					}
					if dir == 0 || dir == 1 || dir == 7 {
						waypointZ--
					}
					npc.waypointCurrent = (npc.waypointCurrent + 1) % 10
					npc.waypointsX[npc.waypointCurrent] = waypointX
					npc.waypointsZ[npc.waypointCurrent] = waypointZ
					npc.x = waypointX
					npc.z = waypointZ
					npc.sprite = dir
					ofs += 3
				}
			}

			c.npcs[c.npcCount] = npc
			c.npcCount++
		}

		for ofs+34 < length*8 {
			serverIndex := readBits(buf, ofs, 12)
			ofs += 12
			x := readBits(buf, ofs, 6)
			ofs += 6
			z := readBits(buf, ofs, 6)
			ofs += 6
			sprite := readBits(buf, ofs, 4)
			ofs += 4

			if x > 31 {
				x -= 64
			}

			if z > 31 {
				z -= 64
			}

			id := readBits(buf, ofs, 10)
			ofs += 10

			if n := c.createNPC(serverIndex, id, c.x+x, c.z+z, sprite); n != nil {
				if removedNPCs != nil {
					delete(removedNPCs, n)
				}
			}
		}

		for k := range removedNPCs {
			c.script.callOnNPCDespawned(k)
		}
	case 25: // load area
		c.serverIndex = p.readShort()
		c.worldX = p.readShort()
		c.worldZ = p.readShort()
		c.planeIndex = p.readShort()
		c.planeMultiplier = p.readShort()
		if c.account.debug {
			fmt.Printf("[%s] Server Index: %d\n", c.user, c.serverIndex)
		}
	case 211: // remove world entities
		count := length / 4

		for i := 0; i < count; i++ {
			x1 := (c.x + int(int16(p.readShort()))) >> 3
			z1 := (c.z + int(int16(p.readShort()))) >> 3

			groundItemCount := 0
			for j := 0; j < c.groundItemCount; j++ {
				x2 := (c.groundItems[j].x >> 3) - x1
				z2 := (c.groundItems[j].z >> 3) - z1
				if x2 != 0 || z2 != 0 {
					if j != groundItemCount {
						c.groundItems[groundItemCount].id = c.groundItems[j].id
						c.groundItems[groundItemCount].x = c.groundItems[j].x
						c.groundItems[groundItemCount].z = c.groundItems[j].z
					}
					groundItemCount++
				} else {
					if c.script != nil && c.script.onGroundItemDespawned != nil {
						c.script.callOnGroundItemDespawned(c.groundItems[j])
					}
				}
			}
			c.groundItemCount = groundItemCount

			objectCount := 0
			for j := 0; j < c.objectCount; j++ {
				x2 := (c.objects[j].x >> 3) - x1
				z2 := (c.objects[j].z >> 3) - z1

				if x2 == 0 && z2 == 0 {
					c.unregisterGameObject(c.objects[j])
					if c.script != nil && c.script.onObjectDespawned != nil {
						c.script.callOnObjectDespawned(c.objects[j])
					}
				} else {
					if objectCount != j {
						c.objects[objectCount].id = c.objects[j].id
						c.objects[objectCount].x = c.objects[j].x
						c.objects[objectCount].z = c.objects[j].z
						c.objects[objectCount].dir = c.objects[j].dir
					}
					objectCount++
				}
			}
			c.objectCount = objectCount

			wallObjectCount := 0
			for j := 0; j < c.wallObjectCount; j++ {
				x2 := (c.wallObjects[j].x >> 3) - x1
				z2 := (c.wallObjects[j].z >> 3) - z1

				if x2 == 0 && z2 == 0 {
					c.unregisterWallObject(c.wallObjects[j])
					if c.script != nil && c.script.onWallObjectDespawned != nil {
						c.script.callOnWallObjectDespawned(c.wallObjects[j])
					}
				} else {
					if wallObjectCount != j {
						c.wallObjects[wallObjectCount].id = c.wallObjects[j].id
						c.wallObjects[wallObjectCount].x = c.wallObjects[j].x
						c.wallObjects[wallObjectCount].z = c.wallObjects[j].z
						c.wallObjects[wallObjectCount].dir = c.wallObjects[j].dir
					}
					wallObjectCount++
				}
			}

			c.wallObjectCount = wallObjectCount
		}

	case 48: // load scenery
		for p.hasReadableBytes() {
			id := p.readShort()
			x := c.x + int(int8(p.readByte()))
			z := c.z + int(int8(p.readByte()))
			dir := p.readByteAsInt()

			count := 0
			for i := 0; i < c.objectCount; i++ {
				if c.objects[i].x == x && c.objects[i].z == z {
					c.unregisterGameObject(c.objects[i])
					if c.script != nil && c.script.onObjectDespawned != nil {
						c.script.callOnObjectDespawned(c.objects[i])
					}
				} else {
					if count != i {
						c.objects[count].id = c.objects[i].id
						c.objects[count].x = c.objects[i].x
						c.objects[count].z = c.objects[i].z
						c.objects[count].dir = c.objects[i].dir
					}
					count++
				}
			}

			c.objectCount = count

			if id != 60000 {
				c.objects[c.objectCount].id = id
				c.objects[c.objectCount].x = x
				c.objects[c.objectCount].z = z
				c.objects[c.objectCount].dir = dir

				c.registerGameObject(c.objects[c.objectCount])
				if c.script != nil && c.script.onObjectSpawned != nil {
					c.script.callOnObjectSpawned(c.objects[c.objectCount])
				}
				c.objectCount++
			}
		}
	case 91: // load wall objects
		for p.hasReadableBytes() {
			id := p.readShort()
			x := c.x + int(int8(p.readByte()))
			z := c.z + int(int8(p.readByte()))
			dir := p.readByteAsInt()

			count := 0
			for i := 0; i < c.wallObjectCount; i++ {
				x2 := c.wallObjects[i].x - x
				z2 := c.wallObjects[i].z - z
				if x2 == 0 && z2 == 0 {
					c.unregisterWallObject(c.wallObjects[i])
					if c.script != nil && c.script.onWallObjectDespawned != nil {
						c.script.callOnWallObjectDespawned(c.wallObjects[i])
					}
				} else {
					if count != i {
						c.wallObjects[count].id = c.wallObjects[i].id
						c.wallObjects[count].x = c.wallObjects[i].x
						c.wallObjects[count].z = c.wallObjects[i].z
						c.wallObjects[count].dir = c.wallObjects[i].dir
					}
					count++
				}
			}

			c.wallObjectCount = count

			if id != 60000 {
				c.wallObjects[c.wallObjectCount].id = id
				c.wallObjects[c.wallObjectCount].x = x
				c.wallObjects[c.wallObjectCount].z = z
				c.wallObjects[c.wallObjectCount].dir = dir

				c.registerWallObject(c.wallObjects[c.wallObjectCount])
				if c.script != nil && c.script.onWallObjectSpawned != nil {
					c.script.callOnWallObjectSpawned(c.wallObjects[c.wallObjectCount])
				}
				c.wallObjectCount++
			}
		}
	case 90: // update inventory item
		slot := p.readByteAsInt()
		itemID := p.readShort()
		noted := p.readByte() == 1
		amount := 1
		if itemDefs[itemID&0x7FFF].IsStackable != 0 || noted {
			amount = p.readInt()
		}
		c.inventory[slot].id = itemID & 0x7FFF
		c.inventory[slot].equipped = itemID/0x8000 == 1
		c.inventory[slot].amount = amount
		c.inventory[slot].slot = slot
		if slot >= c.inventoryCount {
			c.inventoryCount++
		}
	case 99: // ground items
		for p.hasReadableBytes() {
			needsRemoval := p.peekByte() == 255
			if needsRemoval {
				p.readByte()
				x := (c.x + int(int8(p.readByte()))) >> 3
				z := (c.z + int(int8(p.readByte()))) >> 3
				if supportBankNotes {
					p.readByte()
				}
				count := 0
				for i := 0; i < c.groundItemCount; i++ {
					dx := (c.groundItems[i].x >> 3) - x
					dz := (c.groundItems[i].z >> 3) - z
					if dx != 0 || dz != 0 {
						if i != count {
							c.groundItems[count].id = c.groundItems[i].id
							c.groundItems[count].x = c.groundItems[i].x
							c.groundItems[count].z = c.groundItems[i].z
						}
						count++
					} else {
						if c.script != nil && c.script.onGroundItemDespawned != nil {
							c.script.callOnGroundItemDespawned(c.groundItems[i])
						}
					}
				}
				c.groundItemCount = count
			} else {
				id := p.readShort()
				x := c.x + int(int8(p.readByte()))
				z := c.z + int(int8(p.readByte()))
				if supportBankNotes {
					p.readByte()
				}
				if (id & 0x8000) != 0 {
					id &= 0x7FFF

					count := 0
					for i := 0; i < c.groundItemCount; i++ {
						if c.groundItems[i].x == x && c.groundItems[i].z == z && c.groundItems[i].id == id {
							if c.script != nil && c.script.onGroundItemDespawned != nil {
								c.script.callOnGroundItemDespawned(c.groundItems[i])
							}
						} else {
							if i != count {
								c.groundItems[count].id = c.groundItems[i].id
								c.groundItems[count].x = c.groundItems[i].x
								c.groundItems[count].z = c.groundItems[i].z
							}
							count++
						}
					}

					c.groundItemCount = count
				} else {
					c.groundItems[c.groundItemCount].id = id
					c.groundItems[c.groundItemCount].x = x
					c.groundItems[c.groundItemCount].z = z
					if c.script != nil && c.script.onGroundItemSpawned != nil {
						c.script.callOnGroundItemSpawned(c.groundItems[c.groundItemCount])
					}
					c.groundItemCount++
				}
			}
		}
	case 234: // player visual update
		playerCount := p.readShort()
	next:
		for i := 0; playerCount > i; i++ {
			serverIndex := p.readShort()
			player := c.playersServer[serverIndex]
			updateType := p.readByte()

			switch updateType {
			case 0: // bubble
				itemID := p.readShort()
				_ = itemID
				if player != nil {
					player.skillTime = time.Now()
					player.skillingTimeout = time.Duration(float64(time.Second) * 2.4)
				}
			case 1, 6, 7: // public message
				if updateType == 1 || updateType == 7 {
					crownID := p.readInt()
					if updateType == 7 {
						muted := p.readByte()
						onTutorial := p.readByte()
						_ = muted
						_ = onTutorial
					}
					_ = crownID
					message := p.readString()
					if updateType == 7 && message == "" {
						continue next
					}
					if player != nil {
						if c.account.debug {
							fmt.Printf("[%s] Chat message[%d]: (%s) %s\n", c.user, updateType, player.username, message)
						}
					}
				} else {
					message := p.readString()
					if player != nil {
						if len(message) >= 65 {
							player.lastMessageTimeout = time.Second * 4
						} else {
							player.lastMessageTimeout = time.Second * 3
						}
						player.messageTime = time.Now()
						if c.account.debug {
							fmt.Printf("[%s] Chat message[%d]: (%s) %s\n", c.user, updateType, player.username, message)
						}
					}
				}
			case 2: // damage update
				damage := p.readByteAsInt()
				currentHP := p.readByteAsInt()
				maxHP := p.readByteAsInt()
				if player != nil {
					player.currentHP = currentHP
					player.maxHP = maxHP
					if serverIndex == c.serverIndex {
						c.currentStats[3] = currentHP
						c.baseStats[3] = maxHP
					}
					if c.script != nil && c.script.onPlayerDamaged != nil {
						c.script.callOnPlayerDamaged(damage, player)
					}
				}
			case 3: // projectile @ NPC
				projectileType := p.readShort()
				receiverServerIndex := p.readShort()
				if player != nil {
					if npc := c.npcsServer[receiverServerIndex]; npc != nil {
						if c.script != nil && c.script.onNPCProjectile != nil {
							c.script.callOnNPCProjectile(projectileType, npc, player)
						}
					}
				}
			case 4: // projectile @ player
				projectileType := p.readShort()
				shooterServerIndex := p.readShort()
				_ = projectileType
				_ = shooterServerIndex
			case 5: // appearance update
				clan := ""
				username := p.readString()
				itemCount := p.readByteAsInt()
				for i := 0; i < itemCount; i++ {
					p.readShort()
				}
				p.readByte()
				p.readByte()
				p.readByte()
				p.readByte()
				combatLevel := p.readByte()
				p.readByte()
				flag := p.readByte()
				if flag == 1 {
					clan = p.readString()
				}
				p.readByte()
				p.readByte()
				p.readByte()
				p.readInt()
				if player != nil {
					player.combatLevel = int(combatLevel)
					player.username = username
					player.clan = clan
				}
			case 8:
				heal := p.readByte()
				currentHP := p.readByteAsInt()
				maxHP := p.readByteAsInt()
				if player != nil {
					player.currentHP = currentHP
					player.maxHP = maxHP
					if serverIndex == c.serverIndex {
						c.currentStats[3] = currentHP
						c.baseStats[3] = maxHP
					}
				}
				_ = heal
			case 9:
				currentHP := p.readByteAsInt()
				maxHP := p.readByteAsInt()
				if player != nil {
					player.currentHP = currentHP
					player.maxHP = maxHP
					if serverIndex == c.serverIndex {
						c.currentStats[3] = currentHP
						c.baseStats[3] = maxHP
					}
				}
			}
		}
	case 114: // fatigue
		c.fatigue = p.readShort()
		c.accurateFatigue = (float64(p.readShort()) / 750) * 100
	case 53: // Update inventory
		c.inventoryCount = p.readByteAsInt()

		for i := 0; i < c.inventoryCount; i++ {
			id := p.readShort()
			equipped := p.readByte() > 0

			noted := p.readByte()
			_ = noted

			var amount int
			if itemDefs[id].IsStackable != 0 {
				amount = p.readInt()
			} else {
				amount = 1
			}
			c.inventory[i].id = id
			c.inventory[i].amount = amount
			c.inventory[i].equipped = equipped
			c.inventory[i].slot = i
		}

	case 117: // sleep image
		if c.account.debug {
			fmt.Printf("[%s] Got sleep image\n", c.user)
		}
		if !c.sleeping {
			atomic.StoreInt64(&c.sleepingFatigue, int64(c.fatigue))
		}
		c.sleeping = true
		bs := p.readBytes(length)

		c.sleepChan <- bs
	case 104: // npc appearances
		nUpdates := p.readShort()
		for i := 0; i < nUpdates; i++ {
			sender := p.readShort()
			npc := c.npcsServer[sender]
			updateType := p.readByte()
			switch updateType {
			case 1: // npc chat
				recipient := int16(p.readShort())
				if npc != nil {
					message := p.readString()
					if c.account.debug {
						fmt.Printf("[%s] NPC Message[%d:%d]: %s\n", c.user, npc.id, recipient, message)
					}
					npc.messageTime = time.Now()
					if len(message) >= 65 {
						npc.lastMessageTimeout = time.Second * 4
					} else {
						npc.lastMessageTimeout = time.Second * 3
					}
					if recipient != -1 {
						if player := c.playersServer[recipient]; player != nil {
							if c.script != nil && c.script.onNPCMessage != nil {
								c.script.callOnNPCMessage(message, npc, player)
							}
						}
					}
				}
			case 2: // npc damage
				damage := p.readByteAsInt()
				currentHP := p.readByteAsInt()
				maxHP := p.readByteAsInt()
				if npc != nil {
					npc.currentHP = currentHP
					npc.maxHP = maxHP
					if c.script != nil && c.script.onNPCDamaged != nil {
						c.script.callOnNPCDamaged(damage, npc)
					}
				}
			case 3: // projectile @ player
				sprite := p.readShort()
				receiverServerIndex := p.readShort()
				_ = sprite
				_ = receiverServerIndex
			case 4: // projectile @ npc
				sprite := p.readShort()
				shooterServerIndex := p.readShort()
				_ = shooterServerIndex
				_ = sprite
			case 5: // npc skulled
				skull := p.readByte()
				_ = skull
			case 6: // ?
				wield := p.readByte()
				wield2 := p.readByte()
				_ = wield
				_ = wield2
			case 7: //npc bubble
				itemType := p.readShort()
				_ = itemType
			}
		}
	case 156: // stats, experience, and quest points
		for i := 0; i < skillCount; i++ {
			c.currentStats[i] = p.readByteAsInt()
		}
		for i := 0; i < skillCount; i++ {
			c.baseStats[i] = p.readByteAsInt()
		}
		for i := 0; i < skillCount; i++ {
			c.experience[i] = p.readInt() / 4
		}
		c.questPoints = p.readByteAsInt()
	case 159: // update stat and xp
		skill := p.readByteAsInt()
		c.currentStats[skill] = p.readByteAsInt()
		c.baseStats[skill] = p.readByteAsInt()
		c.experience[skill] = p.readInt() / 4
	case 153: // equipment stats
		for i := 0; i < 5; i++ {
			c.equipmentStats[i] = p.readByteAsInt()
		}
	case 206: // toggle prayers
		for i := 0; i < length; i++ {
			c.prayers[i] = p.readByte() == 1
		}
	case 33: // update individual experience
		skill := p.readByteAsInt()
		c.experience[skill] = p.readInt() / 4
	case 244: // sleeping menu fatigue
		atomic.StoreInt64(&c.sleepingFatigue, int64(p.readShort()))
	case 115: // black hole
		c.blackhole = p.readByte() != 0
	case 123: // remove inventory item
		slot := p.readByteAsInt()
		c.inventoryCount -= 1
		for i := slot; i < c.inventoryCount; i++ {
			c.inventory[i].id = c.inventory[i+1].id
			c.inventory[i].amount = c.inventory[i+1].amount
			c.inventory[i].equipped = c.inventory[i+1].equipped
			c.inventory[i].slot = i
		}
	case 84: // wake up
		if c.account.debug && !c.firstWakeupReceived {
			fmt.Printf("[%s] Woke up\n", c.user)
		}
		if c.firstWakeupReceived {
			c.cancelSleep <- struct{}{}
		} else {
			c.firstWakeupReceived = true
		}
		c.sleeping = false
	case 129: // combat style changed
		c.combatStyle = p.readByteAsInt()
		if c.account.debug {
			fmt.Printf("[%s] Combat style: %d\n", c.user, c.combatStyle)
		}
	case 245: // show option menu
		c.optionMenuCount = p.readByteAsInt()
		for i := 0; i < c.optionMenuCount; i++ {
			c.optionMenu[i] = p.readString()
		}
		c.optionMenuVisible = true
	case 252: // hide option menu
		c.optionMenuVisible = false
	case 120: // private message
		playerName := p.readString()
		p.readString()
		p.readInt()
		message := p.readHuffman()
		if c.account.debug {
			fmt.Printf("[%s] Private message [%s]: %s\n", c.user, playerName, message)
		}
		if c.script != nil && c.script.onPrivateMessage != nil {
			c.script.callOnPrivateMessage(message, playerName)
		}
	case 87: // private message sent
		playerName := p.readString()
		message := p.readHuffman()
		if c.account.debug {
			fmt.Printf("[%s] You private message [%s]: %s\n", c.user, playerName, message)
		}
	case 249: // bank update
		slot := p.readByteAsInt()
		id := p.readShort()
		amount := p.readInt()
		if amount == 0 {
			if slot+1 >= len(c.bankItems) {
				c.bankItems = c.bankItems[:slot]
			} else {
				c.bankItems = append(c.bankItems[:slot], c.bankItems[slot+1:]...)
			}
			c.bankItemCount--
			return
		}

		if len(c.bankItems) <= slot {
			c.bankItems = append(c.bankItems, &bankItem{
				id:     id,
				amount: amount,
			})
			c.bankItemCount++
		} else {
			c.bankItems[slot].id = id
			c.bankItems[slot].amount = amount
		}
	case 42: // open bank
		c.banking = true
		c.bankItemCount = p.readShort()
		c.maxBankItemCount = p.readShort()
		c.bankItems = c.bankItems[:0]
		for i := 0; i < c.bankItemCount; i++ {
			c.bankItems = append(c.bankItems, &bankItem{
				id:     p.readShort(),
				amount: p.readInt(),
			})
		}
	case 203: // close bank
		c.banking = false
	case 92: // trade init
		c.tradeRecipient = p.readShort()
		c.tradeScreen1Active = true
		c.tradeScreen2Active = false
		c.recipientTradeCount = 0
		c.myTradeCount = 0
		c.tradeAccepted = false
		c.tradeRecipientAccepted = false
		c.tradeConfirmAccepted = false
		c.recipientConfirmItemCount = 0
		c.myTradeConfirmItemCount = 0
	case 15: // trade accepted
		c.tradeAccepted = p.readByte() > 0
	case 162: // trade recipient accepted
		c.tradeRecipientAccepted = p.readByte() > 0
	case 20: // trade confirm screen
		c.tradeScreen1Active = false
		c.tradeScreen2Active = true
		c.tradeConfirmAccepted = false
		p.readString() // recipient name
		c.recipientConfirmItemCount = p.readByteAsInt()
		for i := 0; i < c.recipientConfirmItemCount; i++ {
			c.recipientConfirmItems[i] = &tradeItem{
				id:     p.readShort(),
				amount: p.readInt(),
			}
		}
		c.myTradeConfirmItemCount = p.readByteAsInt()
		for i := 0; i < c.myTradeConfirmItemCount; i++ {
			c.myTradeConfirmItems[i] = &tradeItem{
				id:     p.readShort(),
				amount: p.readInt(),
			}
		}
	case 128: // close trade
		c.tradeScreen1Active = false
		c.tradeScreen2Active = false
		c.recipientTradeCount = 0
		c.myTradeCount = 0
		c.tradeAccepted = false
		c.tradeRecipientAccepted = false
		c.tradeConfirmAccepted = false
		c.recipientConfirmItemCount = 0
		c.myTradeConfirmItemCount = 0
	case 97: // update trade offers
		c.recipientTradeCount = p.readByteAsInt()
		for i := 0; i < c.recipientTradeCount; i++ {
			id := p.readShort()
			if c.recipientTradeItems[i] != nil && id == c.recipientTradeItems[i].id {
				c.recipientTradeItems[i].amount = p.readInt()
			} else {
				c.recipientTradeItems[i] = &tradeItem{
					id:     id,
					amount: p.readInt(),
				}
			}
		}
		c.myTradeCount = p.readByteAsInt()
		for i := 0; i < c.myTradeCount; i++ {
			id := p.readShort()
			if c.myTradeItems[i] != nil && id == c.myTradeItems[i].id {
				c.myTradeItems[i].amount = p.readInt()
			} else {
				c.myTradeItems[i] = &tradeItem{
					id:     id,
					amount: p.readInt(),
				}
			}
		}
	case 101: // shop open
		c.shopItemCount = p.readByteAsInt()
		p.readByte()           // shop type
		p.readByte()           // sell mod
		buyMod := p.readByte() // buy mod
		pmult := p.readByte()  // price multiplier

		for i := 0; i < c.shopItemCount; i++ {
			id := p.readShort()
			basePrice := itemDefs[id].BasePrice
			amount := p.readShort()
			shopPrice := p.readShort()
			if c.shopItems[i] != nil && c.shopItems[i].id == id {
				c.shopItems[i].amount = amount
				c.shopItems[i].price = calculatePrice(basePrice, shopPrice, int(buyMod), int(pmult), amount, 1, true)
			} else {
				c.shopItems[i] = &shopItem{
					id:     id,
					amount: amount,
					price:  calculatePrice(basePrice, shopPrice, int(buyMod), int(pmult), amount, 1, true),
				}
			}
		}
		c.shopping = true
	case 137: // shop close
		c.shopping = false
	case 131: // Server message
		p.readInt()             // icon sprite
		msgType := p.readByte() // message type
		infoContained := p.readByte()
		msg := p.readString()
		var senderName string
		if (infoContained & 1) != 0 {
			senderName = p.readString() // sender name
			p.readString()              // redundant sender name
		}
		if (infoContained & 2) != 0 {
			p.readString()
		}
		switch msgType {
		case 0:
			if c.account.debug {
				fmt.Printf("[%s] Server message: %s\n", c.user, msg)
			}
			if c.script != nil && c.script.onServerMessage != nil {
				c.script.callOnServerMessage(msg)
			}
		case 1:
			if c.account.debug {
				fmt.Printf("[%s] Private message from [%s]: %s\n", c.user, senderName, msg)
			}
			if c.script != nil && c.script.onPrivateMessage != nil {
				c.script.callOnPrivateMessage(msg, senderName)
			}
		case 3:
			if c.account.debug {
				fmt.Printf("[%s] Quest message: %s\n", c.user, msg)
			}
			if c.script != nil && c.script.onServerMessage != nil {
				c.script.callOnServerMessage(msg)
			}
		case 4:
			if c.account.debug {
				fmt.Printf("[%s] Chat message from [%s]: %s\n", c.user, senderName, msg)
			}
			if c.script != nil && c.script.onChatMessage != nil {
				c.script.callOnChatMessage(msg, senderName)
			}
		case 5:
			if c.account.debug {
				fmt.Printf("[%s] Friend status: %s\n", c.user, msg)
			}
		case 6:
			if c.account.debug {
				fmt.Printf("[%s] [%s] wishes to trade with you\n", c.user, senderName)
			}
			if c.script != nil && c.script.onTradeRequest != nil {
				c.script.callOnTradeRequest(senderName)
			}
		case 7:
			if c.account.debug {
				fmt.Printf("[%s] Inventory: %s\n", c.user, msg)
			}
			if c.script != nil && c.script.onServerMessage != nil {
				c.script.callOnServerMessage(msg)
			}
		case 8:
			if c.account.debug {
				fmt.Printf("[%s] Global message from [%s]: %s\n", c.user, senderName, msg)
			}
		}
	case 149: // friend update
		name := p.readString()
		formerName := p.readString()
		onlineStatus := p.readByteAsInt()
		rename := (1 & onlineStatus) != 0
		online := (4 & onlineStatus) != 0
		if online {
			p.readString() // world
		}
		for i := 0; i < c.friendListCount; i++ {
			if c.friendList[i].username == name || c.friendList[i].username == formerName {
				if rename && c.friendList[i].username == formerName {
					c.friendList[i].username = name
				}
				c.friendList[i].online = online
				return
			}
		}
		c.friendList[c.friendListCount] = &friend{
			username: name,
			online:   online,
		}
		c.friendListCount++
	case 109: // ignore list
		c.ignoreListCount = p.readByteAsInt()
		for i := 0; i < c.ignoreListCount; i++ {
			c.ignoreList[i].arg0 = p.readString()
			c.ignoreList[i].username = p.readString()
			c.ignoreList[i].arg1 = p.readString()
			c.ignoreList[i].old = p.readString()
		}
	case 183: // Can't logout response packet
		if c.account.debug {
			fmt.Printf("[%s] Got a CANT_LOGOUT server response\n", c.user)
		}
	case 5: // quest stage info
		updateQuestType := p.readByte()
		if updateQuestType == 0 {
			c.questCount = p.readByteAsInt()
			for i := 0; i < c.questCount; i++ {
				questID := p.readInt()
				questStage := p.readInt()
				questName := p.readString()

				c.quests[questID] = &quest{
					id:    questID,
					name:  questName,
					stage: int(int32(questStage)),
				}
			}
		} else if updateQuestType == 1 {
			c.quests[p.readInt()].stage = int(int32(p.readInt()))
		}
	case 237: // update ignore name change
		arg0 := p.readString()
		replace := p.readString()
		if len(replace) == 0 {
			replace = arg0
		}
		arg1 := p.readString()
		find := p.readString()
		if len(find) == 0 {
			find = arg0
		}
		rename := p.readByte() == 1
		for i := 0; i < c.ignoreListCount; i++ {
			if rename {
				if c.ignoreList[i].username == find {
					c.ignoreList[i].arg0 = arg0
					c.ignoreList[i].username = replace
					c.ignoreList[i].arg1 = arg1
					c.ignoreList[i].old = find
					return
				}
			} else if c.ignoreList[i].username == replace {
				return
			}
		}

		if rename {
			return
		}

		c.ignoreList[c.ignoreListCount].arg0 = arg0
		c.ignoreList[c.ignoreListCount].username = replace
		c.ignoreList[c.ignoreListCount].arg1 = arg1
		c.ignoreList[c.ignoreListCount].old = find
		c.ignoreListCount++
	case 83: // death
		if c.account.debug {
			fmt.Printf("[%s] Died\n", c.user)
		}
		if c.script != nil && c.script.onDeath != nil {
			c.script.callOnDeath()
		}
	case 59: // show appearance change
		c.appearanceChange = true
	case 194: // sleep word incorrect
		if c.account.debug {
			fmt.Printf("[%s] Sleep word incorrect\n", c.user)
		}
		c.cancelSleep <- struct{}{}
	case 36: // telegrab bubble
	case 222: // welcome box
	case 19: // server configs
	case 4: // close connection
	case 51: // privacy settings
	case 240: // game settings
	case 111: // completed tutorial
	case 182: // show welcome
	case 147: // send kills
	case 113: // send ironman
	case 204: // play sound
	case 250: // unlocked appearances?
	default:
		if c.account.debug {
			fmt.Printf("[%s] Unhandled packet [OPCODE=%d, LENGTH=%d]\n", c.user, opcode, length)
		}
	}
}
