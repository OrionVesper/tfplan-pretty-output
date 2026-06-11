package terraform

type Plan struct {
	FormatVersion    string           `json:"format_version"`
	TerraformVersion string           `json:"terraform_version"`
	ResourceChanges  []ResourceChange `json:"resource_changes"`
}

type ResourceChange struct {
	Address       string `json:"address"`
	ModuleAddress string `json:"module_address,omitempty"`
	Mode          string `json:"mode"`
	Type          string `json:"type"`
	Name          string `json:"name"`
	ProviderName  string `json:"provider_name"`
	Change        Change `json:"change"`
}

type Change struct {
	Actions         Actions                `json:"actions"`
	Before          map[string]interface{} `json:"before"`
	After           map[string]interface{} `json:"after"`
	AfterUnknown    interface{}            `json:"after_unknown"`
	BeforeSensitive interface{}            `json:"before_sensitive"`
	AfterSensitive  interface{}            `json:"after_sensitive"`
}

type Actions []string

func (a Actions) ActionType() ActionType {
	if len(a) == 0 {
		return ActionNoOp
	}
	if len(a) == 2 && a[0] == "delete" && a[1] == "create" {
		return ActionReplace
	}
	switch a[0] {
	case "create":
		return ActionCreate
	case "update":
		return ActionUpdate
	case "delete":
		return ActionDelete
	default:
		return ActionNoOp
	}
}

type ActionType string

const (
	ActionCreate  ActionType = "create"
	ActionUpdate  ActionType = "update"
	ActionDelete  ActionType = "delete"
	ActionReplace ActionType = "replace"
	ActionNoOp    ActionType = "no-op"
)

func (a ActionType) Symbol() string {
	switch a {
	case ActionCreate:
		return "+"
	case ActionUpdate:
		return "~"
	case ActionDelete:
		return "-"
	case ActionReplace:
		return "±"
	default:
		return "="
	}
}

type Summary struct {
	Creates  int `json:"creates"`
	Updates  int `json:"updates"`
	Deletes  int `json:"deletes"`
	Replaces int `json:"replaces"`
	NoOps    int `json:"no_ops"`
	Total    int `json:"total"`
}
