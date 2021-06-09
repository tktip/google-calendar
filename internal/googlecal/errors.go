package googlecal

import "fmt"

//revive:disable
type UserError error

var (
	ErrorMissingEventID      UserError = fmt.Errorf("missing event ID")
	ErrorMissingDates        UserError = fmt.Errorf("missing start or end date")
	ErrorMissingParticipants UserError = fmt.Errorf("no participants in event. must be at least one")
	ErrorBadLocation         UserError = fmt.Errorf("event has empty location")
	ErrorMissingTitle        UserError = fmt.Errorf("event has no title")
	ErrorUnknownDomain       UserError = fmt.Errorf("provided domain name unknown")
	ErrorBadID               UserError = fmt.Errorf("provided ID invalid, must be length 5 to 1024, and contain only lowercase letters and numbers 0-9")
)
