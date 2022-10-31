# Plutonium

Plutonium is a shell bot written by RLN for the [Open RSC](https://rsc.vet/) Uranium server. It's built from the ground up using Go. Scripts for it are written in python via [gpython](https://github.com/go-python/gpython). It was written for fun and may contain bugs. Please file any issues you may encounter. Some code is taken from the Open RSC server (in particular, the pathfinding and chat message code). If you write a script for it, please send a PR so people can use it!

## Download

You can download the bot for linux_amd64 or windows_amd64 from the [releases page](https://gitlab.com/RLN_/plutonium/-/releases). Otherwise you'll need to build the bot and FOCR yourself. 

## Account Configuration

To run Plutonium, you need account files. Each account file specifies the username and password of an account, and the script to run for that account. An account config is created by creating a file in the `accounts` directory. These files must be in [toml](https://toml.io/en/) format with the exception that it cannot contain key/value pairs, multidimensional arrays, or dates.

To get started quickly, rename the example file in the `accounts` directory to be `myaccount.toml`. Then input your username and password, and make sure to change the script and script settings. Windows users should note that [file name extensions](https://support.microsoft.com/en-us/windows/common-file-name-extensions-in-windows-da4a4430-8e76-89c5-59f7-1cdbbc75cb01) must be enabled to rename the file properly.

Example file:

```toml
[account]
user = "myuser"
pass = "mypass"
autologin = true
enabled = true
debug = true

[script]
name = "varrock_east_miner.py"
progress_report = "20m"

[script.settings]
ore_type = "iron"
powermine = false
```

Feel free to change the `debug` field to `false` if you're running several accounts and it's spamming the console.

Note that the progress report field can contain values that `time.ParseDuration` in Go can parse. Check [here](https://pkg.go.dev/time#ParseDuration) for the list of possible suffixes.

## Run bot

On Linux make sure your working directory is the bot directory and run:

```bash
./bot
``` 

On Windows, open a command prompt window in the root Plutonium directory. Then type `bot.exe` and hit enter.

If you want to run just one instance of a bot for any reason, for example to skip the tutorial you can do:

```bash
./bot -u myuser -p mypass -s scripts/skip_tutorial.py
```

Note that you can't specify script settings with this method.

Alternatively, you can run a single account file like this:

```bash
./bot -f /path/to/account_file.toml
```

## Building bot

Make sure you have a new version of [Go](https://go.dev/) installed, along with git. Then clone this repo with 

```bash
git clone https://gitlab.com/RLN_/plutonium
```

Enter the directory it creates and type `go build`.

To check your Go version run:

```bash
go version
```

And ensure that it is at least version 1.18.

## Building FOCR

Note: Ensure you drop the built binary in the root Plutonium directory after building it.

For Linux or Mac OS ensure you have `git`, `g++`, and `make` installed. Clone FOCR via:

```bash
git clone https://gitlab.com/RLN_/focr-resurrection
```

followed by running `make` in the directory it creates.

For windows, you'll need to create a solution in Visual Studio, import the code, and then build focr.exe.

## Create accounts

Plutonium can automatically create accounts for you based on a text file. Run the bot with this command:

```bash
./bot -r /path/to/registration_file.txt
```

Where the registration file should have a new account for every line. For example:

```
myuser1:mypass:email@example.com
myuser2:mypass:email@example.com
myuser3:mypass:email@example.com
myuser4:mypass:email@example.com
myuser5:mypass:email@example.com
```

After creation you may want to run the `skip_tutorial.py` script on the accounts, followed by `get_bag.py` to get a sleeping bag.

## Memory

If you see your memory going up you're probably using walk_path_to from too large of a distance.

If you happen to notice that the VIRT memory allocated is high, it is due to [Go reserving a lot of virtual memory for allocations](https://go.dev/doc/faq#Why_does_my_Go_process_use_so_much_virtual_memory). Check the RES memory of htop to see the real memory footprint.

## Scripts

Documentation can be found [here](https://gitlab.com/RLN_/plutonium/-/blob/main/SCRIPTING.md), and API documentation can be found [here](https://gitlab.com/RLN_/plutonium/-/blob/main/API.md).