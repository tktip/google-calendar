package swagex

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

// global swagger details
// @title FLYVO Google Calendar sync swagger
// @version 1.0
// @description Swagger Google calendar sync.
// @description accepts event data uploads it
// @description to google calendar.

// @host localhost:8080
// @BasePath /

var (
	swaggerDoc = []byte("{}")
)

func init() {
	swaggerLoc := os.Getenv("SWAGGER_LOCATION")
	if swaggerLoc == "" {
		swaggerLoc = "/swagger.json"
	}

	var err error
	swaggerDoc, err = ioutil.ReadFile(swaggerLoc)
	if err != nil || swaggerDoc == nil {
		log.Warn("Could not read swagger doc: " + err.Error())
		swaggerDoc = []byte("{}")
	}
}

//SwaggerEndpoint route details.
// @Summary Swagger doc.
// @Description Returns the swagger doc.
// @Produce application/json
// @Success 200 {string} string "The swagger json."
// @Success 404 {string} string "If swagger has not been added."
// @Router /api-doc [get]
func SwaggerEndpoint(c *gin.Context) {
	w := c.Writer

	w.Header().Set("content-type", "application/json; charset=UTF-8")
	w.Write(swaggerDoc)
}
