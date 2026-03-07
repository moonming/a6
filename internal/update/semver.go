package update

import (
	"fmt"
	"strconv"
	"strings"
)

type Semver struct {
	Major int
	Minor int
	Patch int
	Pre   string
}

func ParseSemver(input string) (Semver, error) {
	v := strings.TrimSpace(input)
	if v == "" {
		return Semver{}, fmt.Errorf("empty version")
	}
	if v == "dev" {
		return Semver{Major: -1, Minor: -1, Patch: -1, Pre: "dev"}, nil
	}

	v = strings.TrimPrefix(v, "v")

	mainPart := v
	pre := ""
	if idx := strings.IndexByte(v, '-'); idx >= 0 {
		mainPart = v[:idx]
		pre = v[idx+1:]
		if pre == "" {
			return Semver{}, fmt.Errorf("invalid pre-release in %q", input)
		}
	}

	parts := strings.Split(mainPart, ".")
	if len(parts) != 3 {
		return Semver{}, fmt.Errorf("invalid semantic version %q", input)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil || major < 0 {
		return Semver{}, fmt.Errorf("invalid major version in %q", input)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil || minor < 0 {
		return Semver{}, fmt.Errorf("invalid minor version in %q", input)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil || patch < 0 {
		return Semver{}, fmt.Errorf("invalid patch version in %q", input)
	}

	return Semver{Major: major, Minor: minor, Patch: patch, Pre: pre}, nil
}

func (s Semver) Compare(other Semver) int {
	if s.Pre == "dev" && other.Pre != "dev" {
		return -1
	}
	if s.Pre != "dev" && other.Pre == "dev" {
		return 1
	}

	if s.Major != other.Major {
		if s.Major < other.Major {
			return -1
		}
		return 1
	}
	if s.Minor != other.Minor {
		if s.Minor < other.Minor {
			return -1
		}
		return 1
	}
	if s.Patch != other.Patch {
		if s.Patch < other.Patch {
			return -1
		}
		return 1
	}

	if s.Pre == "" && other.Pre == "" {
		return 0
	}
	if s.Pre == "" {
		return 1
	}
	if other.Pre == "" {
		return -1
	}

	if s.Pre < other.Pre {
		return -1
	}
	if s.Pre > other.Pre {
		return 1
	}
	return 0
}

func (s Semver) IsNewer(other Semver) bool {
	return s.Compare(other) < 0
}
