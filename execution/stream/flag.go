package stream

import (
	"fmt"

	"strconv"
	"strings"

	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/execution"
	"github.com/getgauge/gauge/filter"
	gm "github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/reporter"
	"github.com/getgauge/gauge/util"
)

var flagsMap = make(map[string]func(*gm.ExecutionRequestFlag) error)

const (
	parallelFlag  = "-parallel"
	verboseFlag   = "-verbose"
	tagsFlag      = "-tags"
	tableRowsFlag = "-table-rows"
	nFlag         = "n"
	strategyFlag  = "-strategy"
	envFlag       = "-env"
	sortFlag      = "-sort"
)

func init() {
	flagsMap[parallelFlag] = func(f *gm.ExecutionRequestFlag) error {
		v, e := strconv.ParseBool(f.GetValue())
		if e != nil {
			return fmt.Errorf("Invalid value for -%s flag. Error: %s", parallelFlag, e.Error())
		}
		reporter.IsParallel = v
		execution.InParallel = v
		return nil
	}
	flagsMap[verboseFlag] = func(f *gm.ExecutionRequestFlag) error {
		v, e := strconv.ParseBool(f.GetValue())
		if e != nil {
			return fmt.Errorf("Invalid value for -%s flag. Error: %s", verboseFlag, e.Error())
		}
		reporter.Verbose = v
		return nil
	}
	flagsMap[tagsFlag] = func(f *gm.ExecutionRequestFlag) error {
		filter.ExecuteTags = f.GetValue()
		return nil
	}
	flagsMap[tableRowsFlag] = func(f *gm.ExecutionRequestFlag) error {
		execution.TableRows = f.GetValue()
		return nil
	}
	flagsMap[nFlag] = func(f *gm.ExecutionRequestFlag) error {
		v, e := strconv.Atoi(f.GetValue())
		if e != nil {
			return fmt.Errorf("Invalid value for -%s flag. Error: %s", nFlag, e.Error())
		}
		execution.NumberOfExecutionStreams = v
		reporter.NumberOfExecutionStreams = v
		filter.NumberOfExecutionStreams = v
		return nil
	}
	flagsMap[strategyFlag] = func(f *gm.ExecutionRequestFlag) error {
		execution.Strategy = f.GetValue()
		return nil
	}
	flagsMap[envFlag] = func(f *gm.ExecutionRequestFlag) error {
		if e := env.LoadEnv(f.GetValue()); e != nil {
			return e
		}
		return nil
	}
	flagsMap[sortFlag] = func(f *gm.ExecutionRequestFlag) error {
		v, e := strconv.ParseBool(f.GetValue())
		if e != nil {
			return fmt.Errorf("Invalid value for -%s flag. Error: %s", sortFlag, e.Error())
		}
		filter.DoNotRandomize = v
		return nil
	}
}

func setFlags(flags []*gm.ExecutionRequestFlag) []error {
	resetFlags()
	var errs []error
	for _, f := range flags {
		operation, ok := flagsMap[f.GetName()]
		if !ok {
			errs = append(errs, fmt.Errorf("Invalid flag `%s=%s`", f.GetName(), strings.TrimSpace(strings.ToLower(f.GetValue()))))
			continue
		}
		err := operation(f)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if err := execution.ValidateFlags(); err != nil {
		errs = append(errs, err)
	}
	return errs
}

func resetFlags() {
	cores := util.NumberOfCores()
	reporter.IsParallel = false
	execution.InParallel = false
	reporter.Verbose = false
	filter.ExecuteTags = ""
	execution.TableRows = ""
	execution.NumberOfExecutionStreams = cores
	reporter.NumberOfExecutionStreams = cores
	filter.NumberOfExecutionStreams = cores
	execution.Strategy = "lazy"
	filter.DoNotRandomize = false
}
