package services

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

type EventAPI struct {
	application_name string
	producer         *KafkaProducer
	jwt_private_key  []byte
}

type EventAPIPayload struct {
	Record interface{} `json:"record"`
}

func NewEventAPI(brokers []string, application_name string, jwt_private_key string) (*EventAPI, error) {
	logger := NewLoggerService("EVENT_API")
	producer, err := NewKafkaProducer(brokers, logger)
	if err != nil {
		return nil, err
	}
	secret, err := base64.StdEncoding.DecodeString(jwt_private_key)
	if err != nil {
		return nil, err
	}

	return &EventAPI{
		application_name: application_name,
		producer:         producer,
		jwt_private_key:  secret,
	}, nil
}

func (e *EventAPI) generateJWT(event_payload EventAPIPayload) (string, error) {
	jwt_payload := jwt.MapClaims{
		"iat":   time.Now().Unix(),
		"jti":   strconv.FormatInt(time.Now().Unix(), 10),
		"iss":   e.application_name,
		"exp":   time.Now().UTC().Add(time.Hour).Unix(),
		"event": event_payload,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt_payload)

	return token.SignedString(e.jwt_private_key)
}

func (e *EventAPI) Notify(event_name string, event_payload EventAPIPayload) error {
	eventType := strings.Split(event_name, ".")[0]
	topic := fmt.Sprintf("%s.events.%s", e.application_name, eventType)
	jwt_token, err := e.generateJWT(event_payload)
	if err != nil {
		return err
	}

	e.producer.ProduceWithKey(topic, strings.Replace(event_name, eventType, "", 1), jwt_token)

	return nil
}
