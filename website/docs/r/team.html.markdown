---
layout: "signalfx"
page_title: "SignalFx: signalfx_resource"
sidebar_current: "docs-signalfx-resource-team"
description: |-
  Allows Terraform to create and manage SignalFx teams
---

# Resource: signalfx_team

Handles management of SignalFx teams.

You can configure [team notification policies](https://docs.signalfx.com/en/latest/managing/teams/team-notifications.html) using this resource and the various `notifications_*` properties.

## Example Usage

```terraform
resource "signalfx_team" "myteam0" {
    name = "Best Team Ever"
    description = "Super great team no jerks definitely"

    notifications_critical = [
      "PagerDuty,credentialId"
    ]

    notificiations_info = [
      "Email,notify@example.com"
    ]
}
```

## Argument Reference

The following arguments are supported in the resource block:

* `name` - (Required) Name of the team.
* `description` - (Optional) Description of the team.
* `notifications_critical` - (Optional) Where to send notifications for critical alerts
* `notifications_default` - (Optional) Where to send notifications for default alerts
* `notifications_info` - (Optional) Where to send notifications for info alerts
* `notifications_major` - (Optional) Where to send notifications for major alerts
* `notifications_minor` - (Optional) Where to send notifications for minor alerts
* `notifications_warning` - (Optional) Where to send notifications for warning alerts
