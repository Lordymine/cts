# ADR 0003 — Mecanismos de remoção: arquivo + comando

- **Status:** aceito
- **Data:** incremento de remoção completa

## Contexto

Categorias diferentes "removem" de formas diferentes:
- **skill morta / config órfã / cache de plugin** → apagar arquivo/dir.
- **agente ativo** → desinstalar o pacote (`npm rm -g`, `bun rm -g`, `uv tool uninstall`) **e** apagar a config.
- **MCP** → remover a entrada de config. O CLI suportado `claude mcp remove <name> -s <scope>` faz isso corretamente (lida com escopo). Editar o `~/.claude.json` na mão reformataria o arquivo inteiro (195KB) e arriscaria ordem/estrutura.

## Decisão

Dois mecanismos, ambos descritos no `Target`:
- **`Paths []string`** → arquivos/dirs a apagar.
- **`Uninstall []string`** → comando a rodar **antes** de apagar (ex.: `["npm","rm","-g",pkg]`, `["claude","mcp","remove",name,"-s","user"]`).

O `Remover`: backup das `Paths` → roda `Uninstall` (se houver, via `Runner` injetado) → apaga as `Paths`. Alvo sem `Paths` **e** sem `Uninstall` é pulado.

**Não** editamos JSON direto (`ConfigEdit` descartado) — `claude mcp remove` é o caminho suportado e seguro.

## Escopo / limitações

- **MCP:** só o escopo *user* ganha comando automático. Server de projeto fica inventário (removido dentro do projeto, via cwd).
- **Agente sem manager conhecido** (go-bin, python venv): removido só por arquivo (config). O binário/install fora das `Dirs` não é tocado nesta versão.

## Por que (os 3 critérios)

- **Difícil de reverter:** define como toda remoção acontece.
- **Surpreendente sem contexto:** esperaria-se delete de arquivo pra tudo; comando-pra-alguns é deliberado.
- **Trade-off real:** comando suportado (`claude mcp remove`, `npm rm`) vs. editar config na mão. Escolhemos o suportado — mais seguro, menos frágil.

## Consequências

- `Runner` injetado no `Remover` → testável (fake runner, sem executar nada).
- Backup continua antes de tudo. No uninstall via comando, o backup cobre as `Paths` (config); o pacote reinstala-se pelo próprio manager.
