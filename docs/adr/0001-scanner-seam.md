# ADR 0001 — `Scanner` como costura entre o CLI e as categorias

- **Status:** aceito
- **Data:** incremento inicial

## Contexto

O cts precisa escanear 4 categorias distintas de lixo (skills, agentes, plugins, MCP), e cada uma vive num lugar diferente e tem regras próprias de "está morto?". O CLI precisa rodar todas e juntar o resultado num relatório único.

## Decisão

Definir uma interface pequena `Scanner` (`Category()` + `Scan(ctx) ([]target.Target, error)`) no pacote `scan`, e implementar **um adapter por categoria** em `internal/scan/<categoria>`. O `scan.Run` recebe os scanners por parâmetro e acumula resultado e erro (`errors.Join`).

## Por que (os 3 critérios de ADR)

- **Difícil de reverter:** a costura define como toda categoria nova é plugada; mudar depois mexe em todos os adapters.
- **Surpreendente sem contexto:** alguém poderia esperar um grande `switch` por categoria no main; a escolha por interface+adapters é deliberada.
- **Trade-off real:** interface+adapters (mais arquivos, extensível, testável isolado) vs. um único arquivo com tudo (menos cerimônia, mas acopla as 4 lógicas e dificulta teste).

## Consequências

- Adicionar categoria = criar um adapter e registrá-lo no `main`. Não toca nos outros.
- Cada adapter é testável isolado com `t.TempDir()`.
- A costura é justificada por **variação real** (4 adapters), não hipotética — não seria criada para 1 só.
