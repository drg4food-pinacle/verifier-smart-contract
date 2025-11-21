package abigen

import "go/version"

var (
	abigenBaseURL  string
	abigenVersions map[Version]abigenVersion
)

// Version represents a solc version.
type Version string

func (v Version) String() string { return string(v) }

// Compare returns -1, 0, or +1 depending on whether v < other, v == other, or
// v > other
func (v Version) Cmp(other Version) int {
	return version.Compare(string(v), string(other))
}

// Versions is a list of all available solc versions.
var Versions []Version
