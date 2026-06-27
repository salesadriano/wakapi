# Design — Armazenamento de API keys com hash

- **Status:** proposto (não implementado)
- **Data:** 2026-06-27
- **Motivação:** as API keys hoje são persistidas em texto plano. Um atacante com
  acesso ao banco obtém todas as chaves válidas. Este documento descreve a
  migração para armazenamento com hash, preservando a compatibilidade com as
  chaves já configuradas pelos usuários nos editores.

> **Por que um PR dedicado:** é uma mudança *security-critical*, *breaking* de UX
> e que envolve migração de **primary key** em SQLite/MySQL/Postgres. Deve ser
> implementada e validada com o toolchain Go disponível (compilar + `go test` +
> testes de migração da matriz de bancos do CI). Não foi implementada na entrega
> de médio prazo justamente por não ser verificável no ambiente atual.

## 1. Estado atual

Existem **dois** mecanismos de API key, ambos em texto plano:

1. `User.ApiKey` — coluna única em `users` ([models/user.go:31](../../models/user.go#L31)).
   Gerada como UUID v4 em `CreateUser` e `ResetApiKey` ([services/user.go:277,334](../../services/user.go#L277)).
   Exibida na settings page para o usuário copiar.
2. Tabela `api_keys` — chaves nomeadas adicionais, com a **própria chave como
   primary key** ([models/api_key.go:4](../../models/api_key.go#L4)).

Lookup de autenticação ([services/user.go:105](../../services/user.go#L105) `GetUserByKey`):
primeiro `FindOne(User{ApiKey: key})`, depois `apiKeyService.GetByApiKey(key)`.

## 2. Estratégia

Hash determinístico (necessário para lookup): **HMAC-SHA-256** com uma chave de
servidor (reusar/derivar de `security.password_salt` ou nova `security.api_key_pepper`).
SHA-256 simples também serve (a chave já tem entropia de UUID), mas HMAC com pepper
protege contra ataque de dicionário caso o banco vaze sem o pepper.

> Importante: como o lookup é por igualdade do hash, é hash determinístico — **não**
> use Argon2id (que é por design não-determinístico/lento e serve para senhas).

### Compatibilidade
As chaves existentes nos editores continuam funcionando: o lookup passa a
`hash(chave_recebida)` e compara com o hash armazenado. **Não** invalida chaves.
O que muda é que a chave deixa de ser legível no banco/UI.

### Impacto de UX (a parte breaking)
A settings page hoje exibe a `User.ApiKey` a qualquer momento. Com hash, só é
possível exibir a chave **uma vez**, no momento de criação/reset. Necessário:
- fluxo "copie agora, não será exibida novamente" no reset/registro;
- para a tabela `api_keys`, mostrar a chave só no retorno de `Insert`.

## 3. Mudanças por camada

### Modelo
- `User`: adicionar `ApiKeyHash string gorm:"uniqueIndex;default:null"`. Manter
  `ApiKey` transiente (`gorm:"-"`) apenas para transportar a chave em claro até a
  resposta HTTP no momento da criação/reset.
- `ApiKey`: trocar a PK. Adicionar `KeyHash string gorm:"primaryKey"` (ou um `id`
  surrogate + `KeyHash uniqueIndex`) e tornar `ApiKey` transiente.

### Migração (`migrations/`)
- Nova migration pré-AutoMigrate que:
  1. cria as colunas de hash;
  2. preenche `api_key_hash`/`key_hash` = `hash(valor_em_claro)` para todas as
     linhas existentes;
  3. (após validação) remove as colunas em claro;
  4. para `api_keys`, recriar a PK exige cuidado por dialeto (SQLite recria a
     tabela; Postgres/MySQL `ALTER TABLE ... DROP/ADD CONSTRAINT`). Validar na
     matriz `sqlite/postgres/mysql/mariadb` do CI.

### Repositórios
- `UserRepository.FindOne` no caminho de auth: lookup por `ApiKeyHash`.
  Ajustar `GetUserByKey` ([services/user.go:114](../../services/user.go#L114)) para
  `FindOne(User{ApiKeyHash: hash(key)})`.
- `ApiKeyRepository` ([repositories/api_key.go](../../repositories/api_key.go)):
  `GetByApiKey`/`Delete` passam a filtrar por `key_hash = hash(key)`; `Insert`
  calcula o hash antes de persistir.

### Serviços
- `services/user.go`: `CreateUser`/`ResetApiKey` geram a chave (UUID), calculam o
  hash, persistem o hash e retornam a chave em claro **uma vez**.
- `services/api_key.go`: idem para chaves nomeadas; `GetByApiKey` hashifica a
  entrada.
- Invalidação de cache: `GetUserByKey` usa cache keyed pela chave em claro — manter
  (a chave em claro chega na request); ok.

### Rotas / Views
- [routes/settings.go](../../routes/settings.go) + templates de settings: fluxo de
  exibição única (banner com a chave logo após criar/resetar).
- `models/view/settings.go` / `common.go`: não embutir a chave em claro nas views
  fora do fluxo de criação.

## 4. Testes
- Unit: `hash()` determinístico; `GetByApiKey` casa por hash; `requireFullAccessKey`.
- Migração: rodar a matriz de bancos do CI com `--migration` validando o backfill e
  a troca de PK em `api_keys`.
- E2E (bruno): autenticar com uma chave pré-existente após a migração (garantia de
  não-regressão de compatibilidade).

## 5. Riscos e rollback
- **Risco alto:** erro no backfill/PK trava autenticação de todos os usuários.
  Mitigação: migration idempotente, manter colunas em claro até uma release
  seguinte (remoção em duas fases), e snapshot/backup antes do deploy.
- **Rollback:** enquanto as colunas em claro existirem (fase 1), reverter é trivial.
  Após a remoção (fase 2), rollback exige restore de backup.

## 6. Sequência sugerida de PRs
1. PR-1: colunas de hash + backfill + lookup por hash (mantendo colunas em claro) +
   fluxo "exibir uma vez". Compatível e reversível.
2. PR-2 (release seguinte): remoção das colunas em claro.
