# atsql

Package atsql provides a wrapper for database/sql/driver.Drivers
for tracing database operations as spans of a transaction traced
by Atatus.

To instrument a driver, you can simply swap your application's
calls to [sql.Register](https://golang.org/pkg/database/sql/#Register)
and [sql.Open](https://golang.org/pkg/database/sql/#Open) to
atsql.Register and atsql.Open respectively. The atsql.Register
function accepts zero or more options to influence how tracing
is performed.
