package signalfx

import (
	"testing"

	"github.com/signalfx/signalfx-go/notification"
	"github.com/stretchr/testify/assert"
)

func TestNotifyStringFromAPI(t *testing.T) {
	values := []*notification.Notification{
		&notification.Notification{
			Type: EmailNotificationType,
			Value: &notification.EmailNotification{
				Type:  EmailNotificationType,
				Email: "foo@example.com",
			},
		},
		&notification.Notification{
			Type: OpsgenieNotificationType,
			Value: &notification.OpsgenieNotification{
				Type:          OpsgenieNotificationType,
				CredentialId:  "XXX",
				ResponderName: "Foo",
				ResponderId:   "ABC123",
				ResponderType: "Team",
			},
		},
		&notification.Notification{
			Type: PagerDutyNotificationType,
			Value: &notification.PagerDutyNotification{
				Type:         PagerDutyNotificationType,
				CredentialId: "XXX",
			},
		},
		&notification.Notification{
			Type: SlackNotificationType,
			Value: &notification.SlackNotification{
				Type:         SlackNotificationType,
				CredentialId: "XXX",
				Channel:      "foobar",
			},
		},
		&notification.Notification{
			Type: TeamNotificationType,
			Value: &notification.TeamNotification{
				Type: TeamNotificationType,
				Team: "ABC123",
			},
		},
		&notification.Notification{
			Type: TeamEmailNotificationType,
			Value: &notification.TeamEmailNotification{
				Type: TeamEmailNotificationType,
				Team: "ABC124",
			},
		},
		&notification.Notification{
			Type: WebhookNotificationType,
			Value: &notification.WebhookNotification{
				Type:         WebhookNotificationType,
				CredentialId: "XXX",
				Secret:       "YYY",
				Url:          "http://www.example.com",
			},
		},
		&notification.Notification{
			Type: BigPandaNotificationType,
			Value: &notification.BigPandaNotification{
				Type:         BigPandaNotificationType,
				CredentialId: "XXX",
			},
		},
		&notification.Notification{
			Type: Office365NotificationType,
			Value: &notification.Office365Notification{
				Type:         Office365NotificationType,
				CredentialId: "XXX",
			},
		},
		&notification.Notification{
			Type: ServiceNowNotificationType,
			Value: &notification.ServiceNowNotification{
				Type:         ServiceNowNotificationType,
				CredentialId: "XXX",
			},
		},
		&notification.Notification{
			Type: VictorOpsNotificationType,
			Value: &notification.VictorOpsNotification{
				Type:         VictorOpsNotificationType,
				CredentialId: "XXX",
				RoutingKey:   "YYY",
			},
		},
		&notification.Notification{
			Type: XMattersNotificationType,
			Value: &notification.XMattersNotification{
				Type:         XMattersNotificationType,
				CredentialId: "XXX",
			},
		},
	}

	expected := []string{
		"Email,foo@example.com",
		"Opsgenie,XXX,Foo,ABC123,Team",
		"PagerDuty,XXX",
		"Slack,XXX,foobar",
		"Team,ABC123",
		"TeamEmail,ABC124",
		"Webhook,XXX,YYY,http://www.example.com",
		"BigPanda,XXX",
		"Office365,XXX",
		"ServiceNow,XXX",
		"VictorOps,XXX,YYY",
		"XMatters,XXX",
	}

	for i, v := range values {
		result, err := getNotifyStringFromAPI(v)
		assert.NoError(t, err, "Got error making notify string")
		assert.Equal(t, expected[i], result)
	}

	for _, v := range expected {
		_, errors := validateNotification(v, "notification")
		assert.Len(t, errors, 0, "Expected no errors from valid notification: %q", v)
	}
}

func TestNotifyValidationBad(t *testing.T) {

	busted := []string{
		"Email,fooexample.com",
		"Opsgenie,XXX,Foo,ABC123",
		"PagerDuty",
		"Slack,XXX,#foobar",
		"Team",
		"FARTS,lol",
		"TeamEmailABC123",
		"Webhook,XXX,YYY,notaurl",
		"BigPanda",
		"Office365",
		"ServiceNow",
		"VictorOps,XXX",
		"XMatters",
	}

	for _, v := range busted {
		_, errors := validateNotification(v, "notification")
		assert.Len(t, errors, 1, "Expected errors from invalid notification %q", v)
	}
}

func TestGetNotifications(t *testing.T) {
	values := []interface{}{
		"Email,test@yelp.com",
		"PagerDuty,credId",
		"Webhook,credId,test,https://foo.bar.com?user=test&action=alert",
		"Opsgenie,credId,respName,respId,respType",
		"Slack,credId,channel",
		"Team,teamId",
		"TeamEmail,teamId",
		"BigPanda,credId",
		"Office365,credId",
		"ServiceNow,credId",
		"VictorOps,credId,routingKey",
		"XMatters,credId",
	}

	expected := []*notification.Notification{
		&notification.Notification{
			Type: EmailNotificationType,
			Value: &notification.EmailNotification{
				Type:  EmailNotificationType,
				Email: "test@yelp.com",
			},
		},
		&notification.Notification{
			Type: PagerDutyNotificationType,
			Value: &notification.PagerDutyNotification{
				Type:         PagerDutyNotificationType,
				CredentialId: "credId",
			},
		},
		&notification.Notification{
			Type: WebhookNotificationType,
			Value: &notification.WebhookNotification{
				Type:         WebhookNotificationType,
				CredentialId: "credId",
				Secret:       "test",
				Url:          "https://foo.bar.com?user=test&action=alert",
			},
		},
		&notification.Notification{
			Type: OpsgenieNotificationType,
			Value: &notification.OpsgenieNotification{
				Type:          OpsgenieNotificationType,
				CredentialId:  "credId",
				ResponderName: "respName",
				ResponderId:   "respId",
				ResponderType: "respType",
			},
		},
		&notification.Notification{
			Type: SlackNotificationType,
			Value: &notification.SlackNotification{
				Type:         SlackNotificationType,
				CredentialId: "credId",
				Channel:      "channel",
			},
		},
		&notification.Notification{
			Type: TeamNotificationType,
			Value: &notification.TeamNotification{
				Type: TeamNotificationType,
				Team: "teamId",
			},
		},
		&notification.Notification{
			Type: TeamEmailNotificationType,
			Value: &notification.TeamEmailNotification{
				Type: TeamEmailNotificationType,
				Team: "teamId",
			},
		},
		&notification.Notification{
			Type: BigPandaNotificationType,
			Value: &notification.BigPandaNotification{
				Type:         BigPandaNotificationType,
				CredentialId: "credId",
			},
		},
		&notification.Notification{
			Type: Office365NotificationType,
			Value: &notification.Office365Notification{
				Type:         Office365NotificationType,
				CredentialId: "credId",
			},
		},
		&notification.Notification{
			Type: ServiceNowNotificationType,
			Value: &notification.ServiceNowNotification{
				Type:         ServiceNowNotificationType,
				CredentialId: "credId",
			},
		},
		&notification.Notification{
			Type: VictorOpsNotificationType,
			Value: &notification.VictorOpsNotification{
				Type:         VictorOpsNotificationType,
				CredentialId: "credId",
				RoutingKey:   "routingKey",
			},
		},
		&notification.Notification{
			Type: XMattersNotificationType,
			Value: &notification.XMattersNotification{
				Type:         XMattersNotificationType,
				CredentialId: "credId",
			},
		},
	}
	nots, err := getNotifications(values)
	assert.NoError(t, err, "No error expected on notification conversion")
	assert.Equal(t, expected, nots)
}
