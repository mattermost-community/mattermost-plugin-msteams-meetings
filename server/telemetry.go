package main

type TelemetrySource string

const (
	telemetryStartSourceWebapp  TelemetrySource = "webapp"
	telemetryStartSourceCommand TelemetrySource = "command"
)

func (p *Plugin) trackConnect(userID string) {
	p.tracker.TrackUserEvent("connect", userID, map[string]interface{}{})
}

func (p *Plugin) trackDisconnect(userID string) {
	p.tracker.TrackUserEvent("disconnect", userID, map[string]interface{}{})
}

func (p *Plugin) trackMeetingStart(userID string, source TelemetrySource) {
	p.tracker.TrackUserEvent("meeting_started", userID, map[string]interface{}{
		"source": source,
	})
}

func (p *Plugin) trackMeetingDuplication(userID string) {
	p.tracker.TrackUserEvent("meeting_duplicated", userID, map[string]interface{}{})
}

func (p *Plugin) trackMeetingForced(userID string) {
	p.tracker.TrackUserEvent("meeting_forced", userID, map[string]interface{}{})
}
