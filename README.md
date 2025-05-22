# goconfig

A library to gather parameters from various places into a single config struct

## What is supported and what is not

* Can:
  * Use any struct to store configs
  * Automatically build configuration argument names
* Cannot (yet):
  * Define maps or arrays with command line args or environment variables
  * Use recursive structs (when a struct has a child of the same type)
  * Operate multilevel pointers (pointer to pointer to ... to pointer to something)
* Cannot:
  * Operate ```chan``` fields
  * Operate "any" (```interface{}```) fields

## How to use it

1. Define a struct to store your config
2. If needed, customize field names with tags, ```env``` for environment variables and ```arg``` for the command line arguments
3. Call ```GetConfig[YourStruct](FromEnv, FromArg)``` to read config values

## How it works

1. Flatten: convert the type graph from a tree (parent<sub>1</sub> → [child<sub>1</sub>, ..., child<sub>n</sub>]) into a list of edges "[child<sub>1</sub> → parent<sub>1</sub>, ..., child<sub>n</sub> → parent<sub>1</sub>, child<sub>n + 1</sub> → parent<sub>m</sub>]"
2. Enumerate configurable fields (simple strings, integer and float numbers) and their names for command line arguments or env variables.
3. Set the field values from the value sources - if value is provided by multiple sources, then we will use the latest.
4. Fold edge list into a tree, create nodes which have values among their branches, other will remain nil.

## Examples
```go
import (
    "github.com/lanseg/goconfig"
)

type HttpSettings struct {
    Timeout int `arg:"timeout" env:"TIMEOUT"`
    Retries int `arg:"retries" env:"RETRIES"`
}

type StorageSettings struct {
    Root string `arg:"root" env:"ROOT"`
}

type Config struct {
    HttpSettings    `arg:"http" env:"HTTP"`
    StorageSettings `arg:"storage" env:"STORAGE"`
}

func main() {
    cfg, err := goconfig.GetConfig[goconfig.Config](goconfig.FromArgs, goconfig.FromEnv)
    if err != nil {
        fmt.Println(err)
        os.Exit(-1)
    }
r
    fmt.Println(cfg)
}
```
