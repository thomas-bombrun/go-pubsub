package server

import (
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
)

func TestSummary(t *testing.T) {
	ts := setupServer(t)
	defer ts.Close()

	cases := []struct {
		prepare func(client *http.Client)
		expect  []byte
	}{
		{
			func(client *http.Client) {
				createDummyTopic(t, ts, "topic1")
				createDummyTopic(t, ts, "topic2")
			},
			[]byte(`{"topic.topic_num":2.0,"subscription.subscription_num":0.0,"topic.message_count":0.0,"subscription.message_count":0.0}`),
		},
		{
			func(client *http.Client) {
				createDummyTopic(t, ts, "topic3")
				setupPublishMessages(t, ts, "topic3", PublishDatas{
					Messages: []PublishData{
						PublishData{Data: []byte(`test1`), Attr: nil},
						PublishData{Data: []byte(`test2`), Attr: map[string]string{"1": "2"}},
					},
				})
			},
			[]byte(`{"topic.topic_num":3.0,"subscription.subscription_num":0.0,"topic.message_count":2.0,"subscription.message_count":0.0}`),
		},
		{
			func(client *http.Client) {
				createDummySubscription(t, ts, ResourceSubscription{
					Topic:      "topic1",
					Name:       "sub1",
					AckTimeout: 1,
				})
				setupPublishMessages(t, ts, "topic1", PublishDatas{
					Messages: []PublishData{PublishData{Data: []byte(`test1`), Attr: nil}},
				})
			},
			[]byte(`{"topic.topic_num":3.0,"subscription.subscription_num":1.0,"topic.message_count":3.0,"subscription.message_count":1.0}`),
		},
	}
	for i, c := range cases {
		client := dummyClient(t)
		c.prepare(client)

		res, err := client.Get(ts.URL + "/stats")
		if err != nil {
			t.Fatalf("#%d: want non error, got %v", i, err)
		}
		defer res.Body.Close()

		payload, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("#%d: want non error, got %v", i, err)
		}
		if !reflect.DeepEqual(c.expect, payload) {
			t.Errorf("#%d: want response payload %s, got %s", i, c.expect, payload)
		}
	}
}

func TestTopicSummary(t *testing.T) {
	ts := setupServer(t)
	defer ts.Close()

	cases := []struct {
		prepare func(client *http.Client)
		expect  []byte
	}{
		{
			func(client *http.Client) {
				createDummyTopic(t, ts, "topic1")
				createDummyTopic(t, ts, "topic2")
			},
			[]byte(`{"topic.topic_num":2.0,"topic.message_count":0.0}`),
		},
	}
	for i, c := range cases {
		client := dummyClient(t)
		c.prepare(client)

		res, err := client.Get(ts.URL + "/stats/topic")
		if err != nil {
			t.Fatalf("#%d: want non error, got %v", i, err)
		}
		defer res.Body.Close()

		payload, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("#%d: want non error, got %v", i, err)
		}
		if !reflect.DeepEqual(c.expect, payload) {
			t.Errorf("#%d: want response payload %s, got %s", i, c.expect, payload)
		}
	}
}

func TestSubscriptionSummary(t *testing.T) {
	ts := setupServer(t)
	defer ts.Close()

	cases := []struct {
		prepare func(client *http.Client)
		expect  []byte
	}{
		{
			func(client *http.Client) {
				createDummyTopic(t, ts, "topic1")
				createDummySubscription(t, ts, ResourceSubscription{
					Topic:      "topic1",
					Name:       "sub1",
					AckTimeout: 1,
				})
				createDummySubscription(t, ts, ResourceSubscription{
					Topic:      "topic1",
					Name:       "sub2",
					AckTimeout: 1,
				})
				setupPublishMessages(t, ts, "topic1", PublishDatas{
					Messages: []PublishData{
						PublishData{Data: []byte(`test1`)},
						PublishData{Data: []byte(`test2`)},
					},
				})
			},
			[]byte(`{"subscription.subscription_num":2.0,"subscription.message_count":4.0}`),
		},
	}
	for i, c := range cases {
		client := dummyClient(t)
		c.prepare(client)

		res, err := client.Get(ts.URL + "/stats/subscription")
		if err != nil {
			t.Fatalf("#%d: want non error, got %v", i, err)
		}
		defer res.Body.Close()

		payload, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("#%d: want non error, got %v", i, err)
		}
		if !reflect.DeepEqual(c.expect, payload) {
			t.Errorf("#%d: want response payload %s, got %s", i, c.expect, payload)
		}
	}
}
