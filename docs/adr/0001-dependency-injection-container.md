# ADR-0001 — Injeção de dependências via container explícito

- **Status:** proposto
- **Data:** 2026-06-27
- **Contexto:** [main.go:95](../../main.go#L95) já registra `// TODO: Refactor entire project to be structured after business domains`.

## Contexto

Hoje a composição da aplicação usa **variáveis globais de pacote** em `main.go`
([main.go:52-93](../../main.go#L52)): 2 globais de infra (`db`, `config`), 13
repositórios e 19 serviços. O wiring (~linhas 116-348) instancia tudo em sequência
e registra handlers.

**Fato verificado:** esses globais são `package main` (identificadores minúsculos)
e, portanto, **só são acessíveis dentro de `main.go`**. (As ocorrências de
`userService`, `summaryService` etc. em outros pacotes são campos/parâmetros locais
homônimos, não os globais.) Ou seja, o blast radius desta refatoração está
**confinado ao pacote `main`** — os serviços/repos já recebem dependências por
construtor (`NewXxx(deps...)`).

### Problemas
- Estado global dificulta testar a inicialização e força ordem implícita de boot.
- Acoplamento e ordem de wiring espalhados num único arquivo grande.
- Sem ponto único para montar um subconjunto da aplicação (ex.: para testes E2E
  in-process ou para os scripts em `scripts/`).

## Decisão

Introduzir um **container de composição** explícito que constrói e segura as
dependências, substituindo as globais. Como o escopo é só `package main`, a
mudança não toca os demais pacotes.

### Esboço

```go
// app/container.go  (novo pacote, sem estado global)
package app

type Container struct {
    Config *config.Config
    DB     *gorm.DB

    UserRepository      repositories.IUserRepository
    HeartbeatRepository repositories.IHeartbeatRepository
    // ... demais repos

    UserService      services.IUserService
    HeartbeatService services.IHeartbeatService
    // ... demais serviços
}

func NewContainer(cfg *config.Config, db *gorm.DB) *Container {
    c := &Container{Config: cfg, DB: db}
    c.UserRepository = repositories.NewUserRepository(db)
    // ... constrói repos
    c.UserService = services.NewUserService(/* deps a partir de c */)
    // ... constrói serviços na ordem de dependência
    return c
}
```

`main.go` passa a:
```go
container := app.NewContainer(config, db)
// handlers recebem container.XxxService em vez das globais
```

## Plano incremental (sem big-bang)

1. **Fase 1** — criar `app.Container` + `NewContainer` replicando exatamente o
   wiring atual. `main.go` instancia o container e **adapta as globais para apontar
   para os campos do container** (`userService = container.UserService`). Zero
   mudança de comportamento; permite validar via `go build` + suíte atual.
2. **Fase 2** — remover as globais; handlers e `StartJobs` passam a referenciar
   `container.X` diretamente.
3. **Fase 3** (opcional) — agrupar o container por domínio (ex.: `UserModule`,
   `HeartbeatModule`) alinhado ao TODO de `main.go:95`.

## Critérios de aceite
- `go build ./...` e `go test ./...` verdes.
- Nenhuma variável global de serviço/repo remanescente em `package main` (fim da fase 2).
- Sem mudança de comportamento observável (mesmas rotas, mesmos jobs).

## Riscos
- **Médio/baixo:** ordem de construção incorreta → falha de boot, detectada no
  build/startup do CI (não é falha silenciosa). Mitigado pela Fase 1 (paridade 1:1).

## Por que não foi implementado nesta entrega
O wiring tem ~230 linhas com ordem e assinaturas de construtor específicas;
reescrevê-lo **sem compilador disponível neste ambiente** tornaria a verificação
impossível. Deve ser executado com toolchain Go (build + testes). A fatia da Fase 1
é pequena e mecânica quando há `go build` para validar.
