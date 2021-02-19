# Crocodile Calculator

Calculate ratios of meetings on your calendar.

<!-- toc -->

- [Setup](#setup)
- [To do](#to-do)

<!-- tocstop -->

## Setup

1. Configure your [OAuth consent screen](https://console.cloud.google.com/apis/credentials/consent)
1. Generate an [OAuth 2.0 Client ID](https://console.cloud.google.com/apis/credentials)
1. Download the generated Client ID as JSON, store it as `credentials.json` in this repository
1. `make deps` to install GoLang dependencies
1. `make run` to run the tool

### Example output

This week has not been a productive one:

```
$ make run
Parsing events between Sunday, February 14 and Saturday, February 20

Total time: 34h35m0s

Intellectual Property at 9h45m0s
Internal Meetings at 3h20m0s
OOO at 8h0m0s
Billable at 10h0m0s
Management at 3h30m0s

Unmatched events:
  Busy

  Billable usage is at 29%

  14h12m29 more of your hours should be billable to reach the billable utilization practice goal
```

## To do

[GitHub Issue tracker](https://github.com/smaslennikov/cal-calc/issues) tracks lacking features.
