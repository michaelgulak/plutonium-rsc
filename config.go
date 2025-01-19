package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-python/gpython/py"
	"github.com/pelletier/go-toml/v2"
)

const (
	noProgressDurSet = time.Duration(math.MaxInt)
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
	needsFirstNotif    bool
}

type ocrSettings struct {
	OcrType     string `toml:"ocr_type"`
	OcrEndpoint string `toml:"ocr_endpoint"`
	Directory   string `toml:"directory"`
	OcrFileName string `toml:"ocr_filename"`
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
	RSAKeyFile  string `toml:"rsa_key_file"`
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
	Directory                string `toml:"directory"`
	ProgressReports          string `toml:"progress_reports"`
	OverwriteProgressReports bool   `toml:"overwrite_progress_reports"`
}

type config struct {
	executablePath  string
	ServerSettings  *serverSettings  `toml:"server"`
	OcrSettings     *ocrSettings     `toml:"ocr"`
	DataSettings    *dataSettings    `toml:"data"`
	AccountSettings *accountSettings `toml:"accounts"`
	ClientSettings  *clientSettings  `toml:"client"`
	ScriptSettings  *scriptSettings  `toml:"script"`
	LogSettings     *logSettings     `toml:"logs"`
}

type accountConfig struct {
	Account struct {
		User      string
		Pass      string
		Autologin interface{}
		Enabled   interface{}
		Debug     interface{}
	}
	Script interface{}
}

func parseConfig(filePath string) {
	ex, err := os.Executable()
	if err != nil {
		fmt.Println("[BOT] Error locating binary")
		os.Exit(1)
	}
	ex = filepath.Dir(ex)
	if filePath == "" {
		filePath = filepath.Join(ex, "settings.toml")
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

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var config accountConfig
	if err = toml.NewDecoder(f).Decode(&config); err != nil {
		return nil, err
	}

	acct := &account{
		username:        config.Account.User,
		password:        config.Account.Pass,
		script:          &script{},
		needsFirstNotif: true,
	}

	if autologin, ok := config.Account.Autologin.(bool); ok {
		acct.autologin = autologin
	} else {
		acct.autologin = true
	}

	if enabled, ok := config.Account.Enabled.(bool); ok {
		acct.enabled = enabled
	} else {
		acct.enabled = true
	}

	if debug, ok := config.Account.Debug.(bool); ok {
		acct.debug = debug
	}

	if script, ok := config.Script.(map[string]interface{}); ok {
		name := script["name"].(string)
		acct.script.path = filepath.Join(settings.executablePath, settings.ScriptSettings.Directory, name)
		acct.script.pathDir = filepath.Dir(acct.script.path)
		if pr, ok := script["progress_report"]; ok {
			dur, err := time.ParseDuration(pr.(string))
			if err != nil || dur < 0 {
				return nil, fmt.Errorf("invalid progress report field: %s", err)
			}
			acct.progressDur = dur
		} else {
			acct.progressDur = noProgressDurSet
		}
		if scriptArgs, ok := script["settings"]; ok {
			acct.script.settings = py.StringDict{}

			var convertType func(v interface{}) py.Object
			convertType = func(v interface{}) py.Object {
				switch v := v.(type) {
				case bool:
					return py.Bool(v)
				case uint64:
					return py.Int(v)
				case int64:
					return py.Int(v)
				case string:
					return py.String(v)
				case float64:
					return py.Float(v)
				case toml.LocalDateTime:
					return py.Float(v.AsTime(time.UTC).Unix())
				case toml.LocalDate:
					return py.Float(v.AsTime(time.UTC).Unix())
				case []interface{}:
					if len(v) == 0 {
						return py.NewList()
					}
					var objs []py.Object
					for _, v0 := range v {
						objs = append(objs, convertType(v0))
					}
					return py.NewListFromItems(objs)
				case map[string]interface{}:
					dict := py.StringDict{}
					for k, v0 := range v {
						dict[k] = convertType(v0)
					}
					return dict
				default: // toml.LocalTime is unhandled
					return nil
				}
			}
			for k, v := range scriptArgs.(map[string]interface{}) {
				setting := convertType(v)
				if setting == nil {
					return nil, fmt.Errorf("unhandled type: %v", v)
				}
				acct.script.settings[k] = setting
			}
		}
	}

	if acct.enabled {
		fmt.Println("enabled")
	} else {
		fmt.Println("disabled")
	}

	return acct, nil
}

func loadAccounts() ([]*account, error) {
	files, err := os.ReadDir(filepath.Join(settings.executablePath, settings.AccountSettings.Directory))
	if err != nil {
		return nil, err
	}
	var accounts []*account
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".toml") {
			continue
		}
		acct, err := loadAccount(filepath.Join(settings.executablePath, settings.AccountSettings.Directory, f.Name()))
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acct)
	}
	return accounts, nil
}
