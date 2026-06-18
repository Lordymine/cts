# cts — instruções de agente

CLI em Go que remove skills/agentes/plugins/MCP mortos da máquina do usuário.
**Esta ferramenta APAGA coisas do usuário — segurança vem antes de qualquer feature.**

## Regra de segurança (inegociável)

- **Dry-run é o padrão.** Nada é removido sem flag explícita de execução + confirmação do usuário.
- **Sempre backup antes de remover.** Vai para `.cts-backups/`.
- **NUNCA** rode `cts cut`/`cts purge` destrutivo contra a máquina real durante dev ou teste.
- **Testes usam diretório temporário** (`t.TempDir()`), nunca caminhos reais do usuário (`~/.claude`, etc.).
- Antes de remover algo: confira o alvo. Se o que está lá contradiz o que foi descrito, pare e avise.

## Como trabalhar aqui (puxe sob demanda — não infle o contexto)

Leia o arquivo só quando o caso bater:

- **Comandos (o que pode / o que NÃO pode) + workflow seguro** → `docs/WORKING.md`
- **Arquitetura, design de módulos, como adicionar um scanner** → `docs/ARCHITECTURE.md`
- **Decisões de design (ADRs)** → `docs/adr/`
- **Convenções idiomáticas de Go** → `~/.claude/go-conventions.md`
- **Regras de trabalho do Rafael (valores, loop XP, princípios)** → `~/.agents/AGENTS.md`

## Rápido

- Gate completo (fmt, vet, lint, test, build): `./scripts/check.sh`
- Rodar: `go run . scan`
- Testar: `go test ./...`

Commit só com o gate verde. Conventional commits. Sem co-author de agente. Feature em branch, nunca direto na main.
