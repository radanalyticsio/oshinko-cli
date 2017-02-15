package app

import (
	"sort"
	"strings"

	kapi "k8s.io/kubernetes/pkg/api"
)

// Environment holds environment variables for new-app
type Environment map[string]string

// ParseEnvironment converts the provided strings in key=value form into environment
// entries.
func ParseEnvironment(vals ...string) Environment {
	env := make(Environment)
	for _, s := range vals {
		if i := strings.Index(s, "="); i == -1 {
			env[s] = ""
		} else {
			env[s[:i]] = s[i+1:]
		}
	}
	return env
}

// NewEnvironment returns a new set of environment variables based on all
// the provided environment variables
func NewEnvironment(envs ...map[string]string) Environment {
	if len(envs) == 1 {
		return envs[0]
	}
	out := make(Environment)
	out.Add(envs...)
	return out
}

// Add adds the environment variables to the current environment
func (e Environment) Add(envs ...map[string]string) {
	for _, env := range envs {
		for k, v := range env {
			e[k] = v
		}
	}
}

// List sorts and returns all the environment variables
func (e Environment) List() []kapi.EnvVar {
	env := []kapi.EnvVar{}
	for k, v := range e {
		env = append(env, kapi.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	sort.Sort(sortedEnvVar(env))
	return env
}

type sortedEnvVar []kapi.EnvVar

func (m sortedEnvVar) Len() int           { return len(m) }
func (m sortedEnvVar) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m sortedEnvVar) Less(i, j int) bool { return m[i].Name < m[j].Name }

// JoinEnvironment joins two different sets of environment variables
// into one, leaving out all the duplicates
func JoinEnvironment(a, b []kapi.EnvVar) (out []kapi.EnvVar) {
	out = a
	for i := range b {
		exists := false
		for j := range a {
			if a[j].Name == b[i].Name {
				exists = true
				break
			}
		}
		if exists {
			continue
		}
		out = append(out, b[i])
	}
	return out
}
