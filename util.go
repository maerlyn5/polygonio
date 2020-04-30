package polygonio

import "time"

var AmericaNewYork *time.Location = nil

//used for debugging
const StringFormat = "2006-01-02 3:04:05 PM MST"

func init() {
	anyc, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}
	AmericaNewYork = anyc

}

func FromDate(t time.Time) time.Time {
	t = t.In(AmericaNewYork)
	if t.Hour() < 1 {
		return t.AddDate(0, 0, -1)
	}

	return t
}

func ToDate(t time.Time) time.Time {
	t = t.In(AmericaNewYork)
	return t.AddDate(0, 0, 1)
}

func TimespanAsDuration(in string) time.Duration {
	switch in {
	case "minute":
		return time.Minute
	case "hour":
		return time.Hour
	case "day": //aproximate
		return time.Hour * 24
	case "month": //aproximate
		return time.Hour * 730
	}
	panic("unknown timespan")
}
