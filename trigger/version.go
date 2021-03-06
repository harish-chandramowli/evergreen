package trigger

import (
	"fmt"
	"strconv"
	"time"

	"github.com/evergreen-ci/evergreen"
	"github.com/evergreen-ci/evergreen/model/event"
	"github.com/evergreen-ci/evergreen/model/notification"
	"github.com/evergreen-ci/evergreen/model/task"
	"github.com/evergreen-ci/evergreen/model/version"
	restModel "github.com/evergreen-ci/evergreen/rest/model"
	"github.com/mongodb/grip/message"
	"github.com/pkg/errors"
)

const (
	objectVersion = "version"
)

func init() {
	registry.registerEventHandler(event.ResourceTypeVersion, event.VersionStateChange, makeVersionTriggers)
}

type versionTriggers struct {
	event    *event.EventLogEntry
	data     *event.VersionEventData
	version  *version.Version
	uiConfig evergreen.UIConfig

	base
}

func makeVersionTriggers() eventHandler {
	t := &versionTriggers{}
	t.base.triggers = map[string]trigger{
		triggerOutcome:                t.versionOutcome,
		triggerFailure:                t.versionFailure,
		triggerSuccess:                t.versionSuccess,
		triggerRegression:             t.versionRegression,
		triggerExceedsDuration:        t.versionExceedsDuration,
		triggerRuntimeChangeByPercent: t.versionRuntimeChange,
	}
	return t
}

func (t *versionTriggers) Fetch(e *event.EventLogEntry) error {
	var err error
	if err = t.uiConfig.Get(); err != nil {
		return errors.Wrap(err, "Failed to fetch ui config")
	}

	t.version, err = version.FindOne(version.ById(e.ResourceId))
	if err != nil {
		return errors.Wrap(err, "failed to fetch version")
	}
	if t.version == nil {
		return errors.New("couldn't find version")
	}

	var ok bool
	t.data, ok = e.Data.(*event.VersionEventData)
	if !ok {
		return errors.Errorf("version '%s' contains unexpected data with type '%T'", e.ResourceId, e.Data)
	}
	t.event = e

	return nil
}

func (t *versionTriggers) Selectors() []event.Selector {
	return MakeVersionSelectors(*t.version)
}

func (t *versionTriggers) makeData(sub *event.Subscription, pastTenseOverride string) (*commonTemplateData, error) {
	api := restModel.APIVersion{}
	if err := api.BuildFromService(t.version); err != nil {
		return nil, errors.Wrap(err, "error building json model")
	}

	data := commonTemplateData{
		ID:              t.version.Id,
		EventID:         t.event.ID,
		SubscriptionID:  sub.ID,
		DisplayName:     t.version.Id,
		Object:          objectVersion,
		Project:         t.version.Identifier,
		URL:             versionLink(&t.uiConfig, t.version.Id),
		PastTenseStatus: t.data.Status,
		apiModel:        &api,
	}
	slackColor := evergreenFailColor
	if data.PastTenseStatus == evergreen.VersionSucceeded {
		data.PastTenseStatus = "succeeded"
		slackColor = evergreenSuccessColor
	}
	data.slack = []message.SlackAttachment{
		{
			Title:     "Evergreen Version",
			TitleLink: data.URL,
			Color:     slackColor,
			Text:      t.version.Message,
		},
	}
	if pastTenseOverride != "" {
		data.PastTenseStatus = pastTenseOverride
	}

	return &data, nil
}

func (t *versionTriggers) generate(sub *event.Subscription, pastTenseOverride string) (*notification.Notification, error) {
	data, err := t.makeData(sub, pastTenseOverride)
	if err != nil {
		return nil, errors.Wrap(err, "failed to collect version data")
	}

	payload, err := makeCommonPayload(sub, t.Selectors(), data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build notification")
	}

	return notification.New(t.event, sub.Trigger, &sub.Subscriber, payload)
}

func (t *versionTriggers) versionOutcome(sub *event.Subscription) (*notification.Notification, error) {
	if t.data.Status != evergreen.VersionSucceeded && t.data.Status != evergreen.VersionFailed {
		return nil, nil
	}

	return t.generate(sub, "")
}

func (t *versionTriggers) versionFailure(sub *event.Subscription) (*notification.Notification, error) {
	if t.data.Status != evergreen.VersionFailed {
		return nil, nil
	}

	return t.generate(sub, "")
}

func (t *versionTriggers) versionSuccess(sub *event.Subscription) (*notification.Notification, error) {
	if t.data.Status != evergreen.VersionSucceeded {
		return nil, nil
	}

	return t.generate(sub, "")
}

func (t *versionTriggers) versionExceedsDuration(sub *event.Subscription) (*notification.Notification, error) {
	if t.data.Status != evergreen.VersionSucceeded && t.data.Status != evergreen.VersionFailed {
		return nil, nil
	}
	thresholdString, ok := sub.TriggerData[event.VersionDurationKey]
	if !ok {
		return nil, fmt.Errorf("subscription %s has no build time threshold", sub.ID)
	}
	threshold, err := strconv.Atoi(thresholdString)
	if err != nil {
		return nil, fmt.Errorf("subscription %s has an invalid time threshold", sub.ID)
	}

	maxDuration := time.Duration(threshold) * time.Second
	if !t.version.StartTime.Add(maxDuration).Before(t.version.FinishTime) {
		return nil, nil
	}
	return t.generate(sub, fmt.Sprintf("exceeded %d seconds", threshold))
}

func (t *versionTriggers) versionRuntimeChange(sub *event.Subscription) (*notification.Notification, error) {
	if t.data.Status != evergreen.VersionSucceeded && t.data.Status != evergreen.VersionFailed {
		return nil, nil
	}
	percentString, ok := sub.TriggerData[event.VersionPercentChangeKey]
	if !ok {
		return nil, fmt.Errorf("subscription %s has no percentage increase", sub.ID)
	}
	percent, err := strconv.ParseFloat(percentString, 64)
	if err != nil {
		return nil, fmt.Errorf("subscription %s has an invalid percentage", sub.ID)
	}

	lastGreen, err := t.version.LastSuccessful()
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving last green build")
	}
	if lastGreen == nil {
		return nil, nil
	}
	thisVersionDuration := float64(t.version.FinishTime.Sub(t.version.StartTime))
	prevVersionDuration := float64(lastGreen.FinishTime.Sub(lastGreen.StartTime))
	shouldNotify, percentChange := runtimeExceedsThreshold(percent, prevVersionDuration, thisVersionDuration)
	if !shouldNotify {
		return nil, nil
	}
	return t.generate(sub, fmt.Sprintf("changed in runtime by %.1f%% (over threshold of %s%%)", percentChange, percentString))
}

func (t *versionTriggers) versionRegression(sub *event.Subscription) (*notification.Notification, error) {
	if t.data.Status != evergreen.VersionFailed || t.version.Requester != evergreen.RepotrackerVersionRequester {
		return nil, nil
	}

	versionTasks, err := task.FindWithDisplayTasks(task.ByVersion(t.version.Id))
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving tasks for version")
	}
	for i := range versionTasks {
		task := &versionTasks[i]
		isRegression, _, err := isTaskRegression(sub.ID, task)
		if err != nil {
			return nil, errors.Wrap(err, "error evaluating task regression")
		}
		if isRegression {
			return t.generate(sub, "")
		}
	}
	return nil, nil
}

func MakeVersionSelectors(v version.Version) []event.Selector {
	selectors := []event.Selector{
		{
			Type: selectorID,
			Data: v.Id,
		},
		{
			Type: selectorProject,
			Data: v.Identifier,
		},
		{
			Type: selectorObject,
			Data: objectVersion,
		},
		{
			Type: selectorRequester,
			Data: v.Requester,
		},
	}
	if v.AuthorID != "" {
		selectors = append(selectors, event.Selector{Type: selectorOwner, Data: v.AuthorID})
	}
	return selectors
}
