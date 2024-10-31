# Plutonium Scripting

Scripts are written in python. Many methods are provided to scripts in order to control the bot.

## Script layout

The basis of scripting is the `loop` function. You will want to define a function like this:

```python
def loop():
    return 5000
```

Whatever you return from `loop` will be the wait time in milliseconds before `loop` gets called again.

## Script Settings

A `settings` object is available in the global scope which contains what was defined in the account toml file under the `[script.settings]` section. You won't be able to access the settings object in submodules outside of functions. It's accessible within functions.

For example with a config file like this:

```toml
[account]
user = "myuser"
pass = "mypass"
autologin = true
enabled = true
debug = true

[script]
name = "autofighter.py"

[script.settings]
fight_mode = 0
npc_ids = [29, 34]
```

The settings object would contain both `fight_mode` and `npc_ids`. They can be accessed like this inside a script:

```python
def loop():
    if get_combat_style() != settings.fight_mode:
        set_combat_style(settings.fight_mode)
        return 1000

    if in_combat():
        return 500

    if get_fatigue() >= 95:
        use_sleeping_bag()
        return 1000

    npc = get_nearest_npc_by_id(ids=settings.npc_ids, in_combat=False, reachable=True)
    if npc != None:
        attack_npc(npc)
        return 500
  
    return 500
```

Note the usage of the `settings` object.

## Progress reports

Within the `[script]` block of the `account.toml` file you are able to define a `progress_report` field with a duration. This duration is waited upon to call a function named `on_progress_report` that should return a string dict. For example with a config like this:

```toml
[script]
name = "seers_yews.py"
progress_report = "20m"
```

Would cause the bot to call the `on_progress_report` function every 20 minutes, generating a report in `logs/progress_reports`. An `on_progress_report` function may look like this:

```python
def on_progress_report():
    return {"Woodcutting Level": get_max_stat(8),
            "Logs Banked": logs_banked}
```

Which would generate a report like this:

```
+-------------------+-------+
|        KEY        | VALUE |
+-------------------+-------+
| Woodcutting Level |    81 |
| Logs Banked       |  6216 |
+-------------------+-------+
```

## API

Check out the API documentation [here](https://gitlab.com/openrsc/plutonium/-/blob/main/API.md).

## Limitations of gpython

Here I will document the limitations I know about gpython.

#### Dicts

Dicts can only have string keys. You can do `mydict[str(myint)] = myvalue` then access it the same way.