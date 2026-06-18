// Package target define o que o cts encontra e pode remover.
// É domínio puro: não importa nada de IO.
package target

// Category classifica um alvo de limpeza.
type Category string

const (
	Skill  Category = "skill"
	Agent  Category = "agent"
	Plugin Category = "plugin"
	MCP    Category = "mcp"
)

// Target é algo que o cts achou: uma skill, agente, plugin ou MCP.
// Dead marca candidato seguro a remoção; Reason explica por quê.
type Target struct {
	Name      string
	Category  Category
	Paths     []string // arquivos/dirs que somem se remover
	SizeBytes int64
	Dead      bool
	Reason    string
	Uninstall []string // comando opcional a rodar antes de apagar (ex.: npm rm -g pkg, claude mcp remove)
}
