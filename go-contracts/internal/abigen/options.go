package abigen

import "strings"

type Option func(*abigenArgs)

type abigenArgs struct {
	ContractPath string
	OutGo        string
	Pkg          string
	Additional   []string
}

func WithOutput(outGo string) Option {
	return func(a *abigenArgs) {
		a.OutGo = outGo
	}
}

func WithPackage(pkg string) Option {
	return func(a *abigenArgs) {
		a.Pkg = strings.ToLower(pkg)
	}
}

func WithAdditionalArgs(args ...string) Option {
	return func(a *abigenArgs) {
		a.Additional = append(a.Additional, args...)
	}
}
