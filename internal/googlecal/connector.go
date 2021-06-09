package googlecal

//revive:disable:cyclomatic

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	global "github.com/tktip/google-calendar/pkg/googlecal"
	"golang.org/x/net/context"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

var (
	configs CalendarConfig
)

func init() {
	b, err := ioutil.ReadFile(os.Getenv("CREDENTIALS"))
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	configs, err = ParseConfig(b)
	if err != nil {
		log.Fatalf("Unable to parse client secret file as config: %v", err)
	}
}

//CalendarConnector - base struct of dsl.
type CalendarConnector struct {
	domain  DomainName
	context context.Context

	//Use pointers to allow patch semantics
	guestsCanModify          *bool
	autoAccept               *bool
	guestsCanInvite          *bool
	guestsCanSeeGuests       *bool
	privateEvent             *bool
	informGuestsAboutUpdates *bool
}

//NewCalendarConnector - create calendar connector
func NewCalendarConnector(ctx context.Context, domain string) *CalendarConnector {
	eC := CalendarConnector{
		domain:  DomainName(domain),
		context: ctx,
	}
	return &eC
}

func (e *CalendarConnector) getCalendarService() (*calendar.Service, error) {
	config := configs[e.domain]
	if config == nil {
		return nil, ErrorUnknownDomain
	}

	if e.context == nil {
		panic("no context provided")
	}

	srv, err := calendar.NewService(e.context,
		option.WithTokenSource(
			config.TokenSource(e.context),
		),
	)
	return srv, err
}

//GuestsCanModify - whether guests can alter event
func (e *CalendarConnector) GuestsCanModify(b *bool) *CalendarConnector {
	e.guestsCanModify = b
	return e
}

//GuestsAutoAccept - whether event is auto accepted by guests
func (e *CalendarConnector) GuestsAutoAccept(b *bool) *CalendarConnector {
	e.autoAccept = b
	return e
}

//EventIsprivate - whether event is private or not
func (e *CalendarConnector) EventIsprivate(b *bool) *CalendarConnector {
	e.privateEvent = b
	return e
}

//GuestsMayInviteOthers - whether guests may invite others
func (e *CalendarConnector) GuestsMayInviteOthers(b *bool) *CalendarConnector {
	e.guestsCanInvite = b
	return e
}

//GuestsMaySeeOtherGuests - whether guests may see other guests
func (e *CalendarConnector) GuestsMaySeeOtherGuests(b *bool) *CalendarConnector {
	e.guestsCanSeeGuests = b
	return e
}

//InformGuestsAboutUpdates - whether mail should be sent on event change
func (e *CalendarConnector) InformGuestsAboutUpdates(b *bool) *CalendarConnector {
	e.informGuestsAboutUpdates = b
	return e
}

func isValidEventID(ID string) bool {
	if len(ID) < 5 || len(ID) > 1024 {
		return false
	}

	matched, err := regexp.MatchString("^[a-z0-9]*$", ID)
	if err != nil {
		panic(err.Error())
	}
	return matched
}

//isNewEventValid - checks if event contains mandatory fields
func isNewEventValid(event global.Event) error {
	if event.Start == nil || *event.Start == "" || event.End == nil || *event.End == "" {
		return ErrorMissingDates
	} else if event.Title == nil || *event.Title == "" {
		return ErrorMissingTitle
	}
	return nil
}

//common functionality for event create, update & patch
func (e *CalendarConnector) copyGoogleEventUpdate(event global.Event, update *calendar.Event) {
	if update == nil {
		return
	}

	if e.guestsCanModify != nil {
		update.GuestsCanModify = *e.guestsCanModify
	}

	if e.privateEvent != nil {
		visibility := "default"
		if *e.privateEvent {
			visibility = "private"
		}
		update.Visibility = visibility
	}

	if e.guestsCanInvite != nil {
		update.GuestsCanInviteOthers = e.guestsCanInvite
	}

	if e.guestsCanSeeGuests != nil {
		update.GuestsCanSeeOtherGuests = e.guestsCanSeeGuests
	}

	if event.Title != nil && *event.Title != "" {
		update.Summary = *event.Title
	}

	if event.Location != nil {
		update.Location = *event.Location
	}

	if event.Description != nil {
		update.Description = *event.Description
	}

	if event.Start != nil {
		update.Start = &calendar.EventDateTime{
			DateTime: *event.Start,
			TimeZone: "Europe/Oslo",
		}
	}

	if event.End != nil {
		update.End = &calendar.EventDateTime{
			DateTime: *event.End,
			TimeZone: "Europe/Oslo",
		}
	}

	respStatus := "needsAction"
	if e.autoAccept != nil && *e.autoAccept {
		respStatus = "accepted"
	}
	fmt.Println("RespStatus: " + respStatus)
	if event.Participants != nil {
		participants := []*calendar.EventAttendee{}
		for _, participant := range *event.Participants {
			participants = append(participants, &calendar.EventAttendee{
				Email:          participant,
				ResponseStatus: respStatus,
			})
		}
		update.Attendees = participants
	}

	if event.Organizer != nil {
		update.Organizer = event.Organizer
	}
}

//CreateEvent creates and uploads an event in Google Calendar
//based on contents of a global.Event struct
func (e *CalendarConnector) CreateEvent(event global.Event) (eventID string, err error) {
	err = isNewEventValid(event)
	if err != nil {
		return "", err
	}

	if event.ID != nil && !isValidEventID(*event.ID) {
		return "", ErrorBadID
	}

	srv, err := e.getCalendarService()
	if err != nil {
		return "", err
	}

	gEvent := calendar.Event{}
	if event.ID != nil {
		gEvent.Id = *event.ID
	}

	e.copyGoogleEventUpdate(event, &gEvent)

	insert := srv.Events.Insert("primary", &gEvent)
	if e.informGuestsAboutUpdates != nil && *e.informGuestsAboutUpdates {
		insert = insert.SendUpdates("all")
	}

	_event, err := insert.Do()
	if err != nil {
		return "", err
	}
	return _event.Id, nil
}

//DeleteEvent deletes event with ID
func (e *CalendarConnector) DeleteEvent(eventID string) error {
	if eventID == "" {
		return ErrorMissingEventID
	}
	config := configs[e.domain]
	if config == nil {
		return ErrorUnknownDomain
	}

	srv, err := e.getCalendarService()
	if err != nil {
		return err
	}

	delete := srv.Events.Delete("primary", eventID)
	if e.informGuestsAboutUpdates != nil && *e.informGuestsAboutUpdates {
		delete = delete.SendUpdates("all")
	}

	return delete.Do()
}

//PatchEvent updates an existing event using patch semantics
//Note: Will not replace entire event, just specified fields.
func (e *CalendarConnector) PatchEvent(event global.Event) error {
	if event.ID == nil || *event.ID == "" {
		return ErrorMissingEventID
	}

	srv, err := e.getCalendarService()
	if err != nil {
		return err
	}

	gEvent := calendar.Event{}
	e.copyGoogleEventUpdate(event, &gEvent)

	patch := srv.Events.Patch("primary", *event.ID, &gEvent)
	if e.informGuestsAboutUpdates != nil && *e.informGuestsAboutUpdates {
		patch = patch.SendUpdates("all")
	}

	_, err = patch.Do()
	return err
}

//UpdateEvent updates an existing event (overwrite)
//Note: Will overwrite any existing fields, as entire event object is replaced.
func (e *CalendarConnector) UpdateEvent(event global.Event) error {
	if event.ID == nil || *event.ID == "" {
		return ErrorMissingEventID
	}

	err := isNewEventValid(event)
	if err != nil {
		return err
	}

	srv, err := e.getCalendarService()
	if err != nil {
		return err
	}

	gEvent := calendar.Event{}
	e.copyGoogleEventUpdate(event, &gEvent)

	update := srv.Events.Update("primary", *event.ID, &gEvent)
	if e.informGuestsAboutUpdates != nil && *e.informGuestsAboutUpdates {
		update = update.SendUpdates("all")
	}

	_, err = update.Do()
	return err
}

//RemoveParticipants removes specified participants from an event
func (e *CalendarConnector) RemoveParticipants(eventID string, toRemove []string) error {
	if eventID == "" {
		return ErrorMissingEventID
	}

	srv, err := e.getCalendarService()
	if err != nil {
		return err
	}

	emailsToIgnore := map[string]bool{}
	for _, v := range toRemove {
		emailsToIgnore[v] = true
	}

	existingEvent, err := e.GetCalendarEvent(eventID)
	if err != nil {
		return err
	}

	participants := []*calendar.EventAttendee{}
	for _, v := range existingEvent.Attendees {
		if !emailsToIgnore[v.Email] {
			participants = append(participants, v)
		}
	}

	patchEvent := calendar.Event{
		Attendees: participants,
	}
	if len(participants) == 0 { //overwrite on empty, to delete participant list
		existingEvent.Attendees = participants
		update := srv.Events.Update("primary", eventID, existingEvent)
		if e.informGuestsAboutUpdates != nil && *e.informGuestsAboutUpdates {
			update = update.SendUpdates("all")
		}
		_, err = update.Do()
	} else {
		patch := srv.Events.Patch("primary", eventID, &patchEvent)
		if e.informGuestsAboutUpdates != nil && *e.informGuestsAboutUpdates {
			patch = patch.SendUpdates("all")
		}
		_, err = patch.Do()
	}
	return err
}

//AddParticipants adds users to an existing events participant list
func (e *CalendarConnector) AddParticipants(eventID string, toAdd []string) error {
	if eventID == "" {
		return ErrorMissingEventID
	}

	srv, err := e.getCalendarService()
	if err != nil {
		return err
	}

	existingEvent, err := e.GetCalendarEvent(eventID)
	if err != nil {
		return err
	}

	existingUsersMap := map[string]bool{}
	for _, v := range existingEvent.Attendees {
		existingUsersMap[v.Email] = true
	}

	for _, user := range toAdd {
		if !existingUsersMap[user] {
			existingEvent.Attendees = append(
				existingEvent.Attendees,
				&calendar.EventAttendee{Email: user},
			)
		}
	}

	patchEvent := &calendar.Event{
		Attendees: existingEvent.Attendees,
	}

	patch := srv.Events.Patch("primary", eventID, patchEvent)
	if e.informGuestsAboutUpdates != nil && *e.informGuestsAboutUpdates {
		patch = patch.SendUpdates("all")
	}

	_, err = patch.Do()
	return err
}

//GetCalendarEvent returns event by id. Returns calendar.Event type event.
func (e *CalendarConnector) GetCalendarEvent(ID string) (*calendar.Event, error) {
	config := configs[e.domain]
	if config == nil {
		return nil, ErrorUnknownDomain
	}
	srv, err := e.getCalendarService()
	if err != nil {
		return nil, err
	}

	get := srv.Events.Get("primary", ID)
	return get.Do()
}

//GetEvents returns event by id. Returns calendar.Event type event.
func (e *CalendarConnector) GetEvents(min string, max string, showDeleted bool) (
	*calendar.Events,
	error,
) {
	config := configs[e.domain]
	if config == nil {
		return nil, ErrorUnknownDomain
	}
	srv, err := e.getCalendarService()
	if err != nil {
		return nil, err
	}

	list := srv.Events.List("primary")

	list.SingleEvents(true)
	list.ShowDeleted(showDeleted)
	if min != "" {
		list.TimeMin(min)
	}

	if max != "" {
		list.TimeMax(max)
	}
	return list.Do()
}
