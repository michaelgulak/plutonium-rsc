package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/go-python/gpython/py"
	"github.com/pelletier/go-toml"
)

var (
	settings config
)

type account struct {
	username           string
	password           string
	autologin          bool
	enabled            bool
	debug              bool
	script             *script
	progressFile       *os.File
	progressDur        time.Duration
	lastProgressReport time.Time
}

type dataSettings struct {
	Directory   string `toml:"directory"`
	Landscape   string `toml:"landscape"`
	ItemDefs    string `toml:"item_defs"`
	SceneryLocs string `toml:"scenery_locs"`
	TileDefs    string `toml:"tile_defs"`
	DoorDefs    string `toml:"door_defs"`
	ObjectDefs  string `toml:"object_defs"`
	WordHashes  string `toml:"word_hashes"`
}

type accountSettings struct {
	Directory string `toml:"directory"`
}

type clientSettings struct {
	ClientVersion int `toml:"client_version"`
}

type scriptSettings struct {
	Directory string `toml:"directory"`
}

type serverSettings struct {
	Address string `toml:"address"`
}

type logSettings struct {
	Directory       string `toml:"directory"`
	ProgressReports string `toml:"progress_reports"`
}

type config struct {
	executablePath  string
	ServerSettings  *serverSettings  `toml:"server"`
	DataSettings    *dataSettings    `toml:"data"`
	AccountSettings *accountSettings `toml:"accounts"`
	ClientSettings  *clientSettings  `toml:"client"`
	ScriptSettings  *scriptSettings  `toml:"script"`
	LogSettings     *logSettings     `toml:"logs"`
}

func parseConfig(filePath string) {
	ex, err := os.Executable()
	if err != nil {
		fmt.Println("[BOT] Error locating binary")
		os.Exit(1)
	}
	ex = filepath.Dir(ex)
	if filePath == "settings.toml" {
		filePath = ex + string(os.PathSeparator) + filePath
	}

	f, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening settings file: %s\n", err)
		os.Exit(1)
	}
	defer f.Close()
	err = toml.NewDecoder(f).Decode(&settings)
	if err != nil {
		fmt.Printf("Error parsing settings file: %s\n", err)
		os.Exit(1)
	}
	settings.executablePath = ex
}

func loadAccount(path string) (*account, error) {
	fmt.Printf("[BOT] Loading account [%s]...", filepath.Base(path))

	acctFile, err := toml.LoadFile(path)
	if err != nil {
		return nil, err
	}
	acct := &account{
		username: acctFile.Get("account.user").(string),
		password: acctFile.Get("account.pass").(string),
		script:   &script{},
	}
	if acctFile.Has("account.autologin") {
		acct.autologin = acctFile.Get("account.autologin").(bool)
	}
	if acctFile.Has("account.enabled") {
		acct.enabled = acctFile.Get("account.enabled").(bool)
	} else {
		acct.enabled = true
	}
	if acctFile.Has("account.debug") {
		acct.debug = acctFile.Get("account.debug").(bool)
	}
	if acctFile.Has("script") {
		acct.script.name = settings.executablePath + string(os.PathSeparator) + settings.ScriptSettings.Directory + string(os.PathSeparator) + acctFile.Get("script.name").(string)
		acct.script.pathDir = filepath.Dir(acct.script.name)
		if acctFile.Has("script.progress_report") {
			str := acctFile.Get("script.progress_report").(string)
			dur, err := time.ParseDuration(str)
			if err != nil {
				return nil, fmt.Errorf("invalid progress report field: %s", err)
			}
			acct.progressDur = dur
		} else {
			acct.progressDur = time.Duration(math.MaxInt)
		}
		if acctFile.Has("script.settings") {
			scriptArgs := acctFile.Get("script.settings").(*toml.Tree).ToMap()
			acct.script.settings = py.StringDict{}
			for k, v := range scriptArgs {
				switch v := v.(type) {
				case map[string]interface{}:
					return nil, errors.New("key/value pairs aren't supported")
				case time.Time:
					return nil, errors.New("dates aren't supported")
				case bool:
					acct.script.settings[k] = py.Bool(v)
				case uint64:
					acct.script.settings[k] = py.Int(v)
				case int64:
					acct.script.settings[k] = py.Int(v)
				case []interface{}:
					if len(v) == 0 {
						acct.script.settings[k] = py.NewList()
					} else {
						switch v[0].(type) {
						case []interface{}:
							return nil, errors.New("multidimensional arrays aren't supported")
						case time.Time:
							return nil, errors.New("dates aren't supported")
						case map[string]interface{}:
							return nil, errors.New("key/value pairs aren't supported")
						case int64:
							var objs []py.Object
							for _, v0 := range v {
								objs = append(objs, py.Int(v0.(int64)))
							}
							acct.script.settings[k] = py.NewListFromItems(objs)
						case uint64:
							var objs []py.Object
							for _, v0 := range v {
								objs = append(objs, py.Int(v0.(uint64)))
							}
							acct.script.settings[k] = py.NewListFromItems(objs)
						case string:
							var objs []py.Object
							for _, v0 := range v {
								objs = append(objs, py.String(v0.(string)))
							}
							acct.script.settings[k] = py.NewListFromItems(objs)
						case float64:
							var objs []py.Object
							for _, v0 := range v {
								objs = append(objs, py.Float(v0.(float64)))
							}
							acct.script.settings[k] = py.NewListFromItems(objs)
						default:
							return nil, fmt.Errorf("unhandled type: %v", reflect.ValueOf(v))
						}
					}
				case string:
					acct.script.settings[k] = py.String(v)
				case float64:
					acct.script.settings[k] = py.Float(v)
				default:
					return nil, fmt.Errorf("unhandled type: %v", reflect.ValueOf(v))
				}
			}
		}
	}

	fmt.Println("complete")

	return acct, nil
}

func loadAccounts() ([]*account, error) {
	files, err := os.ReadDir(settings.executablePath + string(os.PathSeparator) + settings.AccountSettings.Directory)
	if err != nil {
		return nil, err
	}
	var accounts []*account
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".toml") {
			continue
		}
		acct, err := loadAccount(settings.executablePath + string(os.PathSeparator) + settings.AccountSettings.Directory + string(os.PathSeparator) + f.Name())
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acct)
	}
	return accounts, nil
}
