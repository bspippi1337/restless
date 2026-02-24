# Restless ⚡

Terminal-First API Workbench

Version: 4.0.4

Restless is a modular OpenAPI-aware execution engine built for shell-native development.

## Install

go build -o restless ./cmd/restless

## Example

restless openapi import petstore.json  
restless openapi run <id> GET /pets  

## Philosophy

- Scriptable
- Deterministic
- CI-native
- Modular

See /docs for full documentation.
