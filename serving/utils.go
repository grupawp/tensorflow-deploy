package serving

import (
	"github.com/grupawp/tensorflow-deploy/app"
)

func removeVersionLabels(labels map[string]int64, version int64) map[string]int64 {
	result := make(map[string]int64, 0)
	for k, v := range labels {
		if v == version {
			continue
		}
		result[k] = v
	}
	return result
}

func removeLabel(labels map[string]int64, label string) map[string]int64 {
	result := make(map[string]int64, 0)
	for k, v := range labels {
		if k == label {
			continue
		}
		result[k] = v
	}
	return result
}

func isLabelStable(labels map[string]int64, version int64) bool {
	return labels[app.StableLabel] == version
}

func versionByLabel(labels map[string]int64, label string) (int64, bool) {
	versionID, found := labels[label]
	return versionID, found
}
