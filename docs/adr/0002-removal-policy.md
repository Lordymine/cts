# ADR 0002 — Política de remoção: morto automático, ativo explícito, keepers protegidos

- **Status:** aceito
- **Data:** incremento de remoção

## Contexto

O cts lida com duas coisas distintas:
1. **Morto / órfão / quebrado** — config sem binário, symlink quebrado. Lixo comprovado; o cts tem certeza de que é seguro remover.
2. **Ativo / instalado mas indesejado** — ex.: `qwen` instalado e funcionando que o usuário não quer mais. O cts **não tem como saber sozinho** que um item ativo é indesejado — isso é julgamento do usuário.

Tratar os dois igual seria perigoso (apagar algo que funciona) ou inútil (só limpar lixo óbvio, sem replicar a faxina real que removeu agentes ativos).

## Decisão

- **`scan`** mostra tudo: mortos (marcados `✗`) + ativos (inventário).
- **`purge`** (lote automático) remove **só os mortos** (`Dead == true`). Nunca toca em ativo sozinho.
- **Ativos só por seleção explícita.** Item ativo (não quebrado) nunca entra no `purge`. Para removê-lo, o usuário **seleciona na lista** (`cut <nome>` ou seleção interativa) — escolha deliberada, item a item, com confirmação.
- **Sem lista de "protegidos" hardcoded.** O cts não tem como saber quais agentes são "keepers" — isso é conhecimento do usuário, frágil e arbitrário. A proteção do ativo vem de ser **seleção-only**; o **backup** é o undo universal contra acidente.
- **Dry-run é o default** em qualquer remoção; **backup** em `.cts-backups/` antes de apagar.

## Por que (os 3 critérios)

- **Difícil de reverter:** define o contrato de segurança de toda remoção futura.
- **Surpreendente sem contexto:** alguém poderia esperar que `purge` apagasse tudo que `scan` listou; aqui `purge` é só os mortos, de propósito.
- **Trade-off real:** poder (remover ativo) vs. segurança (não apagar o que funciona sem o usuário pedir). Resolvido separando a **iniciativa**: o cts age sozinho só no lixo; o resto é escolha explícita.

## Consequências

- A camada `internal/remove` **não tem lista de protegidos**: `purge` filtra `Dead == true`; ativo só sai por nome/seleção explícita.
- **Backup e dry-run são responsabilidade da camada de remoção**, não dos scanners — os scanners continuam read-only. O backup é o que protege contra remoção acidental de um ativo.
