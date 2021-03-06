package server

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/takashabe/go-pubsub/models"
	"github.com/takashabe/go-pubsub/stats"
)

// TopicServer is topic frontend server
type TopicServer struct{}

// Create is create topic
func (s *TopicServer) Create(w http.ResponseWriter, r *http.Request, id string) {
	t, err := models.NewTopic(id)
	if err != nil {
		Error(w, http.StatusNotFound, err, "failed to create topic")
		return
	}
	JSON(w, http.StatusCreated, t)

	stats.GetTopicAdapter().AddTopic(t.Name, 1)
}

// Get is get already exist topic
func (s *TopicServer) Get(w http.ResponseWriter, r *http.Request, id string) {
	t, err := models.GetTopic(id)
	if err != nil {
		Error(w, http.StatusNotFound, err, "not found topic")
		return
	}
	JSON(w, http.StatusOK, t)
}

// List is gets topic list
func (s *TopicServer) List(w http.ResponseWriter, r *http.Request) {
	t, err := models.ListTopic()
	if err != nil {
		Error(w, http.StatusNotFound, err, "not found topic")
		return
	}
	sort.Sort(models.ByTopicName(t))
	JSON(w, http.StatusOK, t)
}

// ResponseListSubscription represent response json of ListSubscription
type ResponseListSubscription struct {
	SubscriptionNames []string `json:"subscriptions"`
}

// ListSubscription is gets topic depends subscription list
func (s *TopicServer) ListSubscription(w http.ResponseWriter, r *http.Request, id string) {
	t, err := models.GetTopic(id)
	if err != nil {
		Error(w, http.StatusNotFound, err, "not found topic")
		return
	}
	subs, err := t.GetSubscriptions()
	if err != nil {
		Error(w, http.StatusNotFound, err, "not found subscription")
		return
	}
	sort.Sort(models.BySubscriptionName(subs))
	res := ResponseListSubscription{
		SubscriptionNames: make([]string, 0, len(subs)),
	}
	for _, s := range subs {
		res.SubscriptionNames = append(res.SubscriptionNames, s.Name)
	}
	JSON(w, http.StatusOK, res)
}

// Delete is delete topic
func (s *TopicServer) Delete(w http.ResponseWriter, r *http.Request, id string) {
	t, err := models.GetTopic(id)
	if err != nil {
		Error(w, http.StatusNotFound, err, "topic already not exist")
		return
	}
	if err := t.Delete(); err != nil {
		Error(w, http.StatusInternalServerError, err, "failed to delete topic")
		return
	}
	JSON(w, http.StatusNoContent, "")

	stats.GetTopicAdapter().AddTopic(t.Name, -1)
}

// PublishData represent post publish data
type PublishData struct {
	Data []byte            `json:"data"`
	Attr map[string]string `json:"attributes"`
}

// PublishDatas represent PublishData group
type PublishDatas struct {
	Messages []PublishData `json:"messages"`
}

// ResponsePublish represent reponse publish api
type ResponsePublish struct {
	MessageIDs []string `json:"message_ids"`
}

// Publish is publish message
func (s *TopicServer) Publish(w http.ResponseWriter, r *http.Request, id string) {
	// parse request
	decorder := json.NewDecoder(r.Body)
	var datas PublishDatas
	if err := decorder.Decode(&datas); err != nil {
		Error(w, http.StatusNotFound, err, "failed to parsed request")
		return
	}

	// publish message
	t, err := models.GetTopic(id)
	if err != nil {
		Error(w, http.StatusNotFound, err, "not found topic")
		return
	}
	pubIDs := make([]string, 0)
	for _, d := range datas.Messages {
		id, err := t.Publish(d.Data, d.Attr)
		if err != nil {
			Error(w, http.StatusInternalServerError, err, "failed publish message")
			return
		}
		pubIDs = append(pubIDs, id)
	}
	JSON(w, http.StatusOK, ResponsePublish{MessageIDs: pubIDs})

	stats.GetTopicAdapter().AddMessage(t.Name, len(datas.Messages))
}
