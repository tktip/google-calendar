package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/tktip/google-calendar/internal/googlecal"
	global "github.com/tktip/google-calendar/pkg/googlecal"
)

type calendarQueryParams struct {
	BroadcastChanges *bool `form:"broadcastChanges"`
	GuestsCanModify  *bool `form:"guestsCanModify"`
	GuestsMayInvite  *bool `form:"guestsMayInvite"`
	GuestsVisible    *bool `form:"guestsVisible"`
	GuestsAutoAccept *bool `form:"guestsAutoAccept"`
	PrivateEvent     *bool `form:"privateEvent"`
}

func getQueryParams(c *gin.Context) (queryParams calendarQueryParams, ok bool) {
	err := c.BindQuery(&queryParams)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"id": "", "error": err.Error()})
		return queryParams, false
	}

	return queryParams, true
}

func createCalendarConnector(c *gin.Context) (connector *googlecal.CalendarConnector, ok bool) {

	queryParams, ok := getQueryParams(c)
	if !ok {
		return nil, false
	}

	return googlecal.NewCalendarConnector(c.Request.Context(), c.Param("domain")).
		InformGuestsAboutUpdates(queryParams.BroadcastChanges).
		GuestsCanModify(queryParams.GuestsCanModify).
		GuestsAutoAccept(queryParams.GuestsAutoAccept).
		EventIsprivate(queryParams.PrivateEvent).
		GuestsMayInviteOthers(queryParams.GuestsMayInvite).
		GuestsMaySeeOtherGuests(queryParams.GuestsVisible), true

}

// @Summary Create event in Google Calendar
// @Description Accepts an event and adds it to the google calendar.
// @Description Event changes by users are only local, not global.
// @Description Key values cannot be changed (i.e. date)
// @Description Returns the id of newly created event.
// @Produce json
// @Accept json
// @Param body body global.Event true "Event details"
// @Param domain path string true "Domain of event"
// @Param broadcastChanges query bool false "Whether to mail users about event"
// @Param guestsCanModify query bool false "Whether guests may modify the event"
// @Param guestsMayInvite query bool false "Whether guests may invite others"
// @Param guestsVisible query bool false "Whether guests are visible"
// @Success 200 "event was created and uploaded"
// @Failure 400 {string} string "If body missing or param is missing from body"
// @Failure 422 {string} string "On bad body"
// @Failure 500 {string} string "On unexpected error"
// @Router /event/{domain}/create [post]
func addEventToGoogle(c *gin.Context) {
	event := global.Event{}
	err := c.BindJSON(&event)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"id": "", "error": err.Error()})
		return
	}

	calendarConnector, ok := createCalendarConnector(c)
	if !ok {
		return
	}

	id, err := calendarConnector.CreateEvent(event)

	if err != nil {
		if _, ok := err.(googlecal.UserError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"id": "", "error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"id": "", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id, "error": nil})
}

// @Summary Delete event from google
// @Description Deletes event from google
// @Produce json
// @Param id path string true "ID of event to delete"
// @Param domain path string true "Domain of event"
// @Param broadcastChanges query bool false "Whether to mail users about event"
// @Failure 400 {string} string "If no ID provided"
// @Failure 500 {string} string "On unexpected error"
// @Router /event/{domain}/delete [delete]
func deleteEvent(c *gin.Context) {
	calendarConnector, ok := createCalendarConnector(c)
	if !ok {
		return
	}

	err := calendarConnector.DeleteEvent(c.Param("id"))

	if err != nil {
		if _, ok := err.(googlecal.UserError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"deleted": false, "error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"deleted": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true, "error": nil})
}

// @Summary Update google event
// @Description Update google event, overwrite preexisting values
// @Produce json
// @Accept json
// @Param body body global.Event true "New event details to set"
// @Param id path string true "ID of event to update"
// @Param domain path string true "Domain of event"
// @Param broadcastChanges query bool false "Whether to mail users about event"
// @Param guestsCanModify query bool false "Whether guests may modify the event"
// @Param guestsMayInvite query bool false "Whether guests may invite others"
// @Param guestsVisible query bool false "Whether guests are visible"
// @Success 200 {string} string "On successful update"
// @Failure 400 {string} string "If ID is missing, or query params bad"
// @Failure 422 {string} string "If body is missing or bad"
// @Failure 500 {string} string "On unexpected error"
// @Router /event/{domain}/update [patch]
func updateEvent(c *gin.Context) {

	event := global.Event{}
	err := c.BindJSON(&event)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"success": false, "error": err.Error()})
		return
	}

	calendarConnector, ok := createCalendarConnector(c)
	if !ok {
		return
	}

	err = calendarConnector.UpdateEvent(event)

	if err != nil {
		if _, ok := err.(googlecal.UserError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "error": nil})
}

// @Summary Patch google event
// @Description Patch google event, adding or replacing specified.
// @Produce json
// @Accept json
// @Param body body global.Event true "New event details to set"
// @Param id path string true "ID of event to update"
// @Param domain path string true "Domain of event"
// @Param broadcastChanges query bool false "Whether to mail users about event"
// @Param guestsCanModify query bool false "Whether guests may modify the event"
// @Param guestsMayInvite query bool false "Whether guests may invite others"
// @Param guestsVisible query bool false "Whether guests are visible"
// @Success 200 {string} string "If successfully patched"
// @Failure 400 {string} string "If ID is missing, or query params bad"
// @Failure 422 {string} string "If body is missing or bad"
// @Failure 500 {string} string "On unexpected error"
// @Router /event/{domain}/update [patch]
func patchEvent(c *gin.Context) {

	event := global.Event{}
	err := c.BindJSON(&event)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"id": "", "error": err.Error()})
		return
	}

	calendarConnector, ok := createCalendarConnector(c)
	if !ok {
		return
	}

	err = calendarConnector.PatchEvent(event)

	if err != nil {
		if _, ok := err.(googlecal.UserError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "error": nil})
}

// @Summary Remove event participants
// @Description Patch google event, removing specified participants
// @Produce json
// @Param eventId path string true "ID of event to update"
// @Param participants path string true "list of participants to remove (emails), comma separated"
// @Param domain path string true "Domain of event"
// @Param broadcastChanges query bool false "Whether to mail users about change"
// @Success 200 {string} string "On successfully removed"
// @Failure 400 {string} string "If ID is missing"
// @Failure 500 {string} string "On unexpected error"
// @Router /event/{domain}/participants/{eventId}/{participants} [delete]
func removeParticipants(c *gin.Context) {

	calendarConnector, ok := createCalendarConnector(c)
	if !ok {
		return
	}
	err := calendarConnector.RemoveParticipants(c.Param("eventId"),
		strings.Split(c.Param("participants"), ","),
	)

	if err != nil {
		if _, ok := err.(googlecal.UserError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "error": nil})
}

// @Summary Add event participants
// @Description Patch google event, adding specified participants
// @Description (can also use event/{domain}/patch)
// @Produce json
// @Param eventId path string true "ID of event to update"
// @Param participants path string true "list of participants to add (emails), comma separated"
// @Param domain path string true "Domain of event"
// @Param broadcastChanges query bool false "Whether to mail users about change"
// @Success 200 {string} string "On successfully added"
// @Failure 400 {string} string "If ID is missing"
// @Failure 500 {string} string "On unexpected error"
// @Router /event/{domain}/participants/{eventId}/{participants} [delete]
func addParticipants(c *gin.Context) {

	calendarConnector, ok := createCalendarConnector(c)
	if !ok {
		return
	}

	err := calendarConnector.AddParticipants(c.Param("eventId"),
		strings.Split(c.Param("participants"), ","),
	)

	if err != nil {
		if _, ok := err.(googlecal.UserError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "error": nil})
}

// @Summary Retrieve google event
// @Description Retrieve google event
// @Produce json
// @Param id path string true "ID of event to get"
// @Param domain path string true "Domain of event"
// @Failure 400 {string} string "If body or ID is missing"
// @Failure 500 {string} string "On unexpected error"
// @Router /event/{domain}/get [GET]
// @Success 200 {object} T "The google calendar event"
// \@Success 200 {object} calendar.Event "The google calendar event"
// \Have created an issue on github.
func getEvent(c *gin.Context) {

	event, err := googlecal.NewCalendarConnector(c.Request.Context(), c.Param("domain")).
		GetCalendarEvent(c.Param("id"))
	if err != nil {
		if _, ok := err.(googlecal.UserError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"event": nil, "error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"event": nil, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"event": event, "error": nil})
}

// @Summary Retrieve google event list
// @Description Retrieve google event list
// @Produce json
// @Param domain path string true "Domain of event"
// @Param startTimeMin path string true "Lower bounds for start time of events"
// @Param endTimeMax path string true "Upper bounds for end time of events"
// @Failure 400 {string} string "On missing param"
// @Failure 500 {string} string "On unexpected error"
// @Router /event/{domain}/list/:startTimeMin/:endTimeMax [GET]
// @Success 200 {object} T "The google calendar event"
func getEvents(c *gin.Context) {

	b, _ := strconv.ParseBool(c.Query("showDeleted")) //default: false

	//Return events between min (start time) and max (end time)
	//No min or max means include everything
	events, err := googlecal.
		NewCalendarConnector(c.Request.Context(), c.Param("domain")).
		GetEvents(c.Param("startTimeMin"), c.Param("endTimeMax"), b)

	if err != nil {
		if _, ok := err.(googlecal.UserError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"events": nil, "error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"events": nil, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"events": events, "error": nil})
}

//ListenAndServe starts the api
func ListenAndServe() {
	r := gin.New()
	r.Use(gin.Logger()) // request logging

	r.POST("/:domain/event/create", addEventToGoogle)
	r.DELETE("/:domain/event/delete/:id", deleteEvent)
	r.DELETE("/:domain/event/participants/:eventId/:participants", removeParticipants)
	r.POST("/:domain/event/participants/:eventId/:participants", addParticipants)
	r.PATCH("/:domain/event/patch", patchEvent)
	r.PUT("/:domain/event/update", updateEvent)
	r.GET("/:domain/event/get/:id", getEvent)
	//r.GET("/api-doc", swagex.SwaggerEndpoint)
	r.GET("/:domain/event/list/:startTimeMin/:endTimeMax", getEvents)

	logrus.Infof("Ready to serve! Listening on port 5555")
	http.ListenAndServe(":5555", r)
}
