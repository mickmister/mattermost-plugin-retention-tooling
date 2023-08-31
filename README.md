# Mattermost Rentention Tooling plugin ![CI](https://github.com/mickmister/mattermost-plugin-retention-tooling/actions/ci.yml/badge.svg)

This plugin provides data rentention tools to augment the [data retention capabilities](https://docs.mattermost.com/comply/data-retention-policy.html) of Mattermost Enterprise Edition.

## Tools

### De-activated User Clean-up

TODO

### Channel Archiver

Will auto-archive any channels that have had no activity for more than some configurable number of days. 

**Job**: can be configured via the system console to run monthly/weekly/daily on a specific day of the week and time of day. 

**Slash command**: Can be run on-demand via `/channel-archiver` slash command.

