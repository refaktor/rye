package evaldo

import (
	"time"

	"github.com/refaktor/rye/env"
)

var builtins_time = map[string]*env.Builtin{

	//
	// ##### Date and Time ##### ""
	//
	// Tests:
	// equal { now |unix-micro? |type? } 'integer
	// Args:
	// * time: Time object to convert
	// Returns:
	// * integer representing Unix time in microseconds
	"unix-micro?": {
		Argsn: 1,
		Doc:   "Converts a time object to Unix time in microseconds (microseconds since January 1, 1970 UTC).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Time:
				return *env.NewInteger(s1.Value.UnixMicro())
			default:
				return MakeArgError(ps, 1, []env.Type{env.TimeType}, "unix-micro?")
			}
		},
	},

	// Tests:
	// equal { now |unix-milli? |type? } 'integer
	// Args:
	// * time: Time object to convert
	// Returns:
	// * integer representing Unix time in milliseconds
	"unix-milli?": {
		Argsn: 1,
		Doc:   "Converts a time object to Unix time in milliseconds (milliseconds since January 1, 1970 UTC).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Time:
				return *env.NewInteger(s1.Value.UnixMilli())
			default:
				return MakeArgError(ps, 1, []env.Type{env.TimeType}, "unix-milli?")
			}
		},
	},

	// Tests:
	// equal { date "2023-01-01" |weekday? } 0  ; Sunday is 0
	// Args:
	// * time: Time object to extract weekday from
	// Returns:
	// * integer representing day of week (0=Sunday, 1=Monday, ..., 6=Saturday)
	"weekday?": {
		Argsn: 1,
		Doc:   "Extracts the day of the week from a time object as an integer (0=Sunday through 6=Saturday).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Time:
				return *env.NewInteger(int64(s1.Value.Weekday()))
			default:
				return MakeArgError(ps, 1, []env.Type{env.TimeType}, "weekday?")
			}
		},
	},

	// date time functions

	// Tests:
	// equal { date "2023-01-15" |type? } 'time
	// equal { date "2023-01-15" |day? } 15
	// Args:
	// * datestr: String in "YYYY-MM-DD" format
	// Returns:
	// * time object representing the specified date
	"date": {
		Argsn: 1,
		Doc:   "Creates a time object from a date string in YYYY-MM-DD format.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch s1 := arg0.(type) {
			case env.String:
				t, err := time.Parse("2006-01-02", s1.Value)
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "date")
				}
				return *env.NewTime(t)
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "date")
			}
		},
	},

	// Tests:
	// equal { datetime "2023-01-15T14:30:45" |type? } 'time
	// equal { datetime "2023-01-15T14:30:45" |hour? } 14
	// Args:
	// * datetimestr: String in "YYYY-MM-DDThh:mm:ss" format
	// Returns:
	// * time object representing the specified date and time
	"datetime": {
		Argsn: 1,
		Doc:   "Creates a time object from a datetime string in YYYY-MM-DDThh:mm:ss format.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) (res env.Object) {
			switch s1 := arg0.(type) {
			case env.String:
				t, err := time.Parse("2006-01-02T15:04:05", s1.Value)
				if err != nil {
					return MakeBuiltinError(ps, err.Error(), "datetime")
				}
				return *env.NewTime(t)
			default:
				return MakeArgError(ps, 1, []env.Type{env.StringType}, "datetime")
			}
		},
	},

	// Tests:
	// equal { now |type? } 'time
	// Args:
	// * none
	// Returns:
	// * time object representing the current time
	"now": {
		Argsn: 0,
		Doc:   "Creates a time object representing the current local time.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			return *env.NewTime(time.Now())
		},
	},

	// Tests:
	// equal { date "2023-01-15" |yearday? } 15  ; January 15 is the 15th day of the year
	// Args:
	// * time: Time object to extract day of year from
	// Returns:
	// * integer representing day of year (1-366)
	"yearday?": {
		Argsn: 1,
		Doc:   "Extracts the day of year from a time object as an integer (1-366).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Time:
				return *env.NewInteger(int64(s1.Value.YearDay()))
			default:
				return MakeArgError(ps, 1, []env.Type{env.TimeType}, "hour?")
			}
		},
	},

	// Tests:
	// equal { date "2023-01-15" |year? } 2023
	// Args:
	// * time: Time object to extract year from
	// Returns:
	// * integer representing the year
	"year?": {
		Argsn: 1,
		Doc:   "Extracts the year from a time object as an integer.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Time:
				return *env.NewInteger(int64(s1.Value.Year()))
			default:
				return MakeArgError(ps, 1, []env.Type{env.TimeType}, "year?")
			}
		},
	},

	// Tests:
	// equal { date "2023-01-15" |month? } 1
	// Args:
	// * time: Time object to extract month from
	// Returns:
	// * integer representing the month (1-12)
	"month?": {
		Argsn: 1,
		Doc:   "Extracts the month from a time object as an integer (1-12).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Time:
				return *env.NewInteger(int64(s1.Value.Month()))
			default:
				return MakeArgError(ps, 1, []env.Type{env.TimeType}, "month?")
			}
		},
	},

	// Tests:
	// equal { date "2023-01-15" |days-in-month? } 31
	// equal { date "2023-02-15" |days-in-month? } 28
	// Args:
	// * time: Time object to calculate days in month for
	// Returns:
	// * integer representing the number of days in the month
	"days-in-month?": {
		Argsn: 1,
		Doc:   "Calculates the number of days in the month of the given time object.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Time:
				days := time.Date(s1.Value.Year(), s1.Value.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
				return *env.NewInteger(int64(days))
			default:
				return MakeArgError(ps, 1, []env.Type{env.TimeType}, "days-in-month?")
			}
		},
	},

	// Tests:
	// equal { datetime "2023-01-15T14:30:45" |time? } "14:30:45"
	// Args:
	// * time: Time object to extract time from
	// Returns:
	// * string representing the time in "hh:mm:ss" format
	"time?": {
		Argsn: 1,
		Doc:   "Extracts the time part from a time object as a formatted string (hh:mm:ss).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Time:
				return *env.NewString(s1.Value.Format("15:04:05"))
			default:
				return MakeArgError(ps, 1, []env.Type{env.TimeType}, "time?")
			}
		},
	},

	// Tests:
	// equal { datetime "2023-01-15T14:30:45" |date? } "2023-01-15"
	// Args:
	// * time: Time object to extract date from
	// Returns:
	// * string representing the date in "YYYY-MM-DD" format
	"date?": {
		Argsn: 1,
		Doc:   "Extracts the date part from a time object as a formatted string (YYYY-MM-DD).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Time:
				return *env.NewString(s1.Value.Format("2006-01-02"))
			default:
				return MakeArgError(ps, 1, []env.Type{env.TimeType}, "date?")
			}
		},
	},

	// Tests:
	// equal { date "2023-01-15" |format-imap-date } "15-Jan-2023"
	// Args:
	// * time: Time object to format
	// Returns:
	// * string representing the date in IMAP search format (DD-Mon-YYYY)
	"format-imap-date": {
		Argsn: 1,
		Doc:   "Formats a time object as an IMAP search date string (DD-Mon-YYYY).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Time:
				return *env.NewString(s1.Value.Format("02-Jan-2006"))
			default:
				return MakeArgError(ps, 1, []env.Type{env.TimeType}, "form-imap-date")
			}
		},
	},

	// Tests:
	// equal { date "2023-01-15" |format-date "2006-01-02" } "2023-01-15"
	// equal { date "2023-01-15" |format-date "02/01/2006" } "15/01/2023"
	// equal { datetime "2023-01-15T14:30:45" |format-date "Mon Jan 2 15:04:05 2006" } "Sun Jan 15 14:30:45 2023"
	// Args:
	// * time: Time object to format
	// * layout: String - Go time format layout (reference time: Mon Jan 2 15:04:05 MST 2006)
	// Returns:
	// * string representing the formatted date/time
	"format-date": {
		Argsn: 2,
		Doc:   "Formats a time object using a Go time format layout string. Reference time: Mon Jan 2 15:04:05 MST 2006",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch t := arg0.(type) {
			case env.Time:
				switch layout := arg1.(type) {
				case env.String:
					return *env.NewString(t.Value.Format(layout.Value))
				default:
					return MakeArgError(ps, 2, []env.Type{env.StringType}, "format-date")
				}
			default:
				return MakeArgError(ps, 1, []env.Type{env.TimeType}, "format-date")
			}
		},
	},

	// Tests:
	// equal { date "2023-01-15" |day? } 15
	// Args:
	// * time: Time object to extract day from
	// Returns:
	// * integer representing the day of month (1-31)
	"day?": {
		Argsn: 1,
		Doc:   "Extracts the day of month from a time object as an integer (1-31).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Time:
				return *env.NewInteger(int64(s1.Value.Day()))
			default:
				return MakeArgError(ps, 1, []env.Type{env.TimeType}, "day?")
			}
		},
	},

	// Tests:
	// equal { datetime "2023-01-15T14:30:45" |hour? } 14
	// Args:
	// * time: Time object to extract hour from
	// Returns:
	// * integer representing the hour (0-23)
	"hour?": {
		Argsn: 1,
		Doc:   "Extracts the hour from a time object as an integer (0-23).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Time:
				return *env.NewInteger(int64(s1.Value.Hour()))
			default:
				return MakeArgError(ps, 1, []env.Type{env.TimeType}, "hour?")
			}
		},
	},

	// Tests:
	// equal { datetime "2023-01-15T14:30:45" |minute? } 30
	// Args:
	// * time: Time object to extract minute from
	// Returns:
	// * integer representing the minute (0-59)
	"minute?": {
		Argsn: 1,
		Doc:   "Extracts the minute from a time object as an integer (0-59).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Time:
				return *env.NewInteger(int64(s1.Value.Minute()))
			default:
				return MakeArgError(ps, 1, []env.Type{env.TimeType}, "minute?")
			}
		},
	},

	// Tests:
	// equal { datetime "2023-01-15T14:30:45" |second? } 45
	// Args:
	// * time: Time object to extract second from
	// Returns:
	// * integer representing the second (0-59)
	"second?": {
		Argsn: 1,
		Doc:   "Extracts the second from a time object as an integer (0-59).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch s1 := arg0.(type) {
			case env.Time:
				return *env.NewInteger(int64(s1.Value.Second()))
			default:
				return MakeArgError(ps, 1, []env.Type{env.TimeType}, "second?")
			}
		},
	},

	// end of date time functions

	// Tests:
	// equal { time-it { sleep 10 } } 10
	// Args:
	// * milliseconds: Integer number of milliseconds to sleep
	// Returns:
	// * the original milliseconds value
	"sleep": {
		Argsn: 1,
		Doc:   "Pauses execution for the specified number of milliseconds.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				time.Sleep(time.Duration(int(arg.Value)) * time.Millisecond)
				return arg
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "sleep")
			}
		},
	},

	// Tests:
	// equal { 5 .seconds } 5000
	// Args:
	// * n: Integer number of seconds
	// Returns:
	// * integer representing the equivalent number of milliseconds
	"thousands": {
		Argsn: 1,
		Doc:   "Converts seconds to milliseconds (multiplies by 1000).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				return *env.NewInteger(arg.Value * 1000)
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "sleep")
			}
		},
	},
	// Tests:
	// equal { 5 .seconds } 5000
	// Args:
	// * n: Integer number of seconds
	// Returns:
	// * integer representing the equivalent number of milliseconds
	"millions": {
		Argsn: 1,
		Doc:   "Converts seconds to milliseconds (multiplies by 1000).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				return *env.NewInteger(arg.Value * 1000000)
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "sleep")
			}
		},
	},
	// Tests:
	// equal { 5 .seconds } 5000
	// Args:
	// * n: Integer number of seconds
	// Returns:
	// * integer representing the equivalent number of milliseconds
	"seconds": {
		Argsn: 1,
		Doc:   "Converts seconds to milliseconds (multiplies by 1000).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				return *env.NewInteger(int64(time.Duration(arg.Value) * 1000))
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "sleep")
			}
		},
	},
	// Tests:
	// equal { 5 .minutes } 300000  ; 5000 * 60
	// Args:
	// * n: Integer number of minutes
	// Returns:
	// * integer representing the equivalent number of milliseconds
	"minutes": {
		Argsn: 1,
		Doc:   "Converts minutes to milliseconds (multiplies by 60000).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				return *env.NewInteger(int64(time.Duration(arg.Value) * 1000 * 60))
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "sleep")
			}
		},
	},

	// Tests:
	// equal { 5 .hours } 18000000  ; 5000 * 60 * 60
	// Args:
	// * n: Integer number of hours
	// Returns:
	// * integer representing the equivalent number of milliseconds
	"hours": {
		Argsn: 1,
		Doc:   "Converts hours to milliseconds (multiplies by 3600000).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				return *env.NewInteger(int64(time.Duration(arg.Value) * 1000 * 60 * 60))
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "sleep")
			}
		},
	},

	// Tests:
	// equal { 5 .days } 432000000  ; 5000 * 60 * 60 * 24
	// Args:
	// * n: Integer number of days
	// Returns:
	// * integer representing the equivalent number of milliseconds
	"days": {
		Argsn: 1,
		Doc:   "Converts days to milliseconds (multiplies by 86400000).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				return *env.NewInteger(int64(time.Duration(arg.Value) * 1000 * 60 * 60 * 24))
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "sleep")
			}
		},
	},
	// Tests:
	// equal { 5 .weeks } 3024000000  ; 5000 * 60 * 60 * 24 * 7
	// Args:
	// * n: Integer number of weeks
	// Returns:
	// * integer representing the equivalent number of milliseconds
	"weeks": {
		Argsn: 1,
		Doc:   "Converts weeks to milliseconds (multiplies by 604800000).",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch arg := arg0.(type) {
			case env.Integer:
				return *env.NewInteger(int64(time.Duration(arg.Value) * 1000 * 60 * 60 * 24 * 7))
			default:
				return MakeArgError(ps, 1, []env.Type{env.IntegerType}, "sleep")
			}
		},
	},

	// TODOC
	// Tests:
	// equal { time-it { sleep 100 } } 100
	"time-it": { // **
		Argsn: 1,
		Doc:   "Accepts a block, does it and times it's execution time.",
		Fn: func(ps *env.ProgramState, arg0 env.Object, arg1 env.Object, arg2 env.Object, arg3 env.Object, arg4 env.Object) env.Object {
			switch bloc := arg0.(type) {
			case env.Block:
				ser := ps.Ser
				ps.Ser = bloc.Series
				start := time.Now()
				EvalBlock(ps)
				MaybeDisplayFailureOrError(ps, ps.Idx, "time-it")
				if ps.ErrorFlag {
					ps.Ser = ser
					return ps.Res
				}
				t := time.Now()
				elapsed := t.Sub(start)
				ps.Ser = ser
				return *env.NewInteger(elapsed.Nanoseconds() / 1000000)
			default:
				return MakeArgError(ps, 1, []env.Type{env.BlockType}, "time-it")
			}
		},
	},
}
