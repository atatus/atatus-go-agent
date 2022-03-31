# atgormv2

Package atgormv2 provides drivers to gorm.io/gorm
for tracing database operations as spans of a transaction traced
by Atatus.

## Usage

Swap `gorm.io/driver/*` to `go.atatus.com/agent/module/atgormv2/driver/*`

Example :-

```golang
import (
    mysql "go.atatus.com/agent/module/atgormv2/driver/mysql"
    "gorm.io/gorm"
)	

db, err := gorm.Open(mysql.Open("dsn"), &gorm.Config{})
```