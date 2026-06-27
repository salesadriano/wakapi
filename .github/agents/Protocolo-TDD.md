### PROTOCOLO DE ENGENHARIA DE TESTES (TDD) — REVISADO
#### Foco em E2E Real, Sem Mocks e Sem Smoke Tests

#### 1. O Ciclo de Desenvolvimento TDD (Red-Green-Refactor)
A prática de TDD exige que os testes automatizados guiem a arquitetura. O fluxo de trabalho diário deve seguir as três fases:
1. **Red:** Escreva o teste baseado nos Critérios de Aceite (BDD) antes de codificar a funcionalidade.
2. **Green:** Escreva o código mínimo necessário para o teste passar na camada correspondente.
3. **Refactor:** Melhore o design do código garantindo que o comportamento externo e as asserções permaneçam intactos.

---

#### 2. Distribuição da Pirâmide de Testes (70/20/10)
A nossa automação será distribuída para maximizar o feedback rápido nas bases e garantir a integração absoluta no topo. 

##### 2.1. Camada Unitária (Backend/Nest.js com Jest) - 70% da Suíte
Foco em regras de negócio puras (Services, Casos de Uso), isoladas de rede e banco de dados.
* **Padrão AAA (Arrange, Act, Assert):** Segregação visual rigorosa entre a organização dos dados, execução e verificação de resultados.
* **Caixa Preta de Unidade:** Teste os resultados e saídas da interface pública das classes, evitando testar a sequência interna de métodos privados.

##### 2.2. Camada de Integração (Backend com Jest + Supertest) - 20% da Suíte
Foco em contratos HTTP (Controllers) e persistência de dados real.
* **Proibição de Mocks na Camada de Dados:** É expressamente proibido simular o banco de dados (usando `jest.fn()`) para testar DAOs ou Repositórios.
* **Testcontainers Obrigatório:** Os testes de integração devem rodar contra instâncias reais e temporárias do PostgreSQL levantadas via *Testcontainers*, garantindo que as queries geradas pelo TypeORM funcionem fisicamente.

##### 2.3. Camada End-to-End (Frontend + Backend + DB integrados via Cypress) - 10% da Suíte
Esta é a última linha de defesa. O Cypress validará fluxos completos de negócio a partir da perspectiva do usuário final, cruzando toda a infraestrutura.
* **Zero Mocks (Proibição de Stubs de Rede):** É terminantemente proibido o uso de ferramentas como `cy.intercept()` para injetar *fixtures* ou mockar respostas do backend no fluxo principal de E2E. O frontend deve se comunicar com um backend real, que por sua vez lê e grava em um banco de dados de teste real.
* **Proibição de "Smoke Tests" Superficiais:** Testes E2E não devem apenas verificar se a página carrega. Eles devem criar jornadas profundas de negócio, cobrindo caminhos felizes, erros de validação, limites de borda (ex: estouro de saldo, limites operacionais) e testes de exaustão para funcionalidades críticas.
* **Múltiplas Asserções por Jornada:** Como testes E2E sem mocks são custosos por natureza, agrupe múltiplas asserções dentro do mesmo fluxo, refletindo o uso real do sistema, em vez de criar dezenas de microtestes E2E isolados que exigem reinicialização completa do ambiente a cada iteração.

---

#### 3. Governança de Dados de Teste e Estado (Crucial para E2E sem Mocks)
Como o E2E agora roda 100% integrado e sem dados falsificados na camada de rede, a gestão do estado do banco de dados é crítica.
* **Criação Explícita de Dados:** Nenhum teste deve depender de dados criados manualmente ou adivinhar IDs. Os dados necessários devem vir do *seeder* inicial padronizado da aplicação ou serem criados programaticamente por testes anteriores dentro do próprio roteiro.
* **Preservação de Rastreabilidade:** As informações geradas por um teste E2E devem permanecer disponíveis para os testes subsequentes do mesmo roteiro, imitando a linha do tempo de um usuário no sistema (ex: um pedido criado no Teste A será faturado no Teste B).
* **Descarte Controlado:** O descarte e a limpeza do banco de dados (teardown) devem ocorrer apenas ao final do roteiro completo de execução, garantindo que o QA possa investigar o estado residual em caso de falha.

---

#### 4. Seletores Resilientes e Prevenção de Flaky Tests
Em um ambiente E2E 100% integrado, a latência de rede e banco de dados real estará presente.
* **Uso Obrigatório de data-cy:** Para evitar que mudanças de design quebrem a automação, a equipe de frontend deve implementar atributos `data-cy` ou `data-test` em todos os elementos interativos. O uso de classes CSS genéricas ou IDs de estilo para testes E2E resultará em reprovação no code review.
* **Isolamento de Flaky Tests:** Testes que falharem de forma intermitente (flaky) devido a condições de corrida em ambiente integrado serão imediatamente colocados em quarentena pelo QA, bloqueando a pipeline até que o desenvolvedor resolva a causa raiz arquitetural.

---

#### 5. Critérios de Aceite para Handoff (DoD para o QA)
Para que eu (QA Expert) aceite a sua entrega para homologação independente, os seguintes itens são **bloqueantes**:
1. O código foi guiado por TDD e passou nas camadas de Unidade e Integração (via Testcontainers) localmente e na CI.
2. A evidência de execução dos testes ocorre obrigatoriamente de forma isolada em um ambiente containerizado, utilizando o comando oficial da arquitetura: `docker compose run --rm <service> test`.
3. Todos os critérios BDD (Given/When/Then) descritos no PRD/Escopo possuem cobertura E2E equivalente rodando contra a base de dados de homologação, provando seu funcionamento fim a fim.
4. Em fluxos frontend, o System Design referencia explicitamente os componentes visuais validados (Design System), garantindo o alinhamento visual requerido.

Com este protocolo, nós tratamos o ambiente E2E com o rigor que ele merece: como um simulador de produção real, e não como uma vitrine de dados estáticos. 

Podem proceder com a aplicação desta norma para as novas implementações. Caso encontrem gargalos de orquestração de banco de dados para a subida do ambiente E2E nos pipelines CI/CD, estou à disposição para revisarmos a estratégia de infraestrutura em conjunto com o DBA.