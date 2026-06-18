# cts — Cut The Shit

CLI em Go que acha e remove **skills, agentes, plugins e MCP mortos** da sua máquina — com dry-run, confirmação e backup. Nasceu de uma faxina manual que virou ferramenta.

## Por quê

Ferramentas de IA (Claude Code, codex, opencode, pi, vários agentes) deixam lixo espalhado: skills órfãs, binários de agentes que você não usa mais, caches de plugin, marketplaces órfãos, MCP servers configurados e esquecidos. Isso ocupa disco e — pior — infla o contexto injetado em todo prompt. O `cts` automatiza a limpeza com segurança.

## Instalar

```bash
go build -o cts.exe .
```

## Uso

```bash
cts scan        # relatório read-only: o que está morto e quanto ocupa
# cut / purge: em construção (sempre com dry-run + confirmação + backup)
```

## Status

MVP em construção, incremento a incremento:

- [x] `scan` de skills
- [ ] `scan` de agentes (bins + config)
- [ ] `scan` de plugins/marketplaces
- [ ] `scan` de MCP servers
- [ ] `cut` / `purge` com backup

## Desenvolvimento

- **Como trabalhar aqui (comandos, segurança, workflow):** [`docs/WORKING.md`](docs/WORKING.md)
- **Arquitetura e design:** [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)
- **Decisões de design:** [`docs/adr/`](docs/adr/)

Gate local antes de commitar: `./scripts/check.sh`
