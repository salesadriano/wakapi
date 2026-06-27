# Deploy do Wakapi em Docker Swarm via Portainer

Guia para publicar o Wakapi (com PostgreSQL) em um cluster **Docker Swarm**
orquestrado pelo **Portainer**, usando o stack [`stack.yml`](../../stack.yml) e o
template [`.env.portainer.example`](../../.env.portainer.example).

## 1. Pré-requisitos

- Cluster Docker Swarm inicializado (`docker swarm init` no nó manager).
- Portainer com acesso ao endpoint do Swarm.
- Imagem publicada no GHCR pelo pipeline (`ghcr.io/<org>/wakapi`). Se o pacote for
  **privado**, cadastre as credenciais em **Portainer → Registries** (não no `.env`).

## 2. Variáveis de ambiente

Use [`.env.portainer.example`](../../.env.portainer.example) como referência. As
**obrigatórias** (sem default — o deploy falha se faltarem):

| Variável | Descrição |
|---|---|
| `WAKAPI_DB_PASSWORD` | senha do PostgreSQL |
| `WAKAPI_PASSWORD_SALT` | *pepper* do hash de senhas (definir uma vez e preservar) |

Recomendadas para produção:

| Variável | Por quê |
|---|---|
| `WAKAPI_PUBLIC_URL` | links absolutos corretos (ex.: reset de senha) |
| `WAKAPI_SESSION_KEY` | mantém sessões/flash entre restarts e re-deploys |
| `WAKAPI_INSECURE_COOKIES` | `true` atrás de proxy TLS; `false` se HTTPS direto |
| `WAKAPI_IMAGE_TAG` | use tag fixa (release ou `sha-<short>`), não `latest`/`develop` |

Gere segredos com `openssl rand -hex 32`.

## 3. Deploy

### Opção A — Web editor
1. Portainer → **Stacks → Add stack**.
2. Cole o conteúdo de `stack.yml` no editor.
3. Em **Environment variables**, adicione as variáveis (copie do `.env.portainer.example`).
4. **Deploy the stack**.

### Opção B — Repositório Git
1. Portainer → **Stacks → Add stack → Git repository**.
2. Repository URL: o repositório; **Compose path**: `stack.yml`.
3. Informe as variáveis em **Environment variables** (ou aponte um arquivo `.env`).
4. **Deploy the stack**. Habilite *automatic updates* (polling/webhook) se quiser
   re-deploy automático a cada push.

## 4. Comportamento operacional

- **Réplicas:** o serviço `wakapi` roda agendadores em processo (agregação,
  relatórios, housekeeping) e por isso é mantido em `replicas: 1`. **Não escale
  acima de 1** — escale o banco, se necessário.
- **Healthchecks:** `wakapi` usa o binário interno `/app/healthcheck` (com
  `start_period` de 120s para acomodar migrações no 1º boot); `db` usa `pg_isready`.
- **Update/rollback:** atualização `start-first` com `failure_action: rollback`
  (monitor 30s) — se a nova task não ficar *healthy*, o Swarm reverte.
- **Persistência:** volumes `wakapi-data` (`/data`) e `wakapi-db-data`
  (`/var/lib/postgresql/data`). O `db` é fixado em nó **manager**
  (`node.role == manager`) para sempre reencontrar seu volume local.

## 5. Endurecimento com Docker Secrets (opcional, recomendado)

O Wakapi lê qualquer variável `WAKAPI_*` também via arquivo, usando o sufixo
`_FILE`. Para não expor senhas como env, crie secrets no Swarm e ajuste o stack:

```bash
printf '%s' 'senha-forte'   | docker secret create wakapi_db_password -
printf '%s' 'salt-aleatorio' | docker secret create wakapi_password_salt -
```

```yaml
# trecho do serviço wakapi
    environment:
      WAKAPI_DB_PASSWORD_FILE: /run/secrets/wakapi_db_password
      WAKAPI_PASSWORD_SALT_FILE: /run/secrets/wakapi_password_salt
    secrets:
      - wakapi_db_password
      - wakapi_password_salt

# o serviço db consome o mesmo secret de senha
    environment:
      POSTGRES_PASSWORD_FILE: /run/secrets/wakapi_db_password
    secrets:
      - wakapi_db_password

# topo do arquivo
secrets:
  wakapi_db_password:
    external: true
  wakapi_password_salt:
    external: true
```

## 6. Reverse proxy (Traefik)

O `stack.yml` traz um bloco `deploy.labels` comentado para o provider Swarm do
Traefik. Descomente, ajuste o `Host(...)` e conecte o serviço à rede externa do
proxy (ex.: `traefik-public`).

## 7. Backup

- **Banco:** `docker exec <task-db> pg_dump -U wakapi wakapi > backup.sql`
  (ou agende um serviço de dump no volume `wakapi-db-data`).
- **Volumes:** snapshot de `wakapi-db-data` e `wakapi-data`.

## 8. Troubleshooting

| Sintoma | Causa provável |
|---|---|
| Deploy falha citando `WAKAPI_DB_PASSWORD is required` | variável obrigatória não informada |
| `wakapi` reinicia em loop | banco indisponível — confira o healthcheck do `db` e a senha |
| Links de e-mail/reset errados | `WAKAPI_PUBLIC_URL` incorreto |
| Sessão/flash some a cada deploy | `WAKAPI_SESSION_KEY` vazio |
| 1º boot demora | migrações em andamento — o `start_period` de 120s cobre isso |
