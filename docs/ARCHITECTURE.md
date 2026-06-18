# Arquitetura — cts

Documento de engenharia. Leia ao mexer em design, adicionar scanner, ou entender o fluxo.

## Visão

`cts` é uma CLI que **escaneia** categorias de lixo na máquina e (em breve) **remove** com segurança. O fluxo é simples e linear:

```
main → scan.Run(scanners...) → []target.Target → relatório
                                              (depois) → cut/purge → backup → remoção
```

## Princípio: módulos profundos, domínio no centro

- **`internal/target`** — domínio puro. Define `Target` (o que foi achado) e `Category`. Não importa nada de IO. É o centro estável; tudo aponta pra cá.
- **`internal/scan`** — coordena. Define a interface `Scanner` (a **costura**) e `Run`, que roda todos os scanners e acumula resultado/erro.
- **`internal/scan/<categoria>`** — um *adapter* por categoria (`skills`, depois `agents`, `plugins`, `mcp`). Cada um sabe varrer o seu canto e decidir o que está morto. Esconde essa lógica atrás de `Scan(ctx)`.
- **`main`** — monta os scanners com os caminhos reais (`~/.claude/skills`, etc.) e imprime. IO e wiring vivem na borda.

Dependências apontam pra dentro: `skills → target`, `scan → target`, `main → scan, skills, target`. `target` não importa ninguém.

## A costura: `Scanner`

```go
type Scanner interface {
	Category() target.Category
	Scan(ctx context.Context) ([]target.Target, error)
}
```

Interface pequena de propósito (um método de trabalho). Em Go, o adapter **não declara** que implementa — basta ter os métodos. Sem herança, sem `implements`: composição e satisfação implícita.

Por que uma interface aqui e não código direto? Porque há **4 adapters** reais vindo (skills, agents, plugins, mcp). Costura justificada por variação real, não hipotética. (Se fosse 1 só, seria abstração prematura — não criaríamos.)

## Como adicionar um scanner novo

1. Crie `internal/scan/<categoria>/<categoria>.go` com um `Scanner` struct.
2. Implemente `Category()` e `Scan(ctx) ([]target.Target, error)`.
3. Esconda a regra de "está morto?" numa função privada (`inspect`-style) — o chamador não precisa saber como decide.
4. Escreva o teste **table-driven** primeiro, usando `t.TempDir()` (nunca caminho real).
5. Registre no `main` (`scan.Run(ctx, ..., novacategoria.New(path))`).

## Erros

Erro é valor. `Scan` devolve `error` embrulhado com contexto (`fmt.Errorf("...: %w", err)`). `scan.Run` **acumula** erros com `errors.Join` e segue — um scanner quebrado não derruba os outros. Diretório inexistente **não é erro** (só não há o que limpar).

## Segurança (quando `cut`/`purge` chegarem)

- Dry-run por padrão; remoção só com flag explícita + confirmação.
- Backup em `.cts-backups/` antes de apagar.
- A remoção será outro módulo (`internal/remove`) com a mesma disciplina: interface pequena, testável com `t.TempDir()`.

## Limitações conhecidas

- Tamanho de skill que é **symlink** aparece como `0B` (`filepath.WalkDir` não segue symlink). A resolver: seguir o alvo do link para medir.
