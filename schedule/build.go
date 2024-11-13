package schedule

/*
type Event struct {
	Label  string
	When   time.Time
	Device string
	Action string
}

type Day struct {
	Events []Event
}

type Annual struct {
	Days []Day
	Leap bool
}


func NewCurrentYear() Annual {
	var cal Annual
	year := time.Now().Year()
	ndays := 365
	if year%4 == 0 && year%100 != 0 || year%400 == 0 {
		ndays = 366
		cal.Leap = true
	}
	cal.Days = make([]Day, ndays)
	return cal
}
*/
