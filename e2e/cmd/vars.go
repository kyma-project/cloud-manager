package main

import "time"

var (
	timeout       time.Duration
	all           bool
	runtimeID     string
	alias         string
	provider      string
	waitDone      bool
	moduleName    string
	modules       []string
	outputFormat  string
	listOnly      bool
	ignoreAll     bool
	ignoreNone    bool
	ignoreAliases []string
	enableAliases []string
	verbose       bool
	dryRun        bool
	force         bool
)
