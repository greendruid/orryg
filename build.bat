@echo off

go generate
go install -ldflags "-H windowsgui"
