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
cts             # modo interativo: lista, você marca o que remover, confirma (com backup)
cts scan        # relatório read-only: o que está morto e quanto ocupa
cts purge       # mostra o que removeria (só os mortos) — dry-run
cts purge --yes # remove os mortos de verdade, com backup em .cts-backups/
```

## Status

MVP em construção, incremento a incremento:

- [x] `scan` de skills
- [x] `scan` de agentes (bins + config órfã)
- [x] core de remoção (dry-run + backup) + `purge`
- [x] lista interativa (selecionar ativos pra remover)
- [ ] `scan` de plugins/marketplaces
- [ ] `scan` de MCP servers
- [ ] full uninstall de pacote ativo (`npm rm -g` etc.)

## Desenvolvimento

- **Como trabalhar aqui (comandos, segurança, workflow):** [`docs/WORKING.md`](docs/WORKING.md)
- **Arquitetura e design:** [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)
- **Decisões de design:** [`docs/adr/`](docs/adr/)

Gate local antes de commitar: `./scripts/check.sh`
