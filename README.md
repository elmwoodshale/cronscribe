# cronscribe

A small Go library for working with standard five-field cron expressions:
validate one, find out exactly what's wrong with it, and turn a parsed
schedule back into a clean, canonical string.

Cron syntax looks simple until you have to validate it. `*/15`, `1-5,9`,
`MON-FRI`, and `7` as an alias for Sunday are all legal, and most hand-rolled
parsers either accept garbage silently or reject valid input. cronscribe
parses the expression into a plain struct of int slices, so the rest of your
program can reason about "what minutes does this run on" without re-parsing
strings.

## Usage

```go
package main

import (
	"fmt"

	"github.com/elmwoodshale/cronscribe"
)

func main() {
	s, err := cronscribe.Parse("*/15 9-17 * * MON-FRI")
	if err != nil {
		fmt.Println("invalid cron expression:", err)
		return
	}

	fmt.Println(s.Minute)    // [0 15 30 45]
	fmt.Println(s.DayOfWeek) // [1 2 3 4 5]
	fmt.Println(s.String())  // 0,15,30,45 9-17 * * 1-5

	daily, _ := cronscribe.Parse("0 9 * * MON-FRI")
	fmt.Println(daily.Describe())
	// at 09:00, on Monday, Tuesday, Wednesday, Thursday and Friday
}
```

A malformed expression comes back as an error that names the field and the
reason:

```go
_, err := cronscribe.Parse("* 25 * * *")
// cron: hour field "25": value 25 out of range [0,23]
```

## Supported syntax

- `*` for "every value"
- comma-separated lists: `1,2,3`
- ranges: `1-5`
- steps: `*/15`, `1-10/2`
- three-letter names for month (`JAN`-`DEC`) and day of week (`SUN`-`SAT`)
- `7` as an alias for Sunday in the day-of-week field

Not supported yet: `@daily`-style shorthands, a seconds field, and
descending ranges that wrap around (`22-2`). Field values outside their
valid range, or with a start greater than their end, are rejected.

`Schedule.Describe()` turns a parsed schedule into an English sentence, for
places like a UI or a log line where "0 9 * * MON-FRI" isn't self-explanatory.
It recognizes common shapes - a fixed time of day, an hourly interval, a
fixed number of minutes - and falls back to listing the raw field values for
anything unusual.

## Design

Every exported function is pure: `Parse` takes a string and returns a
`Schedule` or an error, nothing else; `Schedule.String()` and
`Schedule.Describe()` only read their receiver and return a string. There's
no shared state and no I/O, which makes all three trivial to table-test and
safe to call from anywhere, including concurrently.

## Status

Early. The parser and formatter cover the field syntax described above.
See the repository's open work for what's planned next.

## License

MIT, see [LICENSE](LICENSE).
