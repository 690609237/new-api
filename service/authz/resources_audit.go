package authz

const ResourceAudit = "audit"

const ActionSensitiveRead = "sensitive_read"

var AuditRead = Permission{Resource: ResourceAudit, Action: ActionRead}
var AuditSensitiveRead = Permission{Resource: ResourceAudit, Action: ActionSensitiveRead}

func init() {
	RegisterResource(ResourceDefinition{
		Resource: ResourceAudit,
		LabelKey: "Audit Logs",
		Actions: []ActionDefinition{{
			Action:         ActionRead,
			LabelKey:       "View other accounts' audit logs",
			DescriptionKey: "View audit records from user and admin roles. Root records are always excluded.",
		}, {
			Action:         ActionSensitiveRead,
			LabelKey:       "View matched sensitive words",
			DescriptionKey: "View the configured sensitive words matched by a request.",
		}},
	})
}
