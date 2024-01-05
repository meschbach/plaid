package local

import "fmt"

type labeledError struct {
	doing      string
	underlying error
}

func (l *labeledError) Error() string {
	return fmt.Sprintf("%s because %s", l.doing, l.underlying.Error())
}
