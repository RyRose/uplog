// Package calendar provides a service that interacts with the Google Calendar API.
package calendar

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	return tok, json.NewDecoder(f).Decode(tok)

}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) error {
	slog.Info("saving credential file", "path", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %w", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

type service struct {
	mu              *sync.Mutex
	srv             *calendar.Service
	eventBuffer     map[string]Event
	syncBufferGroup *sync.WaitGroup
}

func (s *service) Close() {
	s.syncBufferGroup.Wait()
}

func (s *service) syncBuffer(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.eventBuffer == nil {
		return nil
	}

	var retErr error
	for _, event := range s.eventBuffer {
		slog.DebugContext(ctx, "syncing event buffer", "date", event.ISO8601Date, "summary", event.Summary)
		if err := s.syncEvent(ctx, event); err != nil {
			retErr = fmt.Errorf("failed to sync event %s: %w", event.ISO8601Date, err)
		}
	}
	s.eventBuffer = nil
	return retErr
}

func (s *service) enqueueEvent(_ context.Context, event Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.eventBuffer == nil {
		s.eventBuffer = make(map[string]Event)
		s.syncBufferGroup.Add(1)
		go func(ctx context.Context) {
			slog.DebugContext(ctx, "kicking off sync buffer goroutine")
			// FIXME: 5 seconds is arbitrary, should be configurable
			<-time.After(time.Second * 5)
			if err := s.syncBuffer(ctx); err != nil {
				slog.ErrorContext(ctx, "failed to sync buffer", "error", err)
			}
			s.syncBufferGroup.Done()
		}(context.Background())
	}
	s.eventBuffer[event.ISO8601Date] = event
	return nil
}

func (s *service) syncEvent(ctx context.Context, event Event) error {
	start, err := time.Parse(time.DateOnly, event.ISO8601Date)
	if err != nil {
		return fmt.Errorf("unable to parse iso8601 date %s: %w", event.ISO8601Date, err)
	}

	events, err := s.srv.Events.List("primary").
		TimeMin(start.Add(-24*time.Hour).Format(time.RFC3339)).
		TimeMax(start.Add(24*time.Hour).Format(time.RFC3339)).
		Context(ctx).
		SharedExtendedProperty("uplog=true", fmt.Sprintf("date=%s", event.ISO8601Date)).
		Do()
	if err != nil {
		return fmt.Errorf("unable to retrieve events for date %s: %w", event.ISO8601Date, err)
	}

	var found bool
	for _, e := range events.Items {
		if !found && event.equals(newEvent(e)) {
			found = true
			continue
		}
		if err := s.srv.Events.Delete("primary", e.Id).Context(ctx).Do(); err != nil {
			slog.WarnContext(ctx, "failed to delete event, skipping event deletion", "id", e.Id, "error", err)
		}
	}

	if found || event.Summary == "" || event.Summary == "REST" {
		return nil
	}

	e, err := event.toCalendarEvent()
	if err != nil {
		return fmt.Errorf("unable to convert event to calendar event: %w", err)
	}

	_, err = s.srv.Events.Insert("primary", e).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("unable to create event: %w", err)
	}
	return nil
}

func (s *service) deleteAfter(ctx context.Context, t time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	events, err := s.srv.Events.List("primary").
		TimeMin(t.Format(time.RFC3339)).
		SharedExtendedProperty("uplog=true").
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("unable to retrieve events after date %v: %w", t, err)
	}

	var retErr error
	for _, event := range events.Items {
		err := s.srv.Events.Delete("primary", event.Id).Context(ctx).Do()
		if err == nil {
			continue
		}
		slog.ErrorContext(ctx, "failed to delete event", "event", event, "error", err)
		retErr = fmt.Errorf("failed to delete event %q: %w", event.Id, err)
	}
	return retErr
}

// Service is a service that interacts with the Google Calendar API.
type Service struct {
	mu              *sync.Mutex
	srv             *service
	cfg             *oauth2.Config
	oauthTokenPath  string
	asyncQueue      chan func(context.Context) error
	closed          bool
	asyncQueueDone  *sync.WaitGroup
	stateToVerifier map[string]string
}

func (s *Service) Close() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	close(s.asyncQueue)
	s.asyncQueueDone.Wait()
	if s.srv != nil {
		s.srv.Close()
	}
	s.closed = true
}

// Event represents a calendar event.
type Event struct {
	// Summary is the event summary.
	Summary string
	// ISO8601Date is the date of the event in ISO8601 format.
	ISO8601Date string
	// Description is the event description.
	Description string
}

// equalsWithoutPrefix returns true if the given event is equal to the other event without the prefix.
func (e *Event) equals(other Event) bool {
	return e.Summary == other.Summary && e.ISO8601Date == other.ISO8601Date && e.Description == other.Description
}

// toCalendarEvent converts the event to a calendar event.
func (e *Event) toCalendarEvent() (*calendar.Event, error) {
	start, err := time.Parse(time.DateOnly, e.ISO8601Date)
	if err != nil {
		return nil, fmt.Errorf("unable to parse iso8601 date %s: %w", e.ISO8601Date, err)
	}
	return &calendar.Event{
		Summary:     e.Summary,
		Description: e.Description,
		ColorId:     "7",
		Start: &calendar.EventDateTime{
			Date: e.ISO8601Date,
		},
		End: &calendar.EventDateTime{
			Date: start.Add(24 * time.Hour).Format(time.DateOnly),
		},
		ExtendedProperties: &calendar.EventExtendedProperties{
			Shared: map[string]string{
				"uplog": "true",
				"date":  e.ISO8601Date,
			},
		},
	}, nil
}

// newEvent creates a new event from a calendar event.
func newEvent(evt *calendar.Event) Event {
	return Event{
		Summary:     evt.Summary,
		ISO8601Date: evt.Start.Date,
		Description: evt.Description,
	}
}

// Sync syncs the given event with the calendar.
func (s *Service) Sync(ctx context.Context, event Event) error {
	if s == nil {
		return errors.New("service is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ready(); err != nil {
		return err
	}
	s.asyncQueue <- func(ctx context.Context) error {
		return s.srv.enqueueEvent(ctx, event)
	}
	return nil
}

// DeleteAfter deletes all events after the given date.
func (s *Service) DeleteAfter(ctx context.Context, t time.Time) error {
	if s == nil {
		return errors.New("service is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ready(); err != nil {
		return err
	}
	s.asyncQueue <- func(ctx context.Context) error {
		return s.srv.deleteAfter(ctx, t)
	}
	return nil
}

// Initialized returns true if the service has been initialized
// with a calendar oauth token and is ready to be used.
func (s *Service) Initialized() bool {
	if s == nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.srv != nil
}

// Initializable returns true if the service can be initialized.
// If false, it means Initialized() can never be true.
func (s *Service) Initializable() bool {
	return s != nil
}

// ready checks if the service is ready to be used.
// It returns an error if the service is not initialized or closed.
func (s *Service) ready() error {
	if s.srv == nil {
		return errors.New("service not initialized")
	}
	if s.closed {
		return errors.New("service is closed")
	}
	return nil
}

// Init initializes the service with the given auth code and verifier.
func (s *Service) Init(ctx context.Context, authCode, state string) error {
	if s == nil {
		return errors.New("service is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return errors.New("service is closed")
	}

	if s.srv != nil {
		slog.InfoContext(ctx, "service already initialized")
		return nil
	}

	verifier, ok := s.stateToVerifier[state]
	if !ok {
		return fmt.Errorf("no verifier found for state %s", state)
	}

	token, err := s.cfg.Exchange(ctx, authCode, oauth2.VerifierOption(verifier))
	if err != nil {
		return fmt.Errorf("unable to retrieve token from web: %v", err)
	}
	if err := saveToken(s.oauthTokenPath, token); err != nil {
		slog.WarnContext(ctx, "unable to save token to file", "error", err)
	}
	if err := s.initToken(ctx, token); err != nil {
		return fmt.Errorf("unable to initialize service with token: %w", err)
	}

	// Cleans up the stateToVerifier map since it's no longer needed after initialization.
	s.stateToVerifier = nil

	return nil
}

// initToken initializes the service with the given token.
func (s *Service) initToken(ctx context.Context, token *oauth2.Token) error {
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(s.cfg.Client(ctx, token)))
	if err != nil {
		return fmt.Errorf("unable to retrieve calendar client: %w", err)
	}
	s.srv = &service{
		mu:              &sync.Mutex{},
		srv:             srv,
		syncBufferGroup: &sync.WaitGroup{},
	}
	return nil
}

// AuthCodeURL returns the URL to redirect the user to for authorization.
func (s *Service) AuthCodeURL() (string, error) {
	if s == nil {
		return "", errors.New("service is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stateToVerifier == nil {
		return "", errors.New("service is already initialized")
	}

	state := fmt.Sprint(rand.Int())
	verifier := oauth2.GenerateVerifier()
	s.stateToVerifier[state] = verifier
	return s.cfg.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.S256ChallengeOption(verifier)), nil
}

// NewService creates a new calendar service with the given credentials and oauth token path.
// If the oauth token path is empty, the service will be left uninitialized. The service can
// be initialized later using the Init method.
func NewService(ctx context.Context, credentials []byte, oauthTokenPath string) (*Service, error) {

	// If modifying these scopes, delete your previously saved token.
	config, err := google.ConfigFromJSON(credentials, calendar.CalendarEventsScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %w", err)
	}

	asyncQueue := make(chan func(context.Context) error, 1000)
	srv := &Service{
		cfg:             config,
		mu:              &sync.Mutex{},
		oauthTokenPath:  oauthTokenPath,
		asyncQueue:      asyncQueue,
		asyncQueueDone:  &sync.WaitGroup{},
		stateToVerifier: make(map[string]string),
	}
	srv.asyncQueueDone.Add(1)

	go func(ctx context.Context) {
		for f := range asyncQueue {
			slog.DebugContext(ctx, "executing async function", "queue_length", len(asyncQueue))
			if err := f(ctx); err != nil {
				slog.WarnContext(ctx, "failed to execute function", "error", err)
			}
		}
		srv.asyncQueueDone.Done()
		slog.InfoContext(ctx, "async queue closed")
	}(context.Background())

	token, err := tokenFromFile(oauthTokenPath)
	if err != nil {
		slog.WarnContext(ctx, "unable to read token from file, leaving uninitialized", "error", err)
		return srv, nil
	}

	if err := srv.initToken(ctx, token); err != nil {
		slog.WarnContext(ctx, "unable to initialize service with token, leaving uninitialized", "error", err)
	}
	return srv, nil
}
