# CLAUDE.md

Instruções para o Claude Code (e demais agents) ao trabalhar neste repositório (**Wakapi**, `github.com/muety/wakapi`).

Este arquivo é o ponto de entrada. O protocolo completo, agnóstico a linguagem, vive em [.github/agents/AGENTS.md](.github/agents/AGENTS.md) e é **obrigatório**.

---

## Stack detectada (baseline deste projeto)

| Camada | Tecnologia |
|---|---|
| Backend | **Go 1.26** (`go.mod` → `github.com/muety/wakapi`) |
| Persistência | SQLite / MySQL / Postgres via GORM (ver `migrations/`, `repositories/`) |
| Frontend | HTML templates (`views/`) + Tailwind CSS + JS (`static/assets/`) |
| Build de assets | npm + Tailwind + iconify (`package.json`) |

Camadas de código relevantes: `models/`, `repositories/`, `services/`, `routes/`, `middlewares/`, `migrations/`, `config/`, `utils/`, `helpers/`, `views/`, `static/`.

### Comandos essenciais

```bash
# Backend
go build -o wakapi                                  # compilar
go run .                                            # executar localmente
CGO_ENABLED=0 go test -json -coverprofile=coverage/coverage.out ./... -run ./... | tparse -all   # testes + cobertura
go test ./...                                       # testes simples

# Frontend / assets
npm run build          # build:icons + build:tailwind
npm run watch          # rebuild de assets em mudança de views/js/css
```

> Não há `Makefile`. As tarefas de assets ficam em `package.json` scripts; o ciclo Go usa as ferramentas padrão do `go`.

---

## Protocolo comum obrigatório (resumo operacional)

O detalhamento está em [.github/agents/AGENTS.md](.github/agents/AGENTS.md). Antes de qualquer tarefa:

1. **Carregar** `.github/agents/AGENTS.md` como protocolo comum e ler as memórias:
   - [.github/agents/memoria/MEMORIA-COMPARTILHADA.md](.github/agents/memoria/MEMORIA-COMPARTILHADA.md) (memória geral)
   - [.github/agents/memoria/MEMORIA-PROJETO.md](.github/agents/memoria/MEMORIA-PROJETO.md) (memória deste projeto)
2. **Registrar o prompt** via skill [.github/skills/prompt-logger/](.github/skills/prompt-logger/) em `docs/prompts/`, sanitizando segredos, tokens, credenciais e dados pessoais antes de persistir.
3. **TDD obrigatório** em desenvolvimento, refatoração ou correção: usar [.github/skills/protocolo-tdd/](.github/skills/protocolo-tdd/) como referência (TDD, integração real, E2E real com Cypress quando aplicável).
4. **Registro técnico** da entrega via [.github/skills/review-documentation/](.github/skills/review-documentation/).
5. **Documentação formal e commits são delegados a subagents utilitários:**
   - [documentation-writer.agent.md](.github/agents/documentation-writer.agent.md) — redige registros técnicos, handoffs, reviews, changelogs.
   - [commit-writer.agent.md](.github/agents/commit-writer.agent.md) — gera mensagens de commit semânticas a partir do diff real.
   - O agent originador continua responsável por revisar conteúdo, diff, escopo e segurança.
6. **Atualizar memórias** ao fim conforme o escopo da decisão (geral × projeto), mantendo-as sucintas; detalhes extensos vão para `.github/agents/memoria/historico/`.

### Ciclo de desenvolvimento

```
Senior Developer implementa → documentation-writer (registro técnico)
  → QA Expert valida → (reprovado) volta ao Senior Developer
                     → (aprovado) commit-writer (commit semântico)
  → Tech Lead revisa diff/escopo/segurança e fecha a entrega (PR)
```

Os 8 agents do pacote estão em [.github/agents/](.github/agents/): tech-lead, senior-developer, qa-expert, ux-expert, dba, business-analyst, documentation-writer, commit-writer.

---

## Convenções

- **Idioma de governança:** documentos formais (System Design, PRD, validações QA, pareceres, aprovações, revisões consolidadas, registros técnicos) em **português do Brasil**, salvo indicação explícita em contrário. Código, identificadores, comandos, schemas e payloads permanecem no idioma técnico original.
- **Commits e branches:** convenção semântica de commits + branch naming aderente ao **Gitflow**; entregas formais via Pull Request com label de review. Branch principal: `master`; desenvolvimento em `develop`.
- **Documentação:** Markdown com diagramas Mermaid; manter rastreabilidade com links para arquivos alterados, testes e revisões.
- **Context7 MCP:** quando disponível no workspace, é a fonte preferencial de documentação técnica atualizada. Baseline de instalação versionada em `.vscode/mcp.json` (sem versionar segredos). Ver seção "Context7 MCP no projeto" em AGENTS.md.
- **Comunicação:** durante a execução, reduzir feedbacks visuais e evitar narrar microações; concentrar o detalhamento no encerramento ou handoff.

---

## Skills e templates

- Skills disponíveis em [.github/skills/](.github/skills/). Após detectar a stack, consultar a skill correspondente quando existir. Para Go, não há skill dedicada no pacote — aplicar as boas práticas gerais e o `protocolo-tdd`.
- Templates operacionais (QA, aprovações, system design, Cypress, etc.) em [.github/agents/templates/](.github/agents/templates/) — usar quando o fluxo correspondente for acionado.
