package model

import (
	"fmt"
	. "github.com/mickael-kerjean/filestash/server/common"
	"strings"
)

func MergeConnectionDefaults(conn map[string]string) {
	label := conn["label"]
	if label == "" {
		return
	}
	tmplBind := TmplParams(conn)
	for k, v := range ConnectionDefaults(label) {
		v = resolveConnectionTemplate(v, tmplBind)
		if conn[k] == "" {
			conn[k] = v
		} else {
			conn[k] = resolveConnectionTemplate(conn[k], tmplBind)
		}
	}
}

func resolveConnectionTemplate(value string, tmplBind map[string]string) string {
	if value == "" || !strings.Contains(value, "{{") {
		return value
	}
	out, err := TmplExec(value, tmplBind)
	if err != nil {
		return value
	}
	return out
}

func NewBackend(ctx *App, conn map[string]string) (IBackend, error) {
	isAllowed := func() bool {
		// by default, a hacker could use filestash to establish connections outside of what's
		// define in the config file. We need to prevent this
		possibilities := make([]map[string]interface{}, 0)
		for i := 0; i < len(Config.Conn); i++ {
			d := Config.Conn[i]
			if d["type"] != conn["type"] {
				continue
			}
			if connLabel := conn["label"]; connLabel != "" {
				if fmt.Sprint(d["label"]) != connLabel {
					continue
				}
			}
			if val, ok := d["hostname"]; ok == true {
				if val != conn["hostname"] {
					continue
				}
			}
			if val, ok := d["path"]; ok == true {
				if val == nil {
					val = "/"
				}
				configPath, ok := val.(string)
				if ok == false {
					continue
				}
				connPath := conn["path"]
				if connPath == "" {
					connPath = "/"
				}
				if strings.HasPrefix(connPath, configPath) == false {
					continue
				}
			}
			if val, ok := d["bucket"]; ok == true {
				configBucket := fmt.Sprint(val)
				if configBucket != "" && conn["bucket"] != "" && conn["bucket"] != configBucket {
					continue
				}
			}
			if val, ok := d["url"]; ok == true {
				if val != conn["url"] {
					continue
				}
			}
			possibilities = append(possibilities, Config.Conn[i])
		}
		if len(possibilities) > 0 {
			return true
		}
		return false
	}

	if isAllowed() == false {
		return Backend.Get(BACKEND_NIL), ErrNotAllowed
	}
	return Backend.Get(conn["type"]).Init(conn, ctx)
}

func ConnectionDefaults(label string) map[string]string {
	defaults := map[string]string{}
	if label == "" {
		return defaults
	}
	for i := range Config.Conn {
		if fmt.Sprint(Config.Conn[i]["label"]) != label {
			continue
		}
		for k, v := range Config.Conn[i] {
			if k == "label" {
				continue
			}
			if v == nil {
				continue
			}
			defaults[k] = fmt.Sprintf("%v", v)
		}
		break
	}
	return defaults
}

func GetHome(b IBackend, base string) (string, error) {
	if strings.TrimSpace(base) == "" {
		base = "/"
	}
	home := "/"
	if obj, ok := b.(interface{ Home() (string, error) }); ok {
		tmp, err := obj.Home()
		if err != nil {
			return base, err
		}
		home = EnforceDirectory(tmp)
	} else if _, err := b.Ls(base); err != nil {
		return base, err
	}

	base = EnforceDirectory(base)
	if strings.HasPrefix(home, base) {
		return "/" + home[len(base):], nil
	}
	return "/", nil
}

func MapStringInterfaceToMapStringString(m map[string]interface{}) map[string]string {
	res := make(map[string]string)
	for key, value := range m {
		res[key] = fmt.Sprintf("%v", value)
		if res[key] == "<nil>" {
			res[key] = ""
		}
	}
	return res
}
