package protocol

import "time"

func QSLogin(s Session, user, pass string) error {
	s.Expect("login:")
	s.Send(user)
	s.Expect("Password:")
	s.Send(pass)
	return s.Err()
}

// #DEVICE,2,4,3<CR><LF>
// Integration ID, Component Number, Action Number
// See https://assets.lutron.com/a/documents/040249.pdf
func Device(s Session, id, component, action int) {
}

// #OUTPUT,1,1,75,01:30<CR><LF>
// Integration ID, Action Number, Level, Fade Time (seconds)
// See https://assets.lutron.com/a/documents/040249.pdf
func Output(s Session, id, component, level int, time time.Duration) {
}

// ?OUTPUT,3,1<CR><LF>
// Integration ID, Action
// See https://assets.lutron.com/a/documents/040249.pdf
func Query(s Session, id, action int) {
}

// ~OUTPUT,3,1,90.00<CR><LF>
// Integration ID, Action, Level
// See https://assets.lutron.com/a/documents/040249.pdf
func Monitor(s Session, id, action, level int) {
}
