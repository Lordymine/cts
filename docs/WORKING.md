# Como trabalhar no cts (com segurança)

Leia antes de rodar comandos ou mexer no código. Esta ferramenta **apaga coisas reais** — a disciplina aqui não é opcional.

## Comandos que PODE usar (dev)

| Comando | O que faz |
|---|---|
| `go run . scan` | Roda o scan (read-only, seguro) |
| `go test ./...` | Roda os testes |
| `go test -race ./...` | Testes com detector de corrida |
| `go vet ./...` | Análise estática do compilador |
| `golangci-lint run ./...` | Lint |
| `go build -o cts.exe .` | Compila o binário |
| `./scripts/check.sh` | Gate completo: fmt + vet + lint + test + build |
| `gofmt -w .` | Formata |

## Comandos que NÃO pode (perigoso)

- ❌ **Rodar `cts cut`/`cts purge` destrutivo contra a máquina real durante dev/teste.** Eles removem arquivos do usuário. Em dev, só com dry-run, ou contra um diretório de teste.
- ❌ **Testar contra caminhos reais** (`~/.claude`, `~/.agents`, `~/.codex`...). Todo teste usa `t.TempDir()`. Um teste que apaga não pode tocar no home de verdade.
- ❌ **Commitar `cts.exe`** ou qualquer binário (está no `.gitignore`).
- ❌ **Commitar com o gate vermelho.** Teste/lint/build têm que passar antes.
- ❌ **`git push --force`, rebase destrutivo, reset --hard** sem motivo claro.

## Modelo de segurança da ferramenta

1. **Dry-run é o padrão.** Sem flag explícita, o cts só mostra — não remove.
2. **Confirmação** antes de qualquer remoção real.
3. **Backup** em `.cts-backups/` antes de apagar.
4. **Confira o alvo** antes de remover. Se o conteúdo contradiz o que foi descrito, pare e avise.

## Workflow (loop XP)

1. **Plan** — menor fatia que entrega valor. Um scanner/feature por vez.
2. **Test** — escreve o teste table-driven primeiro (`t.TempDir()`), confirma o comportamento.
3. **Implement** — código mínimo pra passar.
4. **Refactor** — limpa com teste verde.
5. **Gate** — `./scripts/check.sh` verde.
6. **Commit** — Conventional Commits, sem co-author de agente. Pausa pra revisão.

## Git

- Branch por feature (`feat/agents-scanner`), nunca direto na `main`.
- Conventional Commits: `feat:`, `fix:`, `refactor:`, `docs:`, `chore:`, `test:`.
- Mensagem no que foi feito; corpo só quando o "porquê" não é óbvio.
- Sem `Co-Authored-By` de agente.

## Convenções de código

Go idiomático — ver `~/.claude/go-conventions.md`. Resumo: sem herança (composição), interface pequena definida no consumidor, erro como valor com `%w`, pacote pequeno por capacidade, sem `util` genérico, arquivo pequeno e coeso.
