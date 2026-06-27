# ADR-0002 — Decomposição do "god object" `User`

- **Status:** proposto
- **Data:** 2026-06-27

## Contexto

[models/user.go:29-65](../../models/user.go#L29) define `User` com ~35 campos que
misturam responsabilidades distintas: identidade/auth, preferências, compartilhamento,
integrações, assinatura/billing, relatórios/leaderboard e estado administrativo.

A struct é referenciada em praticamente todo o codebase (serviços, repos, rotas,
views, migrations). Decompor em **tipos separados com tabelas próprias** seria uma
mudança de schema de alto risco e altíssima churn. Esta ADR propõe um caminho
**de baixo risco** que melhora a coesão sem quebrar chamadas nem o schema.

## Agrupamento proposto dos campos

| Grupo | Campos |
|---|---|
| Identidade & Auth | `ID`, `ApiKey`, `Email`, `Password`, `AuthType`, `Sub`, `WebauthnID`, `Credentials`, `ResetToken` |
| Preferências | `Location`, `StartOfWeek`, `HeartbeatsTimeoutSec`, `ExcludeUnknownProjects`, `ReadmeStatsBaseUrl` |
| Compartilhamento | `ShareDataMaxDays`, `ShareEditors`, `ShareLanguages`, `ShareProjects`, `ShareOSs`, `ShareMachines`, `ShareLabels`, `ShareActivityChart` |
| Integrações (relay/import) | `WakatimeApiKey`, `WakatimeApiUrl` |
| Assinatura/Billing | `SubscribedUntil`, `SubscriptionRenewal`, `StripeCustomerId` |
| Relatórios/Leaderboard | `ReportsWeekly`, `PublicLeaderboard`, `UnsubscribeToken` |
| Estado/Admin | `IsAdmin`, `HasData`, `InvitedBy`, `CreatedAt`, `LastLoggedInAt` |

## Decisão

Usar **structs embutidas anônimas** com `gorm:"embedded"`. O GORM achata structs
embutidas nas **mesmas colunas** da tabela `users` → **sem migração de schema**. Com
embedding anônimo, o acesso promovido `user.ShareEditors` continua funcionando →
**sem mudança nos call-sites**. O ganho é organização/coesão e um lar natural para
comportamento (métodos por grupo).

### Esboço

```go
type UserSharing struct {
    ShareDataMaxDays   int  `json:"-"`
    ShareEditors       bool `json:"-" gorm:"default:false; type:bool"`
    ShareLanguages     bool `json:"-" gorm:"default:false; type:bool"`
    // ...
}

type UserSubscription struct {
    SubscribedUntil     *CustomTime `json:"-"`
    SubscriptionRenewal *CustomTime `json:"-"`
    StripeCustomerId    string      `json:"-"`
}

type User struct {
    ID       string `json:"id" gorm:"primary_key"`
    ApiKey   string `json:"api_key" gorm:"unique; default:NULL"`
    // ... campos de identidade

    UserSharing      `gorm:"embedded"`
    UserSubscription `gorm:"embedded"`
    // ... demais grupos
}
```

> Atenção: manter as **tags `gorm` e `column` idênticas** nos campos movidos para
> preservar os nomes de coluna (ex.: `ShareOSs` usa `column:share_oss`). Validar com
> os testes de migração da matriz `sqlite/postgres/mysql/mariadb` que o schema
> resultante é **byte-a-byte equivalente** ao atual.

## Plano incremental

1. **Fase 1** — extrair os grupos como structs embutidas anônimas (sem mudança de
   schema nem de call-site). Validar paridade de schema via `--migration` no CI.
2. **Fase 2** — mover comportamento relacionado para métodos dos novos tipos
   (ex.: `UserSubscription.IsActive()`), reduzindo lógica solta nos serviços.
3. **Fase 3** (opcional, futuro) — só se houver necessidade real, promover algum
   grupo a tabela própria (ex.: billing) com migração expand/contract.

## Critérios de aceite
- `go build ./...` e `go test ./...` verdes.
- Testes de migração (4 bancos) confirmam schema equivalente ao atual.
- Nenhuma alteração de comportamento de API/serialização (JSON tags preservadas).

## Riscos
- **Médio:** divergência sutil de tag/coluna ao mover campos → schema diferente.
  Mitigado pela matriz de migração do CI e revisão de DDL antes/depois.

## Por que não foi implementado nesta entrega
Embora a Fase 1 seja de baixo risco conceitual, a equivalência de schema só é
verificável rodando os **testes de migração nos 4 bancos** — indisponível sem
toolchain neste ambiente. Mover campos preservando todas as tags exige `go build`
para garantir que nada quebrou. Executar com Go disponível.
