package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// NewID generates a URL-safe, cryptographically random unique identifier.
func NewID() string {
	b := make([]byte, 12)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// EnvVarsToSlice converts EnvVar slice to ["KEY=VALUE", ...] format.
func EnvVarsToSlice(vars []EnvVar) []string {
	res := make([]string, 0, len(vars))
	for _, v := range vars {
		res = append(res, v.Key+"="+v.Value)
	}
	return res
}

// EnvVarsFromSlice parses ["KEY=VALUE", ...] lines into EnvVar slice.
func EnvVarsFromSlice(lines []string) []EnvVar {
	res := make([]EnvVar, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			res = append(res, EnvVar{Key: strings.TrimSpace(parts[0]), Value: strings.TrimSpace(parts[1])})
		} else if len(parts) == 1 {
			res = append(res, EnvVar{Key: strings.TrimSpace(parts[0]), Value: ""})
		}
	}
	return res
}

// EnvVarsToMap converts EnvVar slice to map[string]string.
func EnvVarsToMap(vars []EnvVar) map[string]string {
	m := make(map[string]string, len(vars))
	for _, v := range vars {
		m[v.Key] = v.Value
	}
	return m
}

// EnvVarsFromMap converts map[string]string to EnvVar slice.
func EnvVarsFromMap(m map[string]string) []EnvVar {
	res := make([]EnvVar, 0, len(m))
	for k, v := range m {
		res = append(res, EnvVar{Key: k, Value: v})
	}
	return res
}

// String returns formatted port mapping e.g. "8080:80/tcp".
func (p PortMapping) String() string {
	proto := p.Protocol
	if proto == "" {
		proto = "tcp"
	}
	return fmt.Sprintf("%d:%d/%s", p.HostPort, p.ContainerPort, proto)
}

// ParsePortMapping parses port mapping strings such as "8080:80", "8080:80/tcp", or "8080:80/udp".
func ParsePortMapping(s string) (PortMapping, error) {
	s = strings.TrimSpace(s)
	proto := "tcp"
	if idx := strings.Index(s, "/"); idx != -1 {
		proto = strings.ToLower(s[idx+1:])
		s = s[:idx]
	}

	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return PortMapping{}, ErrValidation
	}

	hostPort, err := strconv.Atoi(parts[0])
	if err != nil || hostPort <= 0 || hostPort > 65535 {
		return PortMapping{}, ErrValidation
	}

	containerPort, err := strconv.Atoi(parts[1])
	if err != nil || containerPort <= 0 || containerPort > 65535 {
		return PortMapping{}, ErrValidation
	}

	if proto != "tcp" && proto != "udp" {
		return PortMapping{}, ErrValidation
	}

	return PortMapping{
		HostPort:      hostPort,
		ContainerPort: containerPort,
		Protocol:      proto,
	}, nil
}

// String returns formatted volume mount string e.g. "myvol:/data:ro" or "myvol:/data".
func (v VolumeMount) String() string {
	mode := "rw"
	if v.ReadOnly {
		mode = "ro"
	}
	src := v.Name
	if v.HostPath != "" {
		src = v.HostPath
	}
	return fmt.Sprintf("%s:%s:%s", src, v.ContainerPath, mode)
}

// ParseVolumeMount parses volume strings such as "vol:/data", "vol:/data:ro", "/host/path:/data:rw".
func ParseVolumeMount(s string) (VolumeMount, error) {
	parts := strings.Split(strings.TrimSpace(s), ":")
	if len(parts) < 2 || len(parts) > 3 {
		return VolumeMount{}, ErrValidation
	}

	src := parts[0]
	dest := parts[1]
	if src == "" || dest == "" {
		return VolumeMount{}, ErrValidation
	}

	readOnly := false
	if len(parts) == 3 {
		mode := strings.ToLower(parts[2])
		if mode == "ro" {
			readOnly = true
		} else if mode != "rw" {
			return VolumeMount{}, ErrValidation
		}
	}

	vm := VolumeMount{
		ContainerPath: dest,
		ReadOnly:      readOnly,
	}
	if strings.HasPrefix(src, "/") || strings.HasPrefix(src, "./") {
		vm.Type = "bind"
		vm.HostPath = src
		vm.Name = ""
	} else {
		vm.Type = "volume"
		vm.Name = src
	}
	return vm, nil
}

// ExpandEnvVars replaces macros like $(PROJECT_NAME) or $(PRIMARY_DOMAIN) in env var values.
func ExpandEnvVars(envVars []EnvVar, macros map[string]string) []EnvVar {
	if len(envVars) == 0 || len(macros) == 0 {
		return envVars
	}
	out := make([]EnvVar, len(envVars))
	for i, ev := range envVars {
		val := ev.Value
		for k, v := range macros {
			val = strings.ReplaceAll(val, "$("+k+")", v)
		}
		out[i] = EnvVar{
			Key:      ev.Key,
			Value:    val,
			IsSecret: ev.IsSecret,
		}
	}
	return out
}
