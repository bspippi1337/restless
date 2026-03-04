#!/usr/bin/env bash
set -e

TARGET=https://api.github.com

clear
printf "\033[38;5;214m⚙ RESTLESS RECON ENGINE ⚙\033[0m\n"
printf "\033[38;5;220msteampunk network probe initiating...\033[0m\n"
sleep 1

echo
printf "\033[38;5;45mTarget locked:\033[0m $TARGET"
sleep 1

echo
echo "$ restless auto $TARGET"
sleep 1
./build/restless auto $TARGET
sleep 2

echo
printf "\033[38;5;213mMapping topology...\033[0m"
sleep 1

echo
echo "api.github.com"
sleep 0.4
echo "├── users"
sleep 0.4
echo "│   ├── /{user}"
sleep 0.4
echo "│   └── /{user}/repos"
sleep 0.4
echo "├── repos"
sleep 0.4
echo "│   ├── /{owner}/{repo}"
sleep 0.4
echo "│   └── /{owner}/{repo}/issues"
sleep 0.4
echo "└── search"
sleep 1

echo
printf "\033[38;5;208mDispatching swarm probes...\033[0m"
sleep 1

echo "runner-alpha probing /api"
sleep 0.7
echo "runner-beta probing /v1"
sleep 0.7
echo "runner-gamma probing /graphql"
sleep 1

echo
printf "\033[38;5;46m✓ reconnaissance complete\033[0m"
sleep 1
