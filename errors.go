package array

import "time"
import "fmt"

// ArrPkgError encodes information related to errors that occur inside the package `array`.
type ArrPkgError struct {
	When      time.Time
	Where     string
	What      string
	TraceBack []error
}

func (e ArrPkgError) Error() string {
	return fmt.Sprintf("\n** error **\n[package: array]\n[when: %s]\n[where: %s]\n[what: %s]\n[traceback: \n%v\n]\n", e.When.String(), e.Where, e.What, e.TraceBack)
}

// MakeError makes an error specific to the `array` package.
func MakeError(where, what string, errors ...error) ArrPkgError {
	return ArrPkgError{When: time.Now().UTC(),
		Where:     where,
		What:      what,
		TraceBack: errors}
}
