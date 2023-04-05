package hauler_common

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func timeParse(v interface{}, m map[string]interface{}) (timestampOk bool) {
	if vstr, vstrOk := v.(string); vstrOk {
		if _, err := time.Parse(time.RFC3339, vstr); err == nil {
			m["timestamp"] = vstr
			timestampOk = true
		}
	} else if vint, vintOk := v.(int); vintOk {
		now := time.Now().Unix()
		if vint > int(now) {
			m["timestamp"] = time.Unix(int64(vint/1000), 0).Format(time.RFC3339)
			timestampOk = true
		} else {
			m["timestamp"] = time.Unix(int64(vint), 0).Format(time.RFC3339)
			timestampOk = true
		}
	} else if vfloat, vfloatOk := v.(float64); vfloatOk {
		now := time.Now().Unix()
		if int64(vfloat) > now {
			m["timestamp"] = time.Unix(int64(vfloat/1000), 0).Format(time.RFC3339)
			timestampOk = true
		} else {
			m["timestamp"] = time.Unix(int64(vfloat), 0).Format(time.RFC3339)
			timestampOk = true
		}
	}
	return
}

func ParseEvent(meventBytes []byte, flavor string) (map[string]interface{}, error) {
	mevent := map[string]interface{}{}
	// JSON Processing
	je := json.Unmarshal(meventBytes, &mevent)
	if je != nil {
		logEntry.ContextLogger().Errorf("Error unmarshalling event: %v %s", je, mevent)
		return nil, je
	}

	m := map[string]interface{}{}
	// We also flatten, if it contains event as a key, everything under event is flattened and event removed
	if _, ok := mevent["event"]; ok {
		if eventMap, eventIsMap := mevent["event"].(map[string]interface{}); eventIsMap {
			for k, v := range eventMap {
				if strings.HasPrefix(k, "event.") {
					mevent[strings.TrimPrefix(k, "event.")] = v
				} else {
					mevent[k] = v
				}
			}
			delete(mevent, "event")
		}
	}

	for k, v := range mevent {
		if strings.HasPrefix(k, "event.") {
			m[strings.TrimPrefix(k, "event.")] = v
		} else {
			m[k] = v
		}
	}

	var timestampOk = false

	switch flavor {
	// oracle cloud logs flavor
	case "oci":
		if _, ok := mevent["type"]; ok {
			m["proc_id"] = mevent["type"]
		}
		if dataval, datavalok := mevent["data"]; datavalok {
			if data, dataIsMapok := dataval.(map[string]interface{}); dataIsMapok {
				if _, messageok := data["message"]; messageok {
					m["message"] = data["message"]
				}
			}
		}
		timestampOk = true // time is there in oracle json
		fallthrough
	default:
		for k, v := range mevent {
			klower := strings.ToLower(k)
			if _, ok := timestampAlias[klower]; ok {
				timestampOk = timeParse(v, m)
			}
			if _, ok := MessageAliases[klower]; ok {
				if vstr, vstrOk := v.(string); vstrOk {
					m[k] = vstr
				} else if vm, vmok := v.(map[string]interface{}); vmok {
					b, be := json.Marshal(vm)
					if be != nil {
						//logEntry.ContextLogger().Infof("Error sending event: %s %v", be.Error(), vm)
					} else {
						m[k] = fmt.Sprintf("%s", b)
					}
				} else if vm, vmok := v.([]interface{}); vmok {
					b, be := json.Marshal(vm)
					if be != nil {
						//logEntry.ContextLogger().Infof("Error marshaling event: %s %v", be.Error(), vm)
					} else {
						m[k] = fmt.Sprintf("%s", b)
					}
				}
			}
			if _, ok := LevelAliases[klower]; ok {
				m["severity_string"] = v
			}
		}
	}

	if !timestampOk {
		if _, ok := m["timestamp"]; !ok {
			m["timestamp"] = time.Now().Format(time.RFC3339)
		}
	}

	return m, nil
}
