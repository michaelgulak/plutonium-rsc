# Plutonium

Plutonium is a shell bot written for the [Open RSC](https://rsc.vet/) Uranium server than runs on Linux, Windows, MacOS, etc. Being a shell bot means that it runs without any graphics whatsoever, and thus has a much smaller memory and cpu footprint than other bots. Plutonium also makes it easy to run several accounts at once. One thing to take note of is that Open RSC servers have a max player per ip limit that they change sometimes. So the most bots per ip you can run on Uranium is around 10 at any given time.

Plutonium was built from the ground up using Go. Scripts for it are written in python via [gpython](https://github.com/go-python/gpython). Some code is taken from the Open RSC server (in particular, the pathfinding and chat message code).

## Building Plutonium

*Note: Plutonium no longer requires an OCR.*

Make sure you have a new version of [Go](https://go.dev/) installed, along with git. On Linux, open a terminal, or on Windows open a command prompt. Then clone this repo with 

```bash
git clone https://gitlab.com/open-runescape-classic/plutonium
```

Enter the directory it creates (type `cd plutonium` to get in the directory) and type `go build`. A binary named `bot` (or on Windows `bot.exe`) will now be in the directory.

To check your Go version run:

```bash
go version
```

And ensure that it is at least version 1.18.

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

## Run Plutonium

On Linux, open up a terminal, or on Windows, open up a command prompt. Note that for the rest of the README if a command for the bot is shown as `./bot`, then on Windows you would replace that with `.\bot.exe`.

On Linux, make sure your working directory is the bot directory and run:

```bash
./bot
``` 

If you want to run just one instance of a bot for any reason, for example to skip the tutorial you can do:

```bash
./bot -u myuser -p mypass -s scripts/skip_tutorial.py
```

You can't specify script settings with this method.

Alternatively, you can run a single account file like this:

```bash
./bot -f /path/to/account_file.toml
```

## Create accounts

Plutonium can automatically create accounts for you based on a text file. Note that there are rate limits for creating accounts. You may have to take a Very Pesky Note that it might need to be done over multiple days.

Run the bot with this command:

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

Each bot instance should use 10-20x less memory than an APOS client.

If you see your memory going up, you're probably using walk_path_to from too large of a distance. Consider using calculate_path.

If you happen to notice that the VIRT memory allocated is high, it is due to [Go reserving a lot of virtual memory for allocations](https://go.dev/doc/faq#Why_does_my_Go_process_use_so_much_virtual_memory). Check the RES memory of htop to see the real memory footprint.

## Scripts

Documentation can be found [here](https://gitlab.com/open-runescape-classic/plutonium/-/blob/main/SCRIPTING.md), and API documentation can be found [here](https://gitlab.com/open-runescape-classic/plutonium/-/blob/main/API.md).