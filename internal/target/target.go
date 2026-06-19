// Package target defines what cts finds and can remove.
// It is pure domain: it imports no IO.
package target

// Category classifies a cleanup target.
type Category string

const (
	Skill  Category = "skill"
	Agent  Category = "agent"
	Plugin Category = "plugin"
	MCP    Category = "mcp"
)

// Target is something cts found: a skill, agent, plugin or MCP server.
// Dead marks a safe removal candidate; Reason explains why.
type Target struct {
	Name      string
	Category  Category
	Paths     []string // files/dirs that disappear on removal
	SizeBytes int64
	Dead      bool
	Reason    string
	Uninstall []string // optional command to run before deleting (e.g. npm rm -g pkg, claude mcp remove)
}
