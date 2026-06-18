# Prior Art — referências e o que aprendemos

Pesquisa antes de construir: não reinventar, não errar. Atualizar quando achar nova referência relevante.

## Conclusão

Nenhuma ferramenta faz a limpeza **unificada** (skills + agentes + plugins + MCP, multi-agente, com dry-run + backup) que o `cts` faz. O conceito é nosso. Mas há peças maduras que **validam a arquitetura** e dão conhecimento de localização/UX que seria fácil errar do zero.

## Referências

### npkill — `voidcosmos/npkill` (~9.3k★, TypeScript)
Scan de `node_modules` + lista tamanho + select interativo + delete.
- **Arquitetura:** `/src` + `/tests`, quase zero dependência, scan low-level rápido.
- **Segurança:** flag `--dry-run`, multi-select com preview, ⚠️ aviso em alvo crítico. **Deleta permanente, sem backup.**
- **O que pegamos:** o formato scan→select→delete, a flag `--dry-run`, ⚠️ em alvo arriscado, relatório de espaço liberado.
- **O que fazemos melhor:** dry-run é o **default** (não flag), e **backup** antes de apagar.

### mcp-server-manager — `vlazic/mcp-server-manager` (Go)
Manager single-binary cross-platform de MCP.
- **Arquitetura:** `cmd/` + `internal/` (+ `web/`) — valida nosso layout Go.
- **Localização de config (ouro):** por cliente — `~/.claude.json`, `~/.gemini/settings.json`; MCP sob a chave `mcpServers`; formatos diferentes (`type:"http"` no Claude vs `httpUrl` no Gemini).
- **O que pegamos:** o mapa de caminhos/formatos de config de MCP por cliente, pro nosso scanner de MCP. Não adivinhar.

### Outras (contexto)
- `radu2lupu/mcp-cleanup`, `YuancFeng/claude-code-cleanup`: limpam **processos** órfãos — categoria diferente (não é nosso foco de disco/config).
- `Guanff/claude-code-cleanup`: skill scan→confirm→remove de artefatos de sessão — confirma o padrão.

## Lacunas (sem prior-art direto)
- **Detecção de agentes instalados** (npm/bun/uv/go bins + config órfã): ninguém cobre. Desenho nosso — lista de instalados **injetada** (testável), flag de config órfã.
