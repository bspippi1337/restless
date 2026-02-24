# CLI Usage

## Raw Request

restless -X POST -url https://api.example.com -d '{"hello":"ground friend"}'

## OpenAPI

restless openapi import spec.json
restless openapi ls
restless openapi endpoints <id>
restless openapi run <id> GET /path

## Profiles

restless profile set dev base=https://api.example.com
restless profile use dev
restless profile ls

## Preferences

restless prefs show
restless prefs set color=on
