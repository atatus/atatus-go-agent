# Atatus Go Agent

Atatus Go Agent provides the Atatus Go SDK for monitoring your Go applications with Atatus. Using this agent you can monitoring your transactions, get an overview of your database calls and their queries, find the external requests that have an impact on your applications and overall performance of your application.

Since Go is a statically compiled language, you need to manually add Atatus to monitor your source code. You can refer our documentation here for the extended packages that we support for monitoring standard frameworks and packages.

## Installation

### Compatibility and Requirements

For the latest version of the agent, Go 1.8+ is required, due to the use of `context.Context`.

Linux, Windows and MacOS are supported.


### Installing and using the Go agent

The Atatus Go agent is a Go library. You can install the agent as follows:


```bash
go get -u go.atatus.com/agent
```

The following integration packages extend the base Atatus package to support the following frameworks and libraries.

* [echo](#echo)
* [gin](#gin)
* [sql](#sql)
* [gopg](#gopg)
* [gorm](#gorm)
* [gocql](#gocql)
* [redigo](#redigo)
* [goredis](#goredis)
* [goredisv8](#goredisv8)
* [elasticsearch](#elasticsearch)
* [mongo](#mongo)

#### Echo

    import (
        echo "github.com/labstack/echo/v4"

        "go.atatus.com/agent/module/atechov4"
    )

    func main() {
        e := echo.New()
        e.Use(atechov4.Middleware())
        ...
    }

#### Gin

    import (
        "go.atatus.com/agent/module/atgin"
    )

    func main() {
        engine := gin.New()
        engine.Use(atgin.Middleware(engine))
        ...
    }

#### SQL

    import (
        "go.atatus.com/agent/module/atsql"
        _ "go.atatus.com/agent/module/atsql/pq"
        _ "go.atatus.com/agent/module/atsql/sqlite3"
    )

    func main() {
        db, err := atsql.Open("postgres", "postgres://...")
        db, err := atsql.Open("sqlite3", ":memory:")
    }

#### Postgres

    import (
        "github.com/go-pg/pg"

        "go.atatus.com/agent/module/atgopg"
    )

    func main() {
        db := pg.Connect(&pg.Options{})
        atgopg.Instrument(db)

        db.WithContext(ctx).Model(...)
    }

#### Postgres v10

    import (
        "github.com/go-pg/pg/v10"

        "go.atatus.com/agent/module/atgopgv10"
    )

    func main() {
        db := pg.Connect(&pg.Options{})
        atgopg.Instrument(db)

        db.WithContext(ctx).Model(...)
    }

#### GORM

    import (
        "go.atatus.com/agent/module/atgorm"
        _ "go.atatus.com/agent/module/atgorm/dialects/postgres"
    )

    func main() {
        db, err := atgorm.Open("postgres", "")
        ...
        db = atgorm.WithContext(ctx, db)
        db.Find(...) // creates a "SELECT FROM <foo>" span
    }

#### GORM v2

    import (
        "gorm.io/gorm"
        postgres "go.atatus.com/agent/module/atgormv2/driver/postgres"
    )

    func main() {
        db, err := gorm.Open(postgres.Open("dsn"), &gorm.Config{})
        ...
        db.WithContext(ctx).Find(...) // creates a "SELECT FROM <foo>" span
    }

#### GoCQL

    import (
        "github.com/gocql/gocql"

        "go.atatus.com/agent/module/atgocql"
    )

    func main() {
        observer := atgocql.NewObserver()
        config := gocql.NewCluster("cassandra_host")
        config.QueryObserver = observer
        config.BatchObserver = observer

        session, err := config.CreateSession()
        ...
        err = session.Query("SELECT * FROM foo").WithContext(ctx).Exec()
        ...
    }

#### Redigo

    import (
        "net/http"

        "github.com/gomodule/redigo/redis"

        "go.atatus.com/agent/module/atredigo"
    )

    var redisPool *redis.Pool // initialized at program startup

    func handleRequest(w http.ResponseWriter, req *http.Request) {
        // Wrap and bind redis.Conn to request context.
        conn := atredigo.Wrap(redisPool.Get()).WithContext(req.Context())
        defer conn.Close()
        ...
    }

#### Go Redis

    import (
        "net/http"

        "github.com/go-redis/redis"

        "go.atatus.com/agent/module/atgoredis"
    )

    var redisClient *redis.Client // initialized at program startup

    func handleRequest(w http.ResponseWriter, req *http.Request) {
        // Wrap and bind redisClient to the request context.
        client := atgoredis.Wrap(redisClient).WithContext(req.Context())
        ...
    }

#### Go Redis v8

    import (
        "github.com/go-redis/redis/v8"

        atgoredis "go.atatus.com/agent/module/atgoredisv8"
    )

    func main() {
        redisClient := redis.NewClient(&redis.Options{})
        // Add apm hook to redisClient.
        redisClient.AddHook(atgoredis.NewHook())

        redisClient.Get(ctx, "key")
    }

#### Elasticsearch


    import (
        "net/http"

        "github.com/olivere/elastic"

        "go.atatus.com/agent/module/atelasticsearch"
    )

    var client, _ = elastic.NewClient(elastic.SetHttpClient(&http.Client{
        Transport: atelasticsearch.WrapRoundTripper(http.DefaultTransport),
    }))

    func handleRequest(w http.ResponseWriter, req *http.Request) {
        result, err := client.Search("index").Query(elastic.NewMatchAllQuery()).Do(req.Context())
        ...
    }

#### Mongo

    import (
        "context"
        "net/http"

        "go.mongodb.org/mongo-driver/bson"
        "go.mongodb.org/mongo-driver/mongo"
        "go.mongodb.org/mongo-driver/mongo/options"

        "go.atatus.com/agent/module/atmongo"
    )

    var client, _ = mongo.Connect(
        context.Background(),
        options.Client().SetMonitor(atmongo.CommandMonitor()),
    )

    func handleRequest(w http.ResponseWriter, req *http.Request) {
        collection := client.Database("db").Collection("coll")
        cur, err := collection.Find(req.Context(), bson.D{})
        ...
    }

## License

The Atatus Go agent is licensed under the [Apache 2.0](http://apache.org/licenses/LICENSE-2.0.txt) License.

