package hauler_common

import (
	"github.com/logiqai/utils/log"
	"github.com/sirupsen/logrus"
)

const (
	FLAVOR_OCI_BUCKETS = "oci"
)

var (
	logger         = log.NewPackageLogger(logrus.InfoLevel)
	logEntry       = log.NewPackageLoggerEntry(logger, "hauler-common")
	timestampAlias = map[string]interface{}{"timestamp": true, "time": true, "@timestamp": true}
	MessageAliases = map[string]interface{}{"message": true, "@message": true, "log": true, "@log": true}
	LevelAliases   = map[string]interface{}{
		"severity":         true,
		"severe":           true,
		"severitystring":   true,
		"level":            true,
		"lvl":              true,
		"@severity":        true,
		"@level":           true,
		"otel.status_code": true,
	}
)
