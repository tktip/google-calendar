#!/bin/sh

GO111MODULE=on go mod verify
GO111MODULE=on go mod tidy
