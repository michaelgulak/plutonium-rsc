package main

import (
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"

	"github.com/go-python/gpython/py"
	_ "github.com/go-python/gpython/stdlib"
)

type script struct {
	path     string
	pathDir  string
	running  bool
	c        *client
	settings py.StringDict
	wakeupAt int
	loaded   bool
	module   *py.Module
	tick     int

	loopFn                *py.Function
	onProgressReport      *py.Function
	onServerMessage       *py.Function
	onTradeRequest        *py.Function
	onChatMessage         *py.Function
	onPrivateMessage      *py.Function
	onPlayerDamaged       *py.Function
	onNPCDamaged          *py.Function
	onNPCSpawned          *py.Function
	onNPCDespawned        *py.Function
	onNPCProjectile       *py.Function
	onNPCMessage          *py.Function
	onDeath               *py.Function
	onGroundItemSpawned   *py.Function
	onGroundItemDespawned *py.Function
	onObjectSpawned       *py.Function
	onObjectDespawned     *py.Function
	onWallObjectSpawned   *py.Function
	onWallObjectDespawned *py.Function
	onKillSignal          *py.Function
	onLoad                *py.Function
	onSystemUpdate        *py.Function
	onFatigueUpdate       *py.Function
	onSleepWordSent       *py.Function
	onServerTick          *py.Function

	bankOptionMenuTimeout time.Time
}

var (
	npcType        = py.NewType("npc", "")
	playerType     = py.NewType("player", "")
	objectType     = py.NewType("game_object", "")
	wallObjectType = py.NewType("wall_object", "")
	groundItemType = py.NewType("ground_item", "")
	bankItemType   = py.NewType("bank_item", "")
	tradeItemType  = py.NewType("trade_item", "")
	itemType       = py.NewType("inventory_item", "")
	shopItemType   = py.NewType("shop_item", "")
	friendType     = py.NewType("friend", "")
	ignoredType    = py.NewType("ignored", "")
	questType      = py.NewType("quest", "")
	longPathType   = py.NewType("path", "")
)

func init() {
	// npc
	npcType.Dict["id"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*npc).id), nil }}
	npcType.Dict["x"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*npc).x), nil }}
	npcType.Dict["z"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*npc).z), nil }}
	npcType.Dict["sid"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*npc).serverIndex), nil }}
	npcType.Dict["sprite"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*npc).sprite), nil }}
	npcType.Dict["current_hp"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*npc).currentHP), nil }}
	npcType.Dict["max_hp"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*npc).maxHP), nil }}
	npcType.Dict["is_talking"] = py.MustNewMethod("is_talking", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(self.(*npc).isTalking()), nil
	}, py.METH_CLASS, "")
	npcType.Dict["in_combat"] = py.MustNewMethod("in_combat", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(self.(*npc).inCombat()), nil
	}, py.METH_CLASS, "")
	// player
	playerType.Dict["x"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*player).x), nil }}
	playerType.Dict["z"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*player).z), nil }}
	playerType.Dict["pid"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*player).serverIndex), nil }}
	playerType.Dict["sprite"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*player).sprite), nil }}
	playerType.Dict["current_hp"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*player).currentHP), nil }}
	playerType.Dict["max_hp"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*player).maxHP), nil }}
	playerType.Dict["username"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.String(self.(*player).username), nil }}
	playerType.Dict["combat_level"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*player).combatLevel), nil }}
	playerType.Dict["is_talking"] = py.MustNewMethod("is_talking", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(self.(*player).isTalking()), nil
	}, py.METH_CLASS, "")
	playerType.Dict["in_combat"] = py.MustNewMethod("in_combat", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(self.(*player).inCombat()), nil
	}, py.METH_CLASS, "")

	// object
	objectType.Dict["id"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*object).id), nil }}
	objectType.Dict["x"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*object).x), nil }}
	objectType.Dict["z"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*object).z), nil }}
	objectType.Dict["dir"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*object).dir), nil }}
	// wall object
	wallObjectType.Dict["id"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*wallObject).id), nil }}
	wallObjectType.Dict["x"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*wallObject).x), nil }}
	wallObjectType.Dict["z"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*wallObject).z), nil }}
	wallObjectType.Dict["dir"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*wallObject).dir), nil }}
	// ground item
	groundItemType.Dict["id"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*groundItem).id), nil }}
	groundItemType.Dict["x"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*groundItem).x), nil }}
	groundItemType.Dict["z"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*groundItem).z), nil }}
	// bank item
	bankItemType.Dict["id"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*bankItem).id), nil }}
	bankItemType.Dict["amount"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*bankItem).amount), nil }}
	// trade item
	tradeItemType.Dict["id"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*tradeItem).id), nil }}
	tradeItemType.Dict["amount"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*tradeItem).amount), nil }}
	// inventory item
	itemType.Dict["id"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*item).id), nil }}
	itemType.Dict["amount"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*item).amount), nil }}
	itemType.Dict["equipped"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Bool(self.(*item).equipped), nil }}
	itemType.Dict["slot"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*item).slot), nil }}
	// shop item
	shopItemType.Dict["id"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*shopItem).id), nil }}
	shopItemType.Dict["amount"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*shopItem).amount), nil }}
	shopItemType.Dict["price"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*shopItem).price), nil }}
	// friend
	friendType.Dict["username"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.String(self.(*friend).username), nil }}
	friendType.Dict["online"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Bool(self.(*friend).online), nil }}
	// ignored
	ignoredType.Dict["username"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.String(self.(*ignored).username), nil }}
	// quest
	questType.Dict["id"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*quest).id), nil }}
	questType.Dict["name"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.String(self.(*quest).name), nil }}
	questType.Dict["stage"] = &py.Property{Fget: func(self py.Object) (py.Object, error) { return py.Int(self.(*quest).stage), nil }}
	// path
	longPathType.Dict["complete"] = py.MustNewMethod("complete", longPathComplete, py.METH_CLASS, "")
	longPathType.Dict["reset"] = py.MustNewMethod("reset", longPathReset, py.METH_CLASS, "")
	longPathType.Dict["reverse"] = py.MustNewMethod("reverse", longPathReverse, py.METH_CLASS, "")
	longPathType.Dict["next_x"] = py.MustNewMethod("next_x", longPathNextX, py.METH_CLASS, "")
	longPathType.Dict["next_z"] = py.MustNewMethod("next_z", longPathNextZ, py.METH_CLASS, "")
	longPathType.Dict["process"] = py.MustNewMethod("process", longPathProcess, py.METH_CLASS, "")
	longPathType.Dict["set_nearest"] = py.MustNewMethod("set_nearest", longPathSetNearest, py.METH_CLASS, "")
	longPathType.Dict["length"] = py.MustNewMethod("length", longPathLength, py.METH_CLASS, "")
	longPathType.Dict["walk"] = py.MustNewMethod("walk", longPathWalk, py.METH_CLASS, "")
}

func (c *client) loadScript() error {
	s := c.script
	s.c = c

	if !s.loaded {
		f, err := os.Open(s.path)
		if err != nil {
			return err
		}
		defer f.Close()
		raw, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		opts := py.DefaultContextOpts()

		opts.SysPaths = append(opts.SysPaths, s.pathDir)

		ctx := py.NewContext(opts)

		modImpl := &py.ModuleImpl{
			Info: py.ModuleInfo{
				FileDesc: filepath.Base(s.path),
			},
			CodeSrc: string(raw),
			Globals: py.NewStringDict(),
		}

		s.addPreGlobals(modImpl)

		module, err := ctx.ModuleInit(modImpl)
		if err != nil {
			return fmt.Errorf("script compile error: %s", err)
		}

		s.module = module

		s.addGlobalsToAll(module)

		if loopFn, ok := module.Globals["loop"]; !ok {
			return fmt.Errorf("loop function not found")
		} else {
			s.loopFn = loopFn.(*py.Function)
		}

		s.addHooks(s.module)

		s.loaded = true
		s.running = true
	} else {
		s.addGlobalsToAll(s.module)
	}

	return nil
}

func (s *script) loop() int {
	res, err := s.loopFn.M__call__(nil, nil)
	if err != nil {
		py.TracebackDump(err)
		os.Exit(1)
	}
	res0, ok := res.(py.Int)
	if !ok {
		fmt.Printf("[%s] Script [%s] did not return a sleep time\n", s.c.user, s.path)
		os.Exit(1)
	}
	return int(res0)
}

func (s *script) callOnProgressReport() {
	if !s.running {
		return
	}
	if s.onProgressReport != nil {
		res, err := s.onProgressReport.M__call__(nil, nil)
		if err != nil {
			py.TracebackDump(err)
			fmt.Printf("[%s] Error in on_progress_report handler: %s\n", s.c.user, err)
			os.Exit(1)
		}
		dict, ok := res.(py.StringDict)
		if !ok {
			fmt.Printf("[%s] on_progress_report should return a string dict\n", s.c.user)
			os.Exit(1)
		}
		table := tablewriter.NewWriter(s.c.account.progressFile)
		if runtime.GOOS == "windows" {
			table.SetNewLine("\r\n")
		}
		table.SetHeader([]string{"Metric", "Value"})
		type kv struct {
			key   string
			value string
		}
		rows := []*kv{}
		for k, v := range dict {
			rows = append(rows, &kv{k, fmt.Sprint(v)})
		}
		sort.Slice(rows, func(i, j int) bool {
			return strings.Compare(rows[i].key, rows[j].key) == 1
		})
		for _, row := range rows {
			table.Append([]string{row.key, row.value})
		}
		if settings.LogSettings.OverwriteProgressReports {
			s.c.account.progressFile.Truncate(0)
			s.c.account.progressFile.Seek(0, 0)
		}
		table.Render()
	}
}

func (n *npc) Type() *py.Type {
	return npcType
}

func (n *player) Type() *py.Type {
	return playerType
}

func (n *object) Type() *py.Type {
	return objectType
}

func (n *wallObject) Type() *py.Type {
	return wallObjectType
}

func (n *groundItem) Type() *py.Type {
	return groundItemType
}

func (n *bankItem) Type() *py.Type {
	return bankItemType
}

func (n *tradeItem) Type() *py.Type {
	return tradeItemType
}

func (n *item) Type() *py.Type {
	return itemType
}

func (n *shopItem) Type() *py.Type {
	return shopItemType
}

func (n *friend) Type() *py.Type {
	return friendType
}

func (n *ignored) Type() *py.Type {
	return ignoredType
}

func (n *quest) Type() *py.Type {
	return questType
}

func (n *longPath) Type() *py.Type {
	return longPathType
}

func longPathComplete(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
	n := self.(*longPath)
	return py.Bool(n.currentPt >= len(n.points)), nil
}

func longPathReset(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
	n := self.(*longPath)
	n.currentPt = 0
	return py.None, nil
}

func longPathReverse(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
	n := self.(*longPath)
	start := 0
	end := len(n.points) - 1
	for start < end {
		n.points[start], n.points[end] = n.points[end], n.points[start]
		start++
		end--
	}
	return py.None, nil
}

func longPathNextX(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
	n := self.(*longPath)
	if n.currentPt >= len(n.points) {
		return py.Int(-1), nil
	}
	return py.Int(n.points[n.currentPt].x), nil
}

func longPathNextZ(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
	n := self.(*longPath)
	if n.currentPt >= len(n.points) {
		return py.Int(-1), nil
	}
	return py.Int(n.points[n.currentPt].y), nil
}

func longPathProcess(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
	n := self.(*longPath)
	if n.currentPt >= len(n.points) {
		return py.None, nil
	}
	pt := n.points[n.currentPt]
	if pt.x == n.script.c.x && pt.y == n.script.c.z {
		n.currentPt++
	}
	return py.None, nil
}

func longPathSetNearest(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
	n := self.(*longPath)
	x, z := n.script.c.x, n.script.c.z
	minIdx := 0
	minDist := math.MaxInt
	for i, p := range n.points {
		if dist := distance(x, z, p.x, p.y); dist <= minDist {
			minDist = dist
			minIdx = i
		}
	}
	n.currentPt = minIdx
	return py.None, nil
}

func longPathLength(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
	n := self.(*longPath)
	return py.Int(n.length), nil
}

func longPathWalk(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
	n := self.(*longPath)
	if n.currentPt >= len(n.points) {
		return py.None, nil
	}
	pt := n.points[n.currentPt]
	return py.Bool(n.script.c.walkTo(pt.x, pt.y, pt.x, pt.y, false, false)), nil
}

func (s *script) addHooks(module *py.Module) {
	if onProgressReport, ok := module.Globals["on_progress_report"]; ok {
		s.onProgressReport = onProgressReport.(*py.Function)
	}
	if onServerMessage, ok := module.Globals["on_server_message"]; ok {
		s.onServerMessage = onServerMessage.(*py.Function)
	}
	if onTradeRequest, ok := module.Globals["on_trade_request"]; ok {
		s.onTradeRequest = onTradeRequest.(*py.Function)
	}
	if onChatMessage, ok := module.Globals["on_chat_message"]; ok {
		s.onChatMessage = onChatMessage.(*py.Function)
	}
	if onPrivateMessage, ok := module.Globals["on_private_message"]; ok {
		s.onPrivateMessage = onPrivateMessage.(*py.Function)
	}
	if onPlayerDamaged, ok := module.Globals["on_player_damaged"]; ok {
		s.onPlayerDamaged = onPlayerDamaged.(*py.Function)
	}
	if onNPCDamaged, ok := module.Globals["on_npc_damaged"]; ok {
		s.onNPCDamaged = onNPCDamaged.(*py.Function)
	}
	if onNPCSpawned, ok := module.Globals["on_npc_spawned"]; ok {
		s.onNPCSpawned = onNPCSpawned.(*py.Function)
	}
	if onNPCDespawned, ok := module.Globals["on_npc_despawned"]; ok {
		s.onNPCDespawned = onNPCDespawned.(*py.Function)
	}
	if onNPCProjectile, ok := module.Globals["on_npc_projectile"]; ok {
		s.onNPCProjectile = onNPCProjectile.(*py.Function)
	}
	if onNPCMessage, ok := module.Globals["on_npc_message"]; ok {
		s.onNPCMessage = onNPCMessage.(*py.Function)
	}
	if onDeath, ok := module.Globals["on_death"]; ok {
		s.onDeath = onDeath.(*py.Function)
	}
	if onGroundItemSpawned, ok := module.Globals["on_ground_item_spawned"]; ok {
		s.onGroundItemSpawned = onGroundItemSpawned.(*py.Function)
	}
	if onGroundItemDespawned, ok := module.Globals["on_ground_item_despawned"]; ok {
		s.onGroundItemDespawned = onGroundItemDespawned.(*py.Function)
	}
	if onObjectSpawned, ok := module.Globals["on_object_spawned"]; ok {
		s.onObjectSpawned = onObjectSpawned.(*py.Function)
	}
	if onObjectDespawned, ok := module.Globals["on_object_despawned"]; ok {
		s.onObjectDespawned = onObjectDespawned.(*py.Function)
	}
	if onWallObjectSpawned, ok := module.Globals["on_wall_object_spawned"]; ok {
		s.onWallObjectSpawned = onWallObjectSpawned.(*py.Function)
	}
	if onWallObjectDespawned, ok := module.Globals["on_wall_object_despawned"]; ok {
		s.onWallObjectDespawned = onWallObjectDespawned.(*py.Function)
	}
	if onKillSignal, ok := module.Globals["on_kill_signal"]; ok {
		s.onKillSignal = onKillSignal.(*py.Function)
	}
	if onLoad, ok := module.Globals["on_load"]; ok {
		s.onLoad = onLoad.(*py.Function)
	}
	if onSystemUpdate, ok := module.Globals["on_system_update"]; ok {
		s.onSystemUpdate = onSystemUpdate.(*py.Function)
	}
	if onFatigueUpdate, ok := module.Globals["on_fatigue_update"]; ok {
		s.onFatigueUpdate = onFatigueUpdate.(*py.Function)
	}
	if onSleepwordSent, ok := module.Globals["on_sleepword_sent"]; ok {
		s.onSleepWordSent = onSleepwordSent.(*py.Function)
	}
	if onServerTick, ok := module.Globals["on_server_tick"]; ok {
		s.onServerTick = onServerTick.(*py.Function)
	}
}

func (s *script) callOnServerMessage(message string) {
	if !s.running {
		return
	}
	if strings.HasSuffix(message, "moment") {
		s.bankOptionMenuTimeout = time.Unix(0, 0)
	}
	_, err := s.onServerMessage.M__call__(py.Tuple([]py.Object{py.String(message)}), nil)
	if err != nil {
		py.TracebackDump(err)
		os.Exit(1)
	}
}

func (s *script) callOnTradeRequest(name string) {
	if !s.running {
		return
	}
	_, err := s.onTradeRequest.M__call__(py.Tuple([]py.Object{py.String(name)}), nil)
	if err != nil {
		py.TracebackDump(err)
		os.Exit(1)
	}
}

func (s *script) callOnChatMessage(message, from string) {
	if !s.running {
		return
	}
	_, err := s.onChatMessage.M__call__(py.Tuple([]py.Object{py.String(message), py.String(from)}), nil)
	if err != nil {
		py.TracebackDump(err)
		os.Exit(1)
	}
}

func (s *script) callOnPrivateMessage(message, from string) {
	if !s.running {
		return
	}
	_, err := s.onPrivateMessage.M__call__(py.Tuple([]py.Object{py.String(message), py.String(from)}), nil)
	if err != nil {
		py.TracebackDump(err)
		os.Exit(1)
	}
}

func (s *script) callOnPlayerDamaged(damage int, p *player) {
	if !s.running {
		return
	}
	_, err := s.onPlayerDamaged.M__call__(py.Tuple([]py.Object{py.Int(damage), p}), nil)
	if err != nil {
		py.TracebackDump(err)
		os.Exit(1)
	}
}

func (s *script) callOnNPCDamaged(damage int, n *npc) {
	if !s.running {
		return
	}
	_, err := s.onNPCDamaged.M__call__(py.Tuple([]py.Object{py.Int(damage), n}), nil)
	if err != nil {
		py.TracebackDump(err)
		os.Exit(1)
	}
}

func (s *script) callOnNPCSpawned(n *npc) {
	if !s.running {
		return
	}
	_, err := s.onNPCSpawned.M__call__(py.Tuple([]py.Object{n}), nil)
	if err != nil {
		py.TracebackDump(err)
		os.Exit(1)
	}
}

func (s *script) callOnNPCDespawned(n *npc) {
	if !s.running {
		return
	}
	_, err := s.onNPCDespawned.M__call__(py.Tuple([]py.Object{n}), nil)
	if err != nil {
		py.TracebackDump(err)
		os.Exit(1)
	}
}

func (s *script) callOnNPCProjectile(projectileType int, n *npc, p *player) {
	if !s.running {
		return
	}
	_, err := s.onNPCProjectile.M__call__(py.Tuple([]py.Object{py.Int(projectileType), n, p}), nil)
	if err != nil {
		py.TracebackDump(err)
		os.Exit(1)
	}
}

func (s *script) callOnNPCMessage(message string, n *npc, p *player) {
	if !s.running {
		return
	}
	_, err := s.onNPCMessage.M__call__(py.Tuple([]py.Object{py.String(message), n, p}), nil)
	if err != nil {
		py.TracebackDump(err)
		os.Exit(1)
	}
}

func (s *script) callOnDeath() {
	if !s.running {
		return
	}
	_, err := s.onDeath.M__call__(nil, nil)
	if err != nil {
		py.TracebackDump(err)
		os.Exit(1)
	}
}

func (s *script) callOnGroundItemSpawned(g *groundItem) {
	if !s.running {
		return
	}
	_, err := s.onGroundItemSpawned.M__call__(py.Tuple([]py.Object{g}), nil)
	if err != nil {
		py.TracebackDump(err)
		os.Exit(1)
	}
}

func (s *script) callOnGroundItemDespawned(g *groundItem) {
	if !s.running {
		return
	}
	_, err := s.onGroundItemDespawned.M__call__(py.Tuple([]py.Object{g}), nil)
	if err != nil {
		py.TracebackDump(err)
		os.Exit(1)
	}
}

func (s *script) callOnObjectSpawned(o *object) {
	if !s.running {
		return
	}
	_, err := s.onObjectSpawned.M__call__(py.Tuple([]py.Object{o}), nil)
	if err != nil {
		py.TracebackDump(err)
		os.Exit(1)
	}
}

func (s *script) callOnObjectDespawned(o *object) {
	if !s.running {
		return
	}
	_, err := s.onObjectDespawned.M__call__(py.Tuple([]py.Object{o}), nil)
	if err != nil {
		py.TracebackDump(err)
		os.Exit(1)
	}
}

func (s *script) callOnWallObjectSpawned(o *wallObject) {
	if !s.running {
		return
	}
	_, err := s.onWallObjectSpawned.M__call__(py.Tuple([]py.Object{o}), nil)
	if err != nil {
		py.TracebackDump(err)
		os.Exit(1)
	}
}

func (s *script) callOnWallObjectDespawned(o *wallObject) {
	if !s.running {
		return
	}
	_, err := s.onWallObjectDespawned.M__call__(py.Tuple([]py.Object{o}), nil)
	if err != nil {
		py.TracebackDump(err)
		os.Exit(1)
	}
}

func (s *script) callOnKillSignal() {
	if !s.running {
		return
	}
	_, err := s.onKillSignal.M__call__(nil, nil)
	if err != nil {
		py.TracebackDump(err)
		os.Exit(1)
	}
}

func (s *script) callOnLoad() {
	if !s.running {
		return
	}
	_, err := s.onLoad.M__call__(nil, nil)
	if err != nil {
		py.TracebackDump(err)
		os.Exit(1)
	}
}

func (s *script) callOnSystemUpdate(seconds int) {
	if !s.running {
		return
	}
	_, err := s.onSystemUpdate.M__call__(py.Tuple([]py.Object{py.Int(seconds)}), nil)
	if err != nil {
		py.TracebackDump(err)
		os.Exit(1)
	}
}

func (s *script) callOnFatigueUpdate(fatigue int, accurateFatigue float64) {
	if !s.running {
		return
	}
	_, err := s.onFatigueUpdate.M__call__(py.Tuple([]py.Object{py.Int(fatigue), py.Float(accurateFatigue)}), nil)
	if err != nil {
		py.TracebackDump(err)
		os.Exit(1)
	}
}

func (s *script) callOnSleepWordSent() {
	if !s.running {
		return
	}
	_, err := s.onSleepWordSent.M__call__(nil, nil)
	if err != nil {
		py.TracebackDump(err)
		os.Exit(1)
	}
}

func (s *script) callOnServerTick() {
	if !s.running {
		return
	}
	_, err := s.onServerTick.M__call__(py.Tuple([]py.Object{py.Int(s.tick)}), nil)
	if err != nil {
		py.TracebackDump(err)
		os.Exit(1)
	}
	s.tick++
}

func (s *script) addPreGlobals(module *py.ModuleImpl) {
	module.Globals["settings"] = &py.Type{
		ObjectType: py.ObjectType,
		Name:       "settings",
		Dict:       s.settings,
	}
	module.Globals["SLEEPING_BAG"] = py.Int(1263)
	bankers := make([]py.Object, len(bankersID))
	for i := 0; i < len(bankers); i++ {
		bankers[i] = py.Int(bankersID[i])
	}
	module.Globals["BANKERS"] = py.NewListFromItems(bankers)
}

func (s *script) addGlobalsToAll(module *py.Module) {
	for _, v := range module.Globals {
		switch v := v.(type) {
		case *py.Module:
			if v.ModuleImpl.Info.FileDesc != "" {
				s.addGlobalsToAll(v)
			}
		}
	}
	s.addGlobals(module, s.c)
}

func (s *script) addGlobals(module *py.Module, c *client) {
	// readd preglobals so that submodules can access them
	module.Globals["settings"] = &py.Type{
		ObjectType: py.ObjectType,
		Name:       "settings",
		Dict:       s.settings,
	}
	module.Globals["SLEEPING_BAG"] = py.Int(1263)
	bankers := make([]py.Object, len(bankersID))
	for i := 0; i < len(bankers); i++ {
		bankers[i] = py.Int(bankersID[i])
	}
	module.Globals["BANKERS"] = py.NewListFromItems(bankers)

	// general
	module.Globals["stop_account"] = py.MustNewMethod("stop_account", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.scriptKillSignal <- struct{}{}
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["create_packet"] = py.MustNewMethod("create_packet", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.createPacket(byte(args[0].(py.Int)))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["write_byte"] = py.MustNewMethod("write_byte", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.writeByte(byte(args[0].(py.Int)))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["write_short"] = py.MustNewMethod("write_short", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.writeShort(int(args[0].(py.Int)))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["write_int"] = py.MustNewMethod("write_int", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.writeInt(int(args[0].(py.Int)))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["write_bytes"] = py.MustNewMethod("write_bytes", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.writeBytes(args[0].(py.Bytes))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["send_packet"] = py.MustNewMethod("send_packet", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.sendPacket()
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["log"] = py.MustNewMethod("log", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		fmt.Printf("[%s] %s\n", c.user, args[0].(py.String))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["debug"] = py.MustNewMethod("debug", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		if c.account.debug {
			fmt.Printf("[%s] %s\n", c.user, args[0].(py.String))
		}
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["distance"] = py.MustNewMethod("distance", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Int(distance(int(args[0].(py.Int)), int(args[1].(py.Int)), int(args[2].(py.Int)), int(args[3].(py.Int)))), nil
	}, py.METH_STATIC, "")
	module.Globals["in_rect"] = py.MustNewMethod("in_rect", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		x := int(args[0].(py.Int))
		z := int(args[1].(py.Int))
		width := int(args[2].(py.Int))
		height := int(args[3].(py.Int))
		return py.Bool(c.inRect(x, z, width, height)), nil
	}, py.METH_STATIC, "")
	module.Globals["point_in_rect"] = py.MustNewMethod("point_in_rect", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		x1 := int(args[0].(py.Int))
		z1 := int(args[1].(py.Int))
		x2 := int(args[2].(py.Int))
		z2 := int(args[3].(py.Int))
		width := int(args[4].(py.Int))
		height := int(args[5].(py.Int))
		return py.Bool(inRect(x1, z1, x2, z2, width, height)), nil
	}, py.METH_STATIC, "")
	module.Globals["point_in_polygon"] = py.MustNewMethod("point_in_polygon", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		x := args[0].(py.Int)
		z := args[1].(py.Int)
		poly := args[2].(*py.List).Items

		if len(poly) < 6 {
			return py.None, errors.New("point_in_polygon supplied poly should have at least 6 elements")
		}

		inside := false
		length := len(poly) / 2

		for i, j := 0, length-1; i < length; i, j = i+1, i {
			xi, zi := poly[i*2].(py.Int), poly[i*2+1].(py.Int)
			xj, zj := poly[j*2].(py.Int), poly[j*2+1].(py.Int)

			intersect := ((zi > z) != (zj > z)) && (x < (xj-xi)*(z-zi)/(zj-zi)+xi)

			if intersect {
				inside = !inside
			}

		}
		return py.Bool(inside), nil
	}, py.METH_STATIC, "")
	module.Globals["at"] = py.MustNewMethod("at", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		x := int(args[0].(py.Int))
		z := int(args[1].(py.Int))
		return py.Bool(x == c.x && z == c.z), nil
	}, py.METH_STATIC, "")
	module.Globals["open_bank"] = py.MustNewMethod("open_bank", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		if c.optionMenuVisible {
			c.answer(0)
			s.bankOptionMenuTimeout = time.Now().Add(time.Second * 5)
			return py.Int(250), nil
		}
		if time.Now().UnixNano() < s.bankOptionMenuTimeout.UnixNano() {
			return py.Int(250), nil
		}
		npc := c.getNearestNPCByID(false, false, true, 0, 0, -1, bankersID...)
		if npc == nil {
			return py.Int(500), nil
		}
		if c.distanceTo(npc.x, npc.z) > 2 {
			c.walkTo(npc.x, npc.z, npc.x, npc.z, false, false)
			return py.Int(650), nil
		}
		c.talkToNPC(npc)
		s.bankOptionMenuTimeout = time.Now().Add(time.Second * 5)
		return py.Int(250), nil
	}, py.METH_STATIC, "")
	module.Globals["is_sleeping"] = py.MustNewMethod("is_sleeping", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.sleeping), nil
	}, py.METH_STATIC, "")
	module.Globals["random"] = py.MustNewMethod("random", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		min := int(args[0].(py.Int))
		max := int(args[1].(py.Int))
		return py.Int(min + rand.Intn(max-min+1)), nil
	}, py.METH_STATIC, "")
	module.Globals["set_fatigue_tricking"] = py.MustNewMethod("set_fatigue_tricking", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		if args[0].(py.Bool) {
			s.wakeupAt = 99
		} else {
			s.wakeupAt = 0
		}
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["set_wakeup_at"] = py.MustNewMethod("set_wakeup_at", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		s.wakeupAt = int(args[0].(py.Int))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["stop_script"] = py.MustNewMethod("stop_script", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		s.running = false
		c.scriptStopSignal <- struct{}{}
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["create_path"] = py.MustNewMethod("create_path", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		coords := args[0].(*py.List)
		points := make([]*point, coords.Len()/2)
		for i := 0; i < len(points); i++ {
			points[i] = &point{
				x: int(coords.Items[i*2].(py.Int)),
				y: int(coords.Items[i*2+1].(py.Int)),
			}
		}
		return &longPath{
			points: points,
			script: s,
		}, nil
	}, py.METH_STATIC, "")
	module.Globals["calculate_path"] = py.MustNewMethod("calculate_path", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		startX, startZ := int(args[0].(py.Int)), int(args[1].(py.Int))
		x, z := int(args[2].(py.Int)), int(args[3].(py.Int))
		depth := -1
		skipLocal := false
		maxDepth := 512
		if len(args) > 4 {
			switch v := args[4].(type) {
			case py.Int:
				depth = int(v)
			case py.Float:
				depth = int(float64(v))
			}
		}
		if skipLocal0, ok := kwargs["skip_local"]; ok {
			skipLocal = bool(skipLocal0.(py.Bool))
		}
		if maxDepth0, ok := kwargs["max_depth"]; ok {
			maxDepth = int(maxDepth0.(py.Int))
		}
		path := c.generateLongPath(startX, startZ, x, z, depth, skipLocal, maxDepth)
		if path == nil {
			if s.c.account.debug {
				fmt.Printf("[%s] Path not found from (%d, %d) to (%d, %d)\n", c.user, startX, startZ, x, z)
			}
			return py.None, nil
		}
		return path, nil
	}, py.METH_STATIC, "")
	module.Globals["calculate_path_to"] = py.MustNewMethod("calculate_path_to", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		x, z := int(args[0].(py.Int)), int(args[1].(py.Int))
		depth := -1
		skipLocal := false
		maxDepth := 512
		if len(args) > 2 {
			switch v := args[2].(type) {
			case py.Int:
				depth = int(v)
			case py.Float:
				depth = int(float64(v))
			}
		}
		if skipLocal0, ok := kwargs["skip_local"]; ok {
			skipLocal = bool(skipLocal0.(py.Bool))
		}
		if maxDepth0, ok := kwargs["max_depth"]; ok {
			maxDepth = int(maxDepth0.(py.Int))
		}
		path := c.generateLongPath(c.x, c.z, x, z, depth, skipLocal, maxDepth)
		if path == nil {
			if s.c.account.debug {
				fmt.Printf("[%s] Path not found to (%d, %d)\n", c.user, x, z)
			}
			return py.None, nil
		}
		return path, nil
	}, py.METH_STATIC, "")
	module.Globals["calculate_path_through"] = py.MustNewMethod("calculate_path_through", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		var (
			l         *py.List
			depth     = -1
			startX    = c.x
			startZ    = c.z
			skipLocal = false
			maxDepth  = 512
		)
		if l0, ok := kwargs["points"]; ok {
			l = l0.(*py.List)
		} else {
			return py.None, errors.New("`points` kwarg must be supplied")
		}
		coords := make([]*point, l.Len())
		for i, v := range l.Items {
			v := v.(py.Tuple)
			coords[i] = &point{
				x: int(v[0].(py.Int)),
				y: int(v[1].(py.Int)),
			}
		}
		if depth0, ok := kwargs["depth"]; ok {
			switch v := depth0.(type) {
			case py.Int:
				depth = int(v)
			case py.Float:
				depth = int(float64(v))
			}
		}
		if startX0, ok := kwargs["start_x"]; ok {
			startX = int(startX0.(py.Int))
		}
		if startZ0, ok := kwargs["start_z"]; ok {
			startZ = int(startZ0.(py.Int))
		}
		if skipLocal0, ok := kwargs["skip_local"]; ok {
			skipLocal = bool(skipLocal0.(py.Bool))
		}
		if maxDepth0, ok := kwargs["max_depth"]; ok {
			maxDepth = int(maxDepth0.(py.Int))
		}
		path, startError, endError := c.generateLongPathThrough(startX, startZ, coords, depth, skipLocal, maxDepth)
		if path == nil {
			if s.c.account.debug {
				fmt.Printf("[%s] Path not found from coords (%d, %d) to (%d, %d)\n", c.user, startError.x, startError.y, endError.x, endError.y)
			}
			return py.None, nil
		}
		return path, nil
	}, py.METH_STATIC, "")
	module.Globals["walk_path_to"] = py.MustNewMethod("walk_path_to", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		x, z := int(args[0].(py.Int)), int(args[1].(py.Int))
		pt := c.generateLongPathPoint(x, z, -1)
		if pt == nil {
			if s.c.account.debug {
				fmt.Printf("[%s] Path not found to (%d, %d)\n", c.user, x, z)
			}
			return py.Bool(false), nil
		}
		c.walkTo(pt.x, pt.y, pt.x, pt.y, false, false)
		return py.Bool(true), nil
	}, py.METH_STATIC, "")
	module.Globals["walk_path_depth_to"] = py.MustNewMethod("walk_path_depth_to", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		x, z := int(args[0].(py.Int)), int(args[1].(py.Int))
		var depth int
		switch v := args[2].(type) {
		case py.Int:
			depth = int(v)
		case py.Float:
			depth = int(float64(v))
		}
		pt := c.generateLongPathPoint(x, z, depth)
		if pt == nil {
			if s.c.account.debug {
				fmt.Printf("[%s] Path not found to (%d, %d)\n", c.user, x, z)
			}
			return py.Bool(false), nil
		}
		c.walkTo(pt.x, pt.y, pt.x, pt.y, false, false)
		return py.Bool(true), nil
	}, py.METH_STATIC, "")
	module.Globals["get_pid"] = py.MustNewMethod("get_pid", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Int(c.serverIndex), nil
	}, py.METH_STATIC, "")
	module.Globals["logout"] = py.MustNewMethod("logout", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.logout()
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["get_fatigue"] = py.MustNewMethod("get_fatigue", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Int(c.fatigue), nil
	}, py.METH_STATIC, "")
	module.Globals["get_accurate_fatigue"] = py.MustNewMethod("get_accurate_fatigue", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Float(c.accurateFatigue), nil
	}, py.METH_STATIC, "")
	module.Globals["get_x"] = py.MustNewMethod("get_x", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Int(c.x), nil
	}, py.METH_STATIC, "")
	module.Globals["get_z"] = py.MustNewMethod("get_z", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Int(c.z), nil
	}, py.METH_STATIC, "")
	module.Globals["in_combat"] = py.MustNewMethod("in_combat", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.inCombat()), nil
	}, py.METH_STATIC, "")
	module.Globals["is_skilling"] = py.MustNewMethod("is_skilling", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.playersServer[c.serverIndex].isSkilling()), nil
	}, py.METH_STATIC, "")
	module.Globals["is_talking"] = py.MustNewMethod("is_talking", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.playersServer[c.serverIndex].isTalking()), nil
	}, py.METH_STATIC, "")
	module.Globals["get_combat_style"] = py.MustNewMethod("get_combat_style", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Int(c.combatStyle), nil
	}, py.METH_STATIC, "")
	module.Globals["set_combat_style"] = py.MustNewMethod("set_combat_style", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.setCombatStyle(int(args[0].(py.Int)))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["get_max_stat"] = py.MustNewMethod("get_max_stat", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Int(c.baseStats[int(args[0].(py.Int))]), nil
	}, py.METH_STATIC, "")
	module.Globals["get_current_stat"] = py.MustNewMethod("get_current_stat", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Int(c.currentStats[int(args[0].(py.Int))]), nil
	}, py.METH_STATIC, "")
	module.Globals["get_experience"] = py.MustNewMethod("get_experience", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Int(c.experience[int(args[0].(py.Int))]), nil
	}, py.METH_STATIC, "")
	module.Globals["get_hp_percent"] = py.MustNewMethod("get_hp_percent", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Int(c.getHPPercent()), nil
	}, py.METH_STATIC, "")
	module.Globals["is_appearance_screen"] = py.MustNewMethod("is_appearance_screen", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.appearanceChange), nil
	}, py.METH_STATIC, "")
	module.Globals["cast_on_self"] = py.MustNewMethod("cast_on_self", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.castOnSelf(int(args[0].(py.Int)))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["get_total_inventory_count"] = py.MustNewMethod("get_total_inventory_count", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Int(c.inventoryCount), nil
	}, py.METH_STATIC, "")
	module.Globals["get_empty_slots"] = py.MustNewMethod("get_empty_slots", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Int(30 - c.inventoryCount), nil
	}, py.METH_STATIC, "")
	module.Globals["get_item_name"] = py.MustNewMethod("get_item_name", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.String(itemDefs[int(args[0].(py.Int))].Name), nil
	}, py.METH_STATIC, "")
	module.Globals["get_inventory_count_by_id"] = py.MustNewMethod("get_inventory_count_by_id", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		var ids []int

		if len(args) > 0 {
			ids = append(ids, int(args[0].(py.Int)))
		}

		if ids0, ok := kwargs["ids"]; ok {
			for _, id := range ids0.(*py.List).Items {
				ids = append(ids, int(id.(py.Int)))
			}
		}

		return py.Int(c.getInventoryCountByID(ids...)), nil
	}, py.METH_STATIC, "")
	module.Globals["get_inventory_item_except"] = py.MustNewMethod("get_inventory_item_except", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		ids := args[0].(*py.List)
		var selectedItem *item
	next:
		for i := 0; i < c.inventoryCount; i++ {
			it := c.inventory[i]
			for _, id := range ids.Items {
				if int(id.(py.Int)) == it.id {
					continue next
				}
			}
			selectedItem = it
			break
		}
		if selectedItem == nil {
			return py.None, nil
		}
		return selectedItem, nil
	}, py.METH_STATIC, "")
	module.Globals["has_inventory_item"] = py.MustNewMethod("has_inventory_item", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.getInventoryCountByID(int(args[0].(py.Int))) > 0), nil
	}, py.METH_STATIC, "")
	module.Globals["use_sleeping_bag"] = py.MustNewMethod("use_sleeping_bag", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.useSleepingBag()
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["skip_tutorial"] = py.MustNewMethod("skip_tutorial", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.skipTutorial()
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["is_inventory_item_equipped"] = py.MustNewMethod("is_inventory_item_equipped", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.isItemEquippedByID(int(args[0].(py.Int)))), nil
	}, py.METH_STATIC, "")
	module.Globals["get_quest_points"] = py.MustNewMethod("get_quest_points", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Int(c.questPoints), nil
	}, py.METH_STATIC, "")
	module.Globals["get_equipment_stat"] = py.MustNewMethod("get_equipment_stat", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Int(c.equipmentStats[int(args[0].(py.Int))]), nil
	}, py.METH_STATIC, "")
	module.Globals["set_autologin"] = py.MustNewMethod("set_autologin", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.account.autologin = bool(args[0].(py.Bool))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["set_debug"] = py.MustNewMethod("set_debug", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.account.debug = bool(args[0].(py.Bool))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["disconnect_for"] = py.MustNewMethod("disconnect_for", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		waitTime := int(args[0].(py.Int))
		c.scriptDCSignal <- waitTime
		return py.None, nil
	}, py.METH_STATIC, "")

	// world

	module.Globals["walk_to"] = py.MustNewMethod("walk_to", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		x := int(args[0].(py.Int))
		z := int(args[1].(py.Int))
		return py.Bool(c.walkTo(x, z, x, z, false, false)), nil
	}, py.METH_STATIC, "")
	module.Globals["walk_to_entity"] = py.MustNewMethod("walk_to_entity", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		x := int(args[0].(py.Int))
		z := int(args[1].(py.Int))
		return py.Bool(c.walkTo(x, z, x, z, true, false)), nil
	}, py.METH_STATIC, "")
	module.Globals["is_reachable"] = py.MustNewMethod("is_reachable", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.isReachable(int(args[0].(py.Int)), int(args[1].(py.Int)))), nil
	}, py.METH_STATIC, "")
	module.Globals["is_ground_item_reachable"] = py.MustNewMethod("is_ground_item_reachable", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.isGroundItemReachable(args[0].(*groundItem))), nil
	}, py.METH_STATIC, "")
	module.Globals["is_object_reachable"] = py.MustNewMethod("is_object_reachable", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.isObjectReachable(args[0].(*object))), nil
	}, py.METH_STATIC, "")
	module.Globals["is_wall_object_reachable"] = py.MustNewMethod("is_wall_object_reachable", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.isWallObjectReachable(args[0].(*wallObject))), nil
	}, py.METH_STATIC, "")
	module.Globals["distance_to"] = py.MustNewMethod("distance_to", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		x := int(args[0].(py.Int))
		z := int(args[1].(py.Int))
		return py.Int(c.distanceTo(x, z)), nil
	}, py.METH_STATIC, "")
	module.Globals["in_radius_of"] = py.MustNewMethod("in_radius_of", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		if len(args) < 3 {
			return nil, errors.New("x, z, and radius must be set in in_radius_of")
		}

		x := int(args[0].(py.Int))
		z := int(args[1].(py.Int))
		radius := int(args[2].(py.Int))

		return py.Bool(c.inRadiusOf(x, z, radius)), nil
	}, py.METH_STATIC, "")

	// ground items

	module.Globals["get_ground_item_count"] = py.MustNewMethod("get_ground_item_count", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Int(c.groundItemCount), nil
	}, py.METH_STATIC, "")
	module.Globals["get_ground_item_at_index"] = py.MustNewMethod("get_ground_item_at_index", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return c.groundItems[args[0].(py.Int)], nil
	}, py.METH_STATIC, "")
	module.Globals["get_ground_items"] = py.MustNewMethod("get_ground_items", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		objs := make([]py.Object, c.groundItemCount)
		for i := 0; i < c.groundItemCount; i++ {
			objs[i] = c.groundItems[i]
		}
		return py.NewListFromItems(objs), nil
	}, py.METH_STATIC, "")
	module.Globals["get_nearest_ground_item_by_id"] = py.MustNewMethod("get_nearest_ground_item_by_id", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		var (
			ids       []int
			reachable = false
			x         int
			z         int
			radius    = -1
		)

		if len(args) > 0 {
			ids = append(ids, int(args[0].(py.Int)))
			if len(args) > 1 {
				reachable = bool(args[1].(py.Bool))
				if len(args) > 2 {
					x = int(args[2].(py.Int))
					if len(args) > 3 {
						z = int(args[3].(py.Int))
						if len(args) > 4 {
							switch x := args[4].(type) {
							case py.Int:
								radius = int(x)
							case py.Float:
								radius = int(x)
							}
						}
					}
				}
			}
		}

		if ids0, ok := kwargs["ids"]; ok {
			idsList := ids0.(*py.List)
			for _, v := range idsList.Items {
				ids = append(ids, int(v.(py.Int)))
			}
		}

		if reachable0, ok := kwargs["reachable"]; ok {
			reachable = bool(reachable0.(py.Bool))
		}

		if x0, ok := kwargs["x"]; ok {
			x = int(x0.(py.Int))
		}

		if z0, ok := kwargs["z"]; ok {
			z = int(z0.(py.Int))
		}

		if radius0, ok := kwargs["radius"]; ok {
			switch x := radius0.(type) {
			case py.Int:
				radius = int(x)
			case py.Float:
				radius = int(x)
			}
		}

		if len(ids) < 1 {
			return nil, errors.New("must specify id in get_nearest_ground_item_by_id")
		}

		item := c.getNearestGroundItemByID(x, z, radius, reachable, ids...)
		if item == nil {
			return py.None, nil
		}

		return item, nil
	}, py.METH_STATIC, "")
	module.Globals["get_nearest_ground_item_by_id_in_rect"] = py.MustNewMethod("get_nearest_ground_item_by_id_in_rect", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		var (
			ids       []int
			x         = -1
			z         = -1
			width     = -1
			height    = -1
			reachable = false
		)

		if len(args) > 0 {
			ids = append(ids, int(args[0].(py.Int)))
			if len(args) > 1 {
				x = int(args[1].(py.Int))
				if len(args) > 2 {
					z = int(args[2].(py.Int))
					if len(args) > 3 {
						width = int(args[3].(py.Int))
						if len(args) > 4 {
							height = int(args[4].(py.Int))
							if len(args) > 5 {
								reachable = bool(args[5].(py.Bool))
							}
						}
					}
				}
			}
		}

		if ids0, ok := kwargs["ids"]; ok {
			idsList := ids0.(*py.List)
			for _, v := range idsList.Items {
				ids = append(ids, int(v.(py.Int)))
			}
		}

		if reachable0, ok := kwargs["reachable"]; ok {
			reachable = bool(reachable0.(py.Bool))
		}

		if x0, ok := kwargs["x"]; ok {
			x = int(x0.(py.Int))
		}
		if z0, ok := kwargs["z"]; ok {
			z = int(z0.(py.Int))
		}
		if width0, ok := kwargs["width"]; ok {
			width = int(width0.(py.Int))
		}
		if height0, ok := kwargs["height"]; ok {
			height = int(height0.(py.Int))
		}

		if len(ids) < 1 {
			return nil, errors.New("must specify id in get_nearest_ground_item_by_id_in_rect")
		}

		if x == -1 || z == -1 || width == -1 || height == -1 {
			return nil, errors.New("must specify x, z, width, and height in get_nearest_ground_item_by_id_in_rect")
		}

		obj := c.getNearestGroundItemByIDInRect(reachable, x, z, width, height, ids...)
		if obj == nil {
			return py.None, nil
		}

		return obj, nil
	}, py.METH_STATIC, "")
	module.Globals["is_ground_item_at"] = py.MustNewMethod("is_ground_item_at", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.isGroundItemAt(int(args[0].(py.Int)), int(args[1].(py.Int)), int(args[2].(py.Int)))), nil
	}, py.METH_STATIC, "")
	module.Globals["pickup_item"] = py.MustNewMethod("pickup_item", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.pickupGroundItem(args[0].(*groundItem))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["use_item_on_ground_item"] = py.MustNewMethod("use_item_on_ground_item", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.useInventoryItemOnGroundItem(args[0].(*item), args[1].(*groundItem))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["cast_on_ground_item"] = py.MustNewMethod("cast_on_ground_item", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.castOnGroundItem(int(args[0].(py.Int)), args[1].(*groundItem))
		return py.None, nil
	}, py.METH_STATIC, "")

	// inventory items

	module.Globals["get_inventory_items"] = py.MustNewMethod("get_inventory_items", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		objs := make([]py.Object, c.inventoryCount)
		for i := 0; i < c.inventoryCount; i++ {
			objs[i] = c.inventory[i]
		}
		return py.NewListFromItems(objs), nil
	}, py.METH_STATIC, "")
	module.Globals["get_inventory_item_by_id"] = py.MustNewMethod("get_inventory_item_by_id", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		var ids []int

		if len(args) > 0 {
			ids = append(ids, int(args[0].(py.Int)))
		}

		if ids0, ok := kwargs["ids"]; ok {
			for _, id := range ids0.(*py.List).Items {
				ids = append(ids, int(id.(py.Int)))
			}
		}

		item := c.getInventoryItemByID(ids...)
		if item == nil {
			return py.None, nil
		}

		return item, nil
	}, py.METH_STATIC, "")
	module.Globals["use_item"] = py.MustNewMethod("use_item", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.useInventoryItem(args[0].(*item))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["drop_item"] = py.MustNewMethod("drop_item", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.dropInventoryItem(args[0].(*item))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["equip_item"] = py.MustNewMethod("equip_item", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.equipInventoryItem(args[0].(*item))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["unequip_item"] = py.MustNewMethod("unequip_item", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.unequipInventoryItem(args[0].(*item))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["use_item_with_item"] = py.MustNewMethod("use_item_with_item", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.useInventoryItemOnInventoryItem(args[0].(*item), args[1].(*item))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["cast_on_item"] = py.MustNewMethod("cast_on_item", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.castOnInventoryItem(int(args[0].(py.Int)), args[1].(*item))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["use_item_on_object"] = py.MustNewMethod("use_item_on_object", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.useInventoryItemOnObject(args[0].(*item), args[1].(*object))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["use_item_on_wall_object"] = py.MustNewMethod("use_item_on_wall_object", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.useInventoryItemOnWallObject(args[0].(*item), args[1].(*wallObject))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["use_item_on_npc"] = py.MustNewMethod("use_item_on_npc", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.useInventoryItemOnNPC(args[0].(*item), args[1].(*npc))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["use_item_on_player"] = py.MustNewMethod("use_item_on_player", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.useInventoryItemOnPlayer(args[0].(*item), args[1].(*player))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["get_inventory_item_at_index"] = py.MustNewMethod("get_inventory_item_at_index", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return c.inventory[args[0].(py.Int)], nil
	}, py.METH_STATIC, "")

	// npcs
	module.Globals["get_npc_count"] = py.MustNewMethod("get_npc_count", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Int(c.npcCount), nil
	}, py.METH_STATIC, "")
	module.Globals["get_npc_at_index"] = py.MustNewMethod("get_npc_at_index", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return c.npcs[args[0].(py.Int)], nil
	}, py.METH_STATIC, "")
	module.Globals["get_npcs"] = py.MustNewMethod("get_npcs", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		objs := make([]py.Object, c.npcCount)
		for i := 0; i < c.npcCount; i++ {
			objs[i] = c.npcs[i]
		}
		return py.NewListFromItems(objs), nil
	}, py.METH_STATIC, "")
	module.Globals["get_nearest_npc_by_id"] = py.MustNewMethod("get_nearest_npc_by_id", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		var (
			ids       []int
			inCombat  = true
			talking   = true
			reachable = false
			x         int
			z         int
			radius    = -1
		)

		if len(args) > 0 {
			ids = append(ids, int(args[0].(py.Int)))
			if len(args) > 1 {
				inCombat = bool(args[1].(py.Bool))
				if len(args) > 2 {
					talking = bool(args[2].(py.Bool))
					if len(args) > 3 {
						reachable = bool(args[3].(py.Bool))
						if len(args) > 4 {
							x = int(args[4].(py.Int))
							if len(args) > 5 {
								z = int(args[5].(py.Int))
								if len(args) > 6 {
									switch x := args[6].(type) {
									case py.Int:
										radius = int(x)
									case py.Float:
										radius = int(x)
									}
								}
							}
						}
					}
				}
			}
		}

		if ids0, ok := kwargs["ids"]; ok {
			idsList := ids0.(*py.List)
			for _, v := range idsList.Items {
				ids = append(ids, int(v.(py.Int)))
			}
		}
		if inCombat0, ok := kwargs["in_combat"]; ok {
			inCombat = bool(inCombat0.(py.Bool))
		}

		if talking0, ok := kwargs["talking"]; ok {
			talking = bool(talking0.(py.Bool))
		}

		if reachable0, ok := kwargs["reachable"]; ok {
			reachable = bool(reachable0.(py.Bool))
		}

		if x0, ok := kwargs["x"]; ok {
			x = int(x0.(py.Int))
		}

		if z0, ok := kwargs["z"]; ok {
			z = int(z0.(py.Int))
		}

		if radius0, ok := kwargs["radius"]; ok {
			switch x := radius0.(type) {
			case py.Int:
				radius = int(x)
			case py.Float:
				radius = int(x)
			}
		}

		if len(ids) < 1 {
			return nil, errors.New("must specify id in get_nearest_npc_by_id")
		}

		npc := c.getNearestNPCByID(inCombat, talking, reachable, x, z, radius, ids...)
		if npc == nil {
			return py.None, nil
		}

		return npc, nil
	}, py.METH_STATIC, "")
	module.Globals["get_nearest_npc_by_id_in_rect"] = py.MustNewMethod("get_nearest_npc_by_id_in_rect", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		var (
			ids       []int
			x         = -1
			z         = -1
			width     = -1
			height    = -1
			inCombat  = true
			talking   = true
			reachable = false
		)

		if len(args) > 0 {
			ids = append(ids, int(args[0].(py.Int)))
			if len(args) > 1 {
				x = int(args[1].(py.Int))
				if len(args) > 2 {
					z = int(args[2].(py.Int))
					if len(args) > 3 {
						width = int(args[3].(py.Int))
						if len(args) > 4 {
							height = int(args[4].(py.Int))
							if len(args) > 5 {
								inCombat = bool(args[5].(py.Bool))
								if len(args) > 6 {
									talking = bool(args[6].(py.Bool))
									if len(args) > 7 {
										reachable = bool(args[7].(py.Bool))
									}
								}
							}
						}
					}

				}
			}
		}

		if ids0, ok := kwargs["ids"]; ok {
			idsList := ids0.(*py.List)
			for _, v := range idsList.Items {
				ids = append(ids, int(v.(py.Int)))
			}
		}
		if inCombat0, ok := kwargs["in_combat"]; ok {
			inCombat = bool(inCombat0.(py.Bool))
		}

		if talking0, ok := kwargs["talking"]; ok {
			talking = bool(talking0.(py.Bool))
		}

		if reachable0, ok := kwargs["reachable"]; ok {
			reachable = bool(reachable0.(py.Bool))
		}

		if x0, ok := kwargs["x"]; ok {
			x = int(x0.(py.Int))
		}
		if z0, ok := kwargs["z"]; ok {
			z = int(z0.(py.Int))
		}
		if width0, ok := kwargs["width"]; ok {
			width = int(width0.(py.Int))
		}
		if height0, ok := kwargs["height"]; ok {
			height = int(height0.(py.Int))
		}

		if len(ids) < 1 {
			return nil, errors.New("must specify id in get_nearest_npc_by_id_in_rect")
		}

		if x == -1 || z == -1 || width == -1 || height == -1 {
			return nil, errors.New("must specify x, z, width, and height in get_nearest_npc_by_id_in_rect")
		}

		npc := c.getNearestNPCByIDInRect(inCombat, talking, reachable, x, z, width, height, ids...)
		if npc == nil {
			return py.None, nil
		}

		return npc, nil
	}, py.METH_STATIC, "")
	module.Globals["attack_npc"] = py.MustNewMethod("attack_npc", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.attackNPC(args[0].(*npc))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["talk_to_npc"] = py.MustNewMethod("talk_to_npc", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.talkToNPC(args[0].(*npc))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["thieve_npc"] = py.MustNewMethod("thieve_npc", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.thieveNPC(args[0].(*npc))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["cast_on_npc"] = py.MustNewMethod("cast_on_npc", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.castOnNPC(int(args[0].(py.Int)), args[1].(*npc))
		return py.None, nil
	}, py.METH_STATIC, "")

	// quest menu

	module.Globals["is_option_menu"] = py.MustNewMethod("is_option_menu", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.optionMenuVisible), nil
	}, py.METH_STATIC, "")
	module.Globals["get_option_menu"] = py.MustNewMethod("get_option_menu", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		var menu []py.Object
		for i := 0; i < c.optionMenuCount; i++ {
			menu = append(menu, py.String(c.optionMenu[i]))
		}
		return py.NewListFromItems(menu), nil
	}, py.METH_STATIC, "")
	module.Globals["get_option_menu_count"] = py.MustNewMethod("get_option_menu_count", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Int(c.optionMenuCount), nil
	}, py.METH_STATIC, "")
	module.Globals["get_option_menu_option"] = py.MustNewMethod("get_option_menu_option", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.String(c.optionMenu[int(args[0].(py.Int))]), nil
	}, py.METH_STATIC, "")
	module.Globals["get_option_menu_index"] = py.MustNewMethod("get_option_menu_index", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		needle := string(args[0].(py.String))
		idx := -1
		for i := 0; i < c.optionMenuCount; i++ {
			if needle == c.optionMenu[i] {
				idx = i
				break
			}
		}
		return py.Int(idx), nil
	}, py.METH_STATIC, "")
	module.Globals["answer"] = py.MustNewMethod("answer", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.answer(int(args[0].(py.Int)))
		return py.None, nil
	}, py.METH_STATIC, "")

	// players

	module.Globals["get_my_player"] = py.MustNewMethod("get_my_player", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return c.playersServer[c.serverIndex], nil
	}, py.METH_STATIC, "")
	module.Globals["get_players"] = py.MustNewMethod("get_players", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		objs := make([]py.Object, c.playerCount)
		for i := 0; i < c.playerCount; i++ {
			objs[i] = c.players[i]
		}
		return py.NewListFromItems(objs), nil
	}, py.METH_STATIC, "")
	module.Globals["get_player_count"] = py.MustNewMethod("get_player_count", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Int(c.playerCount), nil
	}, py.METH_STATIC, "")
	module.Globals["get_player_at_index"] = py.MustNewMethod("get_player_at_index", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return c.players[args[0].(py.Int)], nil
	}, py.METH_STATIC, "")
	module.Globals["get_player_by_name"] = py.MustNewMethod("get_player_by_name", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		player := c.getPlayerByName(string(args[0].(py.String)))
		if player == nil {
			return py.None, nil
		}
		return player, nil
	}, py.METH_STATIC, "")
	module.Globals["attack_player"] = py.MustNewMethod("attack_player", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.attackPlayer(args[0].(*player))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["cast_on_player"] = py.MustNewMethod("cast_on_player", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.castOnPlayer(int(args[0].(py.Int)), args[1].(*player))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["trade_player"] = py.MustNewMethod("trade_player", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.tradePlayer(args[0].(*player))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["follow_player"] = py.MustNewMethod("follow_player", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.followPlayer(args[0].(*player))
		return py.None, nil
	}, py.METH_STATIC, "")

	// messaging

	module.Globals["send_chat_message"] = py.MustNewMethod("send_chat_message", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.sendChatMessage(string(args[0].(py.String)))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["send_private_message"] = py.MustNewMethod("send_private_message", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.sendPrivateMessage(string(args[0].(py.String)), string(args[1].(py.String)))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["add_friend"] = py.MustNewMethod("add_friend", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.addFriend(string(args[0].(py.String)))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["remove_friend"] = py.MustNewMethod("remove_friend", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.removeFriend(string(args[0].(py.String)))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["add_ignore"] = py.MustNewMethod("add_ignore", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.addIgnore(string(args[0].(py.String)))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["remove_ignore"] = py.MustNewMethod("remove_ignore", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.removeIgnore(string(args[0].(py.String)))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["get_friend_count"] = py.MustNewMethod("get_friend_count", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Int(c.friendListCount), nil
	}, py.METH_STATIC, "")
	module.Globals["get_friends"] = py.MustNewMethod("get_friends", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		objs := make([]py.Object, c.friendListCount)
		for i := 0; i < c.friendListCount; i++ {
			objs[i] = c.friendList[i]
		}
		return py.NewListFromItems(objs), nil
	}, py.METH_STATIC, "")
	module.Globals["is_friend"] = py.MustNewMethod("is_friend", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.isFriend(string(args[0].(py.String)))), nil
	}, py.METH_STATIC, "")
	module.Globals["is_ignored"] = py.MustNewMethod("is_ignored", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.isIgnored(string(args[0].(py.String)))), nil
	}, py.METH_STATIC, "")
	module.Globals["get_ignored"] = py.MustNewMethod("get_ignored", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		objs := make([]py.Object, c.ignoreListCount)
		for i := 0; i < c.ignoreListCount; i++ {
			objs[i] = c.ignoreList[i]
		}
		return py.NewListFromItems(objs), nil
	}, py.METH_STATIC, "")

	// trade

	module.Globals["get_my_trade_items"] = py.MustNewMethod("get_my_trade_items", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		objs := make([]py.Object, c.myTradeCount)
		for i := 0; i < c.myTradeCount; i++ {
			objs[i] = c.myTradeItems[i]
		}
		return py.NewListFromItems(objs), nil
	}, py.METH_STATIC, "")
	module.Globals["get_recipient_trade_items"] = py.MustNewMethod("get_recipient_trade_items", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		objs := make([]py.Object, c.recipientTradeCount)
		for i := 0; i < c.recipientTradeCount; i++ {
			objs[i] = c.recipientTradeItems[i]
		}
		return py.NewListFromItems(objs), nil
	}, py.METH_STATIC, "")
	module.Globals["get_my_confirm_items"] = py.MustNewMethod("get_my_confirm_items", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		objs := make([]py.Object, c.myTradeConfirmItemCount)
		for i := 0; i < c.myTradeConfirmItemCount; i++ {
			objs[i] = c.myTradeConfirmItems[i]
		}
		return py.NewListFromItems(objs), nil
	}, py.METH_STATIC, "")
	module.Globals["get_recipient_confirm_items"] = py.MustNewMethod("get_recipient_confirm_items", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		objs := make([]py.Object, c.recipientConfirmItemCount)
		for i := 0; i < c.recipientConfirmItemCount; i++ {
			objs[i] = c.recipientConfirmItems[i]
		}
		return py.NewListFromItems(objs), nil
	}, py.METH_STATIC, "")
	module.Globals["is_trade_offer_screen"] = py.MustNewMethod("is_trade_offer_screen", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.tradeScreen1Active), nil
	}, py.METH_STATIC, "")
	module.Globals["is_trade_confirm_screen"] = py.MustNewMethod("is_trade_offer_screen", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.tradeScreen2Active), nil
	}, py.METH_STATIC, "")
	module.Globals["trade_offer_item"] = py.MustNewMethod("trade_offer_item", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.offerItemToTrade(int(args[0].(py.Int)), args[1].(*item))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["has_my_offer"] = py.MustNewMethod("has_my_offer", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.hasTradeOffer(int(args[0].(py.Int)), int(args[1].(py.Int)))), nil
	}, py.METH_STATIC, "")
	module.Globals["has_my_confirm"] = py.MustNewMethod("has_my_confirm", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.hasTradeConfirm(int(args[0].(py.Int)), int(args[1].(py.Int)))), nil
	}, py.METH_STATIC, "")
	module.Globals["has_recipient_offer"] = py.MustNewMethod("has_recipient_offer", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.hasRecipientOffer(int(args[0].(py.Int)), int(args[1].(py.Int)))), nil
	}, py.METH_STATIC, "")
	module.Globals["has_recipient_confirm"] = py.MustNewMethod("has_recipient_confirm", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.hasRecipientConfirm(int(args[0].(py.Int)), int(args[1].(py.Int)))), nil
	}, py.METH_STATIC, "")
	module.Globals["is_trade_accepted"] = py.MustNewMethod("is_trade_accepted", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.tradeAccepted), nil
	}, py.METH_STATIC, "")
	module.Globals["is_recipient_trade_accepted"] = py.MustNewMethod("is_recipient_trade_accepted", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.tradeRecipientAccepted), nil
	}, py.METH_STATIC, "")
	module.Globals["is_trade_confirm_accepted"] = py.MustNewMethod("is_trade_confirm_accepted", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.tradeConfirmAccepted), nil
	}, py.METH_STATIC, "")
	module.Globals["accept_trade_offer"] = py.MustNewMethod("accept_trade_offer", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.acceptTradeOffer()
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["confirm_trade"] = py.MustNewMethod("confirm_trade", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.acceptTradeConfirm()
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["decline_trade"] = py.MustNewMethod("decline_trade", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.declineTrade()
		return py.None, nil
	}, py.METH_STATIC, "")

	// objects

	module.Globals["get_objects"] = py.MustNewMethod("get_objects", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		objs := make([]py.Object, c.objectCount)
		for i := 0; i < c.objectCount; i++ {
			objs[i] = c.objects[i]
		}
		return py.NewListFromItems(objs), nil
	}, py.METH_STATIC, "")
	module.Globals["get_nearest_object_by_id"] = py.MustNewMethod("get_nearest_object_by_id", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		var (
			ids       []int
			reachable = false
			x         int
			z         int
			radius    = -1
		)

		if len(args) > 0 {
			ids = append(ids, int(args[0].(py.Int)))
			if len(args) > 1 {
				reachable = bool(args[1].(py.Bool))
				if len(args) > 2 {
					x = int(args[2].(py.Int))
					if len(args) > 3 {
						z = int(args[3].(py.Int))
						if len(args) > 4 {
							switch x := args[4].(type) {
							case py.Int:
								radius = int(x)
							case py.Float:
								radius = int(x)
							}
						}
					}
				}
			}
		}

		if ids0, ok := kwargs["ids"]; ok {
			idsList := ids0.(*py.List)
			for _, v := range idsList.Items {
				ids = append(ids, int(v.(py.Int)))
			}
		}

		if reachable0, ok := kwargs["reachable"]; ok {
			reachable = bool(reachable0.(py.Bool))
		}

		if x0, ok := kwargs["x"]; ok {
			x = int(x0.(py.Int))
		}

		if z0, ok := kwargs["z"]; ok {
			z = int(z0.(py.Int))
		}

		if radius0, ok := kwargs["radius"]; ok {
			switch x := radius0.(type) {
			case py.Int:
				radius = int(x)
			case py.Float:
				radius = int(x)
			}
		}

		if len(ids) < 1 {
			return nil, errors.New("must specify id in get_nearest_object_by_id")
		}

		obj := c.getNearestObjectByID(x, z, radius, reachable, ids...)
		if obj == nil {
			return py.None, nil
		}

		return obj, nil
	}, py.METH_STATIC, "")
	module.Globals["get_nearest_object_by_id_in_rect"] = py.MustNewMethod("get_nearest_object_by_id_in_rect", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		var (
			ids       []int
			x         = -1
			z         = -1
			width     = -1
			height    = -1
			reachable = false
		)

		if len(args) > 0 {
			ids = append(ids, int(args[0].(py.Int)))
			if len(args) > 1 {
				x = int(args[1].(py.Int))
				if len(args) > 2 {
					z = int(args[2].(py.Int))
					if len(args) > 3 {
						width = int(args[3].(py.Int))
						if len(args) > 4 {
							height = int(args[4].(py.Int))
							if len(args) > 5 {
								reachable = bool(args[5].(py.Bool))
							}
						}
					}
				}
			}
		}

		if ids0, ok := kwargs["ids"]; ok {
			idsList := ids0.(*py.List)
			for _, v := range idsList.Items {
				ids = append(ids, int(v.(py.Int)))
			}
		}

		if reachable0, ok := kwargs["reachable"]; ok {
			reachable = bool(reachable0.(py.Bool))
		}

		if x0, ok := kwargs["x"]; ok {
			x = int(x0.(py.Int))
		}
		if z0, ok := kwargs["z"]; ok {
			z = int(z0.(py.Int))
		}
		if width0, ok := kwargs["width"]; ok {
			width = int(width0.(py.Int))
		}
		if height0, ok := kwargs["height"]; ok {
			height = int(height0.(py.Int))
		}

		if len(ids) < 1 {
			return nil, errors.New("must specify id in get_nearest_object_by_id_in_rect")
		}

		if x == -1 || z == -1 || width == -1 || height == -1 {
			return nil, errors.New("must specify x, z, width, and height in get_nearest_object_by_id_in_rect")
		}

		obj := c.getNearestObjectByIDInRect(reachable, x, z, width, height, ids...)
		if obj == nil {
			return py.None, nil
		}

		return obj, nil
	}, py.METH_STATIC, "")
	module.Globals["get_object_count"] = py.MustNewMethod("get_object_count", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Int(c.objectCount), nil
	}, py.METH_STATIC, "")
	module.Globals["get_object_at_index"] = py.MustNewMethod("get_object_at_index", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return c.objects[args[0].(py.Int)], nil
	}, py.METH_STATIC, "")
	module.Globals["get_object_from_coords"] = py.MustNewMethod("get_object_from_coords", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		obj := c.getObjectFromCoords(int(args[0].(py.Int)), int(args[1].(py.Int)))
		if obj == nil {
			return py.None, nil
		}
		return obj, nil
	}, py.METH_STATIC, "")
	module.Globals["is_object_at"] = py.MustNewMethod("is_object_at", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.isObjectAt(int(args[0].(py.Int)), int(args[1].(py.Int)))), nil
	}, py.METH_STATIC, "")
	module.Globals["at_object"] = py.MustNewMethod("at_object", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.atObject(args[0].(*object))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["at_object2"] = py.MustNewMethod("at_object2", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.atObject2(args[0].(*object))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["cast_on_object"] = py.MustNewMethod("cast_on_object", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.castOnObject(int(args[0].(py.Int)), args[1].(*object))
		return py.None, nil
	}, py.METH_STATIC, "")

	// wall objects

	module.Globals["get_wall_objects"] = py.MustNewMethod("get_wall_objects", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		objs := make([]py.Object, c.wallObjectCount)
		for i := 0; i < c.wallObjectCount; i++ {
			objs[i] = c.wallObjects[i]
		}
		return py.NewListFromItems(objs), nil
	}, py.METH_STATIC, "")
	module.Globals["get_nearest_wall_object_by_id"] = py.MustNewMethod("get_nearest_wall_object_by_id", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		var (
			ids       []int
			reachable = false
			x         int
			z         int
			radius    = -1
		)

		if len(args) > 0 {
			ids = append(ids, int(args[0].(py.Int)))
			if len(args) > 1 {
				reachable = bool(args[1].(py.Bool))
				if len(args) > 2 {
					x = int(args[2].(py.Int))
					if len(args) > 3 {
						z = int(args[3].(py.Int))
						if len(args) > 4 {
							switch x := args[4].(type) {
							case py.Int:
								radius = int(x)
							case py.Float:
								radius = int(x)
							}
						}
					}
				}
			}
		}

		if ids0, ok := kwargs["ids"]; ok {
			idsList := ids0.(*py.List)
			for _, v := range idsList.Items {
				ids = append(ids, int(v.(py.Int)))
			}
		}

		if radius0, ok := kwargs["radius"]; ok {
			switch x := radius0.(type) {
			case py.Int:
				radius = int(x)
			case py.Float:
				radius = int(x)
			}
		}

		if reachable0, ok := kwargs["reachable"]; ok {
			reachable = bool(reachable0.(py.Bool))
		}

		if x0, ok := kwargs["x"]; ok {
			x = int(x0.(py.Int))
		}

		if z0, ok := kwargs["z"]; ok {
			z = int(z0.(py.Int))
		}

		if len(ids) < 1 {
			return nil, errors.New("must specify id in get_nearest_wall_object_by_id")
		}

		obj := c.getNearestWallObjectByID(x, z, radius, reachable, ids...)
		if obj == nil {
			return py.None, nil
		}

		return obj, nil
	}, py.METH_STATIC, "")
	module.Globals["get_nearest_wall_object_by_id_in_rect"] = py.MustNewMethod("get_nearest_wall_object_by_id_in_rect", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		var (
			ids       []int
			x         = -1
			z         = -1
			width     = -1
			height    = -1
			reachable = false
		)

		if len(args) > 0 {
			ids = append(ids, int(args[0].(py.Int)))
			if len(args) > 1 {
				x = int(args[1].(py.Int))
				if len(args) > 2 {
					z = int(args[2].(py.Int))
					if len(args) > 3 {
						width = int(args[3].(py.Int))
						if len(args) > 4 {
							height = int(args[4].(py.Int))
							if len(args) > 5 {
								reachable = bool(args[5].(py.Bool))
							}
						}
					}
				}
			}
		}

		if ids0, ok := kwargs["ids"]; ok {
			idsList := ids0.(*py.List)
			for _, v := range idsList.Items {
				ids = append(ids, int(v.(py.Int)))
			}
		}

		if reachable0, ok := kwargs["reachable"]; ok {
			reachable = bool(reachable0.(py.Bool))
		}

		if x0, ok := kwargs["x"]; ok {
			x = int(x0.(py.Int))
		}
		if z0, ok := kwargs["z"]; ok {
			z = int(z0.(py.Int))
		}
		if width0, ok := kwargs["width"]; ok {
			width = int(width0.(py.Int))
		}
		if height0, ok := kwargs["height"]; ok {
			height = int(height0.(py.Int))
		}

		if len(ids) < 1 {
			return nil, errors.New("must specify id in get_nearest_wall_object_by_id_in_rect")
		}

		if x == -1 || z == -1 || width == -1 || height == -1 {
			return nil, errors.New("must specify x, z, width, and height in get_nearest_wall_object_by_id_in_rect")
		}

		obj := c.getNearestWallObjectByIDInRect(reachable, x, z, width, height, ids...)
		if obj == nil {
			return py.None, nil
		}

		return obj, nil
	}, py.METH_STATIC, "")
	module.Globals["get_wall_object_count"] = py.MustNewMethod("get_wall_object_count", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Int(c.wallObjectCount), nil
	}, py.METH_STATIC, "")
	module.Globals["get_wall_object_at_index"] = py.MustNewMethod("get_wall_object_at_index", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return c.wallObjects[args[0].(py.Int)], nil
	}, py.METH_STATIC, "")
	module.Globals["get_wall_object_from_coords"] = py.MustNewMethod("get_wall_object_from_coords", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		obj := c.getWallObjectFromCoords(int(args[0].(py.Int)), int(args[1].(py.Int)))
		if obj == nil {
			return py.None, nil
		}
		return obj, nil
	}, py.METH_STATIC, "")
	module.Globals["is_wall_object_at"] = py.MustNewMethod("is_wall_object_at", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.isWallObjectAt(int(args[0].(py.Int)), int(args[1].(py.Int)))), nil
	}, py.METH_STATIC, "")
	module.Globals["at_wall_object"] = py.MustNewMethod("at_wall_object", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.atWallObject(args[0].(*wallObject))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["at_wall_object2"] = py.MustNewMethod("at_wall_object2", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.atWallObject2(args[0].(*wallObject))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["cast_on_wall_object"] = py.MustNewMethod("cast_on_wall_object", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.castOnWallObject(int(args[0].(py.Int)), args[1].(*wallObject))
		return py.None, nil
	}, py.METH_STATIC, "")

	// banking
	module.Globals["get_bank_items"] = py.MustNewMethod("get_bank_items", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		objs := make([]py.Object, c.bankItemCount)
		for i := 0; i < c.bankItemCount; i++ {
			objs[i] = c.bankItems[i]
		}
		return py.NewListFromItems(objs), nil
	}, py.METH_STATIC, "")
	module.Globals["deposit"] = py.MustNewMethod("deposit", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.deposit(int(args[0].(py.Int)), int(args[1].(py.Int)))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["withdraw"] = py.MustNewMethod("withdraw", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.withdraw(int(args[0].(py.Int)), int(args[1].(py.Int)))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["get_bank_count"] = py.MustNewMethod("get_bank_count", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		var (
			ids []int
		)

		if len(args) > 0 {
			ids = append(ids, int(args[0].(py.Int)))
		}

		if ids0, ok := kwargs["ids"]; ok {
			idsList := ids0.(*py.List)
			for _, v := range idsList.Items {
				ids = append(ids, int(v.(py.Int)))
			}
		}

		if len(ids) < 1 {
			return nil, errors.New("must specify id in get_bank_count")
		}

		return py.Int(c.bankCount(ids...)), nil
	}, py.METH_STATIC, "")
	module.Globals["get_bank_size"] = py.MustNewMethod("get_bank_size", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Int(c.bankItemCount), nil
	}, py.METH_STATIC, "")
	module.Globals["get_bank_item_at_index"] = py.MustNewMethod("get_bank_item_at_index", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return c.bankItems[args[0].(py.Int)], nil
	}, py.METH_STATIC, "")
	module.Globals["has_bank_item"] = py.MustNewMethod("has_bank_item", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.hasBankItem(int(args[0].(py.Int)))), nil
	}, py.METH_STATIC, "")
	module.Globals["is_bank_open"] = py.MustNewMethod("is_bank_open", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.banking), nil
	}, py.METH_STATIC, "")
	module.Globals["close_bank"] = py.MustNewMethod("close_bank", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.closeBank()
		return py.None, nil
	}, py.METH_STATIC, "")

	// prayers
	module.Globals["enable_prayer"] = py.MustNewMethod("enable_prayer", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.enablePrayer(int(args[0].(py.Int)))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["disable_prayer"] = py.MustNewMethod("disable_prayer", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.disablePrayer(int(args[0].(py.Int)))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["is_prayer_enabled"] = py.MustNewMethod("is_prayer_enabled", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.prayers[int(args[0].(py.Int))]), nil
	}, py.METH_STATIC, "")

	// shops
	module.Globals["get_shop_items"] = py.MustNewMethod("get_shop_items", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		objs := make([]py.Object, c.shopItemCount)
		for i := 0; i < c.shopItemCount; i++ {
			objs[i] = c.shopItems[i]
		}
		return py.NewListFromItems(objs), nil
	}, py.METH_STATIC, "")
	module.Globals["get_shop_item_count"] = py.MustNewMethod("get_shop_item_count", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Int(c.shopItemCount), nil
	}, py.METH_STATIC, "")
	module.Globals["get_shop_item_at_index"] = py.MustNewMethod("get_shop_item_at_index", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return c.shopItems[args[0].(py.Int)], nil
	}, py.METH_STATIC, "")
	module.Globals["is_shop_open"] = py.MustNewMethod("is_shop_open", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.shopping), nil
	}, py.METH_STATIC, "")
	module.Globals["get_shop_item_by_id"] = py.MustNewMethod("get_shop_item_by_id", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		v := c.getShopItemByID(int(args[0].(py.Int)))
		if v == nil {
			return py.None, nil
		}
		return v, nil
	}, py.METH_STATIC, "")
	module.Globals["buy_shop_item"] = py.MustNewMethod("buy_shop_item", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.buyItem(int(args[0].(py.Int)), int(args[1].(py.Int)))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["sell_shop_item"] = py.MustNewMethod("sell_shop_item", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.sellItem(int(args[0].(py.Int)), int(args[1].(py.Int)))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["close_shop"] = py.MustNewMethod("close_shop", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.closeShop()
		return py.None, nil
	}, py.METH_STATIC, "")

	// quests
	module.Globals["get_quests"] = py.MustNewMethod("get_quests", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		objs := make([]py.Object, c.questCount)
		for i := 0; i < c.questCount; i++ {
			objs[i] = c.quests[i]
		}
		return py.NewListFromItems(objs), nil
	}, py.METH_STATIC, "")
	module.Globals["get_quest"] = py.MustNewMethod("get_quest", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return c.quests[int(args[0].(py.Int))], nil
	}, py.METH_STATIC, "")
	module.Globals["is_quest_complete"] = py.MustNewMethod("is_quest_complete", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.quests[int(args[0].(py.Int))].isComplete()), nil
	}, py.METH_STATIC, "")

	// misc

	module.Globals["send_appearance_update"] = py.MustNewMethod("send_appearance_update", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		c.sendAppearanceUpdate(int(args[0].(py.Int)), int(args[1].(py.Int)), int(args[2].(py.Int)), int(args[3].(py.Int)), int(args[4].(py.Int)), int(args[5].(py.Int)), int(args[6].(py.Int)))
		return py.None, nil
	}, py.METH_STATIC, "")
	module.Globals["is_system_update"] = py.MustNewMethod("is_system_update", func(self py.Object, args py.Tuple, kwargs py.StringDict) (py.Object, error) {
		return py.Bool(c.systemUpdate), nil
	}, py.METH_STATIC, "")
}
