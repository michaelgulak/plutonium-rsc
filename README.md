# Plutonium

Plutonium is a lightweight shell bot written for the [OpenRSC](https://rsc.vet/) Uranium server that runs on at least Linux, Windows, and MacOS. Being a shell bot means that it runs without any graphics, and thus has a much smaller memory and cpu footprint than other bots. A Plutonium process running 10 accounts uses around 25-30mb memory. Plutonium also makes it easy to run several accounts at once with a single command, and doesn't have the inherent memory leaks of traditional RSC clients - meaning it can run for years on a stable box. 

Scripts for it are written in Python 3.4 via [gpython](https://github.com/go-python/gpython).

Consider that the OpenRSC servers have a max player per ip limit. At the time of this writing, the limit is 10 on Uranium. There is also a registration limit per day so you need to spend multiple days creating accounts in order to reach the maximum of 10.

Some code is ported from Java to Go from the Open RSC server (in particular, the pathfinding and chat message code).

## Building Plutonium

You must build Plutonium as there aren't any official builds of it. Make sure you have at least version 1.21 of [Go](https://go.dev/) installed, along with [git](https://git-scm.com/). On Linux, open a terminal, or on Windows open a command prompt. Then clone this repo with 

```bash
git clone https://gitlab.com/openrsc/plutonium
```

Enter the directory it creates (type `cd plutonium` to get in the directory) and type `go build`. A binary named `bot` (or on Windows `bot.exe`) will now be in the directory.

## Account Configuration

To run Plutonium, you need account files. Each account file specifies the username and password of an account, and the script to run for that account. An account config is created by creating a file in the `accounts` directory. These files must be in [toml](https://toml.io/en/) format.

To get started quickly, rename the example file in the `accounts` directory to be `myaccount.toml`. Then input your username and password, and make sure to change the script and script settings. Windows users should note that [file name extensions](https://answers.microsoft.com/en-us/windows/forum/all/how-can-i-get-the-extension-to-display-along-with/ec523f53-357b-41eb-a6c7-9b6b95a91235) must be enabled to rename the file properly.

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

When you set `progress_report` in an account file while the script you're running is designed to generate progress reports, progress reports will be written to `logs/progress_reports` every interval you set.

## Run Plutonium

On Linux, open up a terminal, or on Windows, open up a command prompt. Note that for the rest of the README if a command for the bot is shown as `./bot`, then on Windows you would replace that with `.\bot.exe`.

On Linux, make sure your working directory is the bot directory and run:

```bash
./bot
``` 

To close Plutonium, press Ctrl+C in the terminal/command prompt.

If you want to run just one instance of a bot for any reason, for example to skip the tutorial you can do:

```bash
./bot -u myuser -p mypass -s scripts/skip_tutorial.py
```

You can't specify script settings with this method. Use the `-l` flag in order to set autologin to true when running the bot this way.

Alternatively, you can run a single account file like this:

```bash
./bot -f /path/to/account_file.toml
```

After you create your account and run `skip_tutorial.py` on it, you may want to run `get_bag.py` to get a sleeping bag.

## Overwrite Progress Reports

If you want progress report files to be overwritten every time they're generated you can set `overwrite_progress_reports` to `true` in `settings.toml`.

## Updating Plutonium

To update Plutonium run these commands in the terminal or command prompt while your working directory is the `plutonium` directory (it will save your changes and replay them over the updated bot):

```bash
git stash
git pull
git stash pop
go build
```

This will only work if you cloned the repo first with git.

## Memory

Known issue: If you see your memory going up, your script is probably calling `calculate_path_to` without a `depth` or `max_depth`.

If you happen to notice that the VIRT memory allocated is high, it is due to [Go reserving a lot of virtual memory for allocations](https://go.dev/doc/faq#Why_does_my_Go_process_use_so_much_virtual_memory). Check the RES memory of htop to see the real memory footprint.

## Scripts

General scripting documentation/features can be found [here](https://gitlab.com/openrsc/plutonium/-/blob/main/SCRIPTING.md), and API documentation can be found [here](https://gitlab.com/openrsc/plutonium/-/blob/main/API.md).