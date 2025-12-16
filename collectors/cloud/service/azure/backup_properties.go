package azure

import (
	"strings"

	"confirmate.io/core/api/ontology"
)

// backupsEmptyCheck checks if the backups list is empty and returns voc.Backup with enabled = false.
func backupsEmptyCheck(backups []*ontology.Backup) []*ontology.Backup {
	if len(backups) == 0 {
		return []*ontology.Backup{
			{
				Enabled: false,
			},
		}
	}

	return backups
}

// backupPolicyName returns the backup policy name of a given Azure ID
func backupPolicyName(id string) string {
	// split according to "/"
	s := strings.Split(id, "/")

	// We cannot really return an error here, so we just return an empty string
	if len(s) < 10 {
		return ""
	}
	return s[10]
}
