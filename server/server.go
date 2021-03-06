package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/takashabe/go-pubsub/datastore"
	"github.com/takashabe/go-pubsub/models"
	"github.com/takashabe/go-pubsub/stats"
	"github.com/takashabe/go-router"
)

// PrintDebugf behaves like log.Printf only in the debug env
func PrintDebugf(format string, args ...interface{}) {
	if env := os.Getenv("GO_PUBSUB_DEBUG"); len(env) != 0 {
		log.Printf("[DEBUG] "+format+"\n", args...)
	}
}

// ErrorResponse is Error response template
type ErrorResponse struct {
	Message string `json:"reason"`
	Error   error  `json:"-"`
}

func (e *ErrorResponse) String() string {
	return fmt.Sprintf("reason: %s, error: %v", e.Message, e.Error)
}

// Respond is response write to ResponseWriter
func Respond(w http.ResponseWriter, code int, src interface{}) {
	var body []byte
	var err error

	switch s := src.(type) {
	case []byte:
		if !json.Valid(s) {
			Error(w, http.StatusInternalServerError, err, "invalid json")
			return
		}
		body = s
	case string:
		body = []byte(s)
	case *ErrorResponse, ErrorResponse:
		// avoid infinite loop
		if body, err = json.Marshal(src); err != nil {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("{\"reason\":\"failed to parse json\"}"))
			return
		}
	default:
		if body, err = json.Marshal(src); err != nil {
			Error(w, http.StatusInternalServerError, err, "failed to parse json")
			return
		}
	}
	w.WriteHeader(code)
	w.Write(body)
}

// Error is wrapped Respond when error response
func Error(w http.ResponseWriter, code int, err error, msg string) {
	e := &ErrorResponse{
		Message: msg,
		Error:   err,
	}
	PrintDebugf("%v", e)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	Respond(w, code, e)
}

// JSON is wrapped Respond when success response
func JSON(w http.ResponseWriter, code int, src interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	Respond(w, code, src)
}

// Routes returns initialized for the topic and subscription router
func Routes() *router.Router {
	r := router.NewRouter()

	ts := TopicServer{}
	topicRoot := "/topic"
	r.Get(topicRoot+"/", ts.List)
	r.Get(topicRoot+"/:id", ts.Get)
	r.Get(topicRoot+"/:id/subscriptions", ts.ListSubscription)
	r.Put(topicRoot+"/:id", ts.Create)
	r.Post(topicRoot+"/:id/publish", ts.Publish)
	r.Delete(topicRoot+"/:id", ts.Delete)

	ss := SubscriptionServer{}
	subscriptionRoot := "/subscription"
	r.Get(subscriptionRoot+"/", ss.List)
	r.Get(subscriptionRoot+"/:id", ss.Get)
	r.Put(subscriptionRoot+"/:id", ss.Create)
	r.Post(subscriptionRoot+"/:id/pull", ss.Pull)
	r.Post(subscriptionRoot+"/:id/ack", ss.Ack)
	r.Post(subscriptionRoot+"/:id/ack/modify", ss.ModifyAck)
	r.Post(subscriptionRoot+"/:id/push/modify", ss.ModifyPush)
	r.Delete(subscriptionRoot+"/:id", ss.Delete)

	ms := Monitoring{}
	monitoringRoot := "/stats"
	r.Get(monitoringRoot+"/", ms.Summary)
	r.Get(monitoringRoot+"/topic", ms.TopicSummary)
	r.Get(monitoringRoot+"/topic/:id", ms.TopicDetail)
	r.Get(monitoringRoot+"/subscription", ms.SubscriptionSummary)
	r.Get(monitoringRoot+"/subscription/:id", ms.SubscriptionDetail)
	return r
}

// Server is topic and subscription frontend server
type Server struct {
	cfg *Config
}

// NewServer return initialized server
func NewServer(path string) (*Server, error) {
	c, err := LoadConfigFromFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load config")
	}
	return &Server{
		cfg: c,
	}, nil
}

// PrepareServer settings datastore and stats configuration
func (s *Server) PrepareServer() error {
	stats.Initialize()
	return s.InitDatastore()
}

// InitDatastore prepare datastore initialize
func (s *Server) InitDatastore() error {
	datastore.SetGlobalConfig(s.cfg.Datastore)
	if err := models.InitDatastoreTopic(); err != nil {
		return errors.Wrap(err, "failed to init datastore topic")
	}
	if err := models.InitDatastoreSubscription(); err != nil {
		return errors.Wrap(err, "failed to init datastore subscription")
	}
	if err := models.InitDatastoreMessage(); err != nil {
		return errors.Wrap(err, "failed to init datastore message")
	}
	if err := models.InitDatastoreMessageStatus(); err != nil {
		return errors.Wrap(err, "failed to init datastore message status")
	}
	return nil
}

// Run start server
func (s *Server) Run(port int) error {
	log.Printf("Pubsub server running at http://localhost:%d/", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), Routes())
}
