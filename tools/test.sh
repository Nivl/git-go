#!/bin/bash

set -ex

go test -v -race -covermode=atomic ./...
