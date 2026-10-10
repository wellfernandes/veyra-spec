# VeyraSpec

> Especificação e validação comportamental de sistemas de software baseada em modelos.

![Status: Em desenvolvimento](https://img.shields.io/badge/Status-Em%20desenvolvimento-yellow)
![Language: Go](https://img.shields.io/badge/Language-Go-blue)

**Idiomas:** [English](README.md) | Português do Brasil

## O que é o VeyraSpec?

VeyraSpec é uma ferramenta escrita em Go para definir e validar o comportamento esperado de sistemas de software.

Sua ideia central é simples:

> **Especifique como um sistema deve se comportar, observe o que realmente aconteceu e valide se a execução seguiu a especificação.**

O VeyraSpec representa o comportamento esperado como um modelo comportamental e o compara com um trace observado da execução.

O modelo inicial é baseado em uma máquina de estados finitos (FSM):

```text
State + Event -> Transition -> Next State
```

Isso oferece uma base determinística para a validação comportamental, mantendo o núcleo independente de tecnologias externas e das fontes dos traces.

## Por que o VeyraSpec?

Testes automatizados tradicionais são essenciais, mas muitas vezes se concentram em comportamentos isolados:

* Uma função retornou o valor esperado?
* Uma API retornou HTTP 200?
* Um componente específico produziu a saída esperada?

Essas verificações nem sempre explicam se uma execução inteira seguiu o caminho comportamental esperado.

O VeyraSpec se concentra em perguntas como:

* Em qual estado a execução deveria estar?
* Qual evento foi observado?
* Qual transição era esperada?
* Onde a execução observada divergiu da especificação?
* Qual foi a primeira divergência significativa?
* Qual caminho levou à violação?

O objetivo não é substituir os testes automatizados, mas oferecer uma camada adicional de **validação comportamental** que possa complementá-los.

## Como funciona

O VeyraSpec trabalha com dois conceitos fundamentais:

### Specification

Uma `Specification` descreve o comportamento esperado de um sistema.

Ela define o modelo comportamental usado durante a validação, incluindo seu estado inicial, estados, eventos e transições válidas.

Uma Specification representa:

> **O que deve acontecer.**

### Trace

Um `Trace` representa uma sequência ordenada de eventos observáveis registrados durante uma execução.

Um Trace representa:

> **O que realmente aconteceu.**

### Validação

O Validation Engine compara o Trace observado com a Specification.

Conceitualmente:

```text
Specification                Trace
(comportamento esperado)     (comportamento observado)
       │                           │
       └─────────────┬─────────────┘
                     ▼
                 Validation
                     │
            ┌────────┴────────┐
            ▼                 ▼
          Valid            Violation
```

Quando a execução não segue uma transição válida, o VeyraSpec produz um resultado estruturado de validação com informações para diagnosticar a divergência.

## Exemplo

Uma especificação simples de pagamento pode ser representada assim:

```yaml
id: payment-flow
name: payment-flow

initial_state: pending
final_states:
  - approved
  - rejected

states:
  - id: pending
    name: Pending
  - id: approved
    name: Approved
  - id: rejected
    name: Rejected

events:
  - id: payment_approved
    name: Payment Approved
  - id: payment_rejected
    name: Payment Rejected

transitions:
  - id: approve-payment
    from: pending
    event: payment_approved
    to: approved

  - id: reject-payment
    from: pending
    event: payment_rejected
    to: rejected
```

O modelo define que um pagamento no estado `pending` pode passar para `approved` ou `rejected`, dependendo do evento observado.

Por exemplo:

```text
pending
   |
   | payment_approved
   v
approved
```

Se o Trace observado contiver:

```text
payment_approved
```

a execução segue o caminho comportamental definido.

Se o Trace contiver:

```text
payment_cancelled
```

enquanto o estado atual for `pending`, não haverá uma transição válida para esse evento.

O VeyraSpec pode relatar uma violação estruturada como:

```text
Violação

Estado atual: pending
Evento observado: payment_cancelled

Eventos esperados:
- payment_approved
- payment_rejected

Posição no Trace: 1
```

O objetivo não é simplesmente informar:

```text
FALHA
```

mas explicar **onde e por que a execução divergiu do comportamento esperado**.

## Modelo de domínio principal

O modelo comportamental do VeyraSpec é composto por:

```text
Specification
    │
    ├── Initial State
    ├── States
    ├── Events
    └── Transitions
             │
             ▼
          Validation
             ▲
             │
           Trace
             │
             ▼
         Violation
```

A relação fundamental é:

```text
State + Event -> Transition -> State
```

`Action` também faz parte do vocabulário do domínio e representa uma operação realizada pelo sistema. Sua relação exata com transições e validação permanece em aberto para refinamento futuro.

## Princípios de arquitetura

O VeyraSpec está sendo desenvolvido com uma abordagem que prioriza o domínio.

O motor de validação deve permanecer independente de:

* frameworks HTTP;
* bancos de dados;
* brokers de mensagens;
* acesso ao sistema de arquivos;
* parsing de YAML;
* plataformas de observabilidade;
* serviços externos;
* preocupações de infraestrutura.

Representações externas são convertidas para o modelo interno de domínio por meio de adapters.

Por exemplo:

```text
YAML
  |
  v
Specification Adapter
  |
  v
Domain Specification
  |
  v
Validation Engine
```

Da mesma forma, as fontes de traces são normalizadas antes de chegarem ao motor de validação:

```text
External Trace Source
        |
        v
      Adapter
        |
        v
   Domain Trace
        |
        v
Validation Engine
```

Isso mantém o núcleo de validação determinístico, testável e independente da tecnologia usada para obter os dados.

## Requisitos e dependências

* **Go 1.27.2 ou superior** é necessário para compilar e executar o VeyraSpec.
* **[`gopkg.in/yaml.v3`](https://pkg.go.dev/gopkg.in/yaml.v3) v3.0.1** faz o parsing e a decodificação de Specifications e Traces em YAML nos adapters. O domínio e o motor de validação não dependem dessa biblioteca.

Todas as outras dependências fazem parte da biblioteca padrão do Go.

## Validação pela linha de comando

O YAML de Trace exige uma sequência ordenada `events`. Cada item deve ter um ID `event` e pode incluir um timestamp; `completeness` é opcional e, quando omitido, assume `unknown`. Timestamps são mantidos como metadados e não alteram a ordem da validação. Veja um exemplo completo em `examples/payment-trace-valid.yaml`.

Valide um Trace em relação a uma Specification com:

```bash
go run ./cmd/veyra-spec validate examples/payment-spec.yaml examples/payment-trace-valid.yaml
```

Adicione `--json` após `validate` para obter uma saída legível por máquinas:

```bash
go run ./cmd/veyra-spec validate --json examples/payment-spec.yaml examples/payment-trace-valid.yaml
```

O comando termina com status `0` quando o Trace é consistente com a Specification. Ele termina com status diferente de zero quando o Trace viola uma transição, não atende a um requisito de conclusão definido ou não pode ser carregado ou validado. Um Trace parcial ou com completude desconhecida, desde que consistente, não é tratado como falha apenas por não estar completo.

Quando a validação falha, o relatório inclui o caminho das transições válidas que levou à divergência do evento ou à falha de conclusão. A saída JSON representa cada caminho como um array ordenado de transições.

## Documentação

A documentação do projeto está dividida em duas áreas complementares:

```text
                         DOCUMENTAÇÃO
                              │
               ┌──────────────┴──────────────┐
               │                             │
           CONCEITOS                         ADRs
               │                             │
          "O que é?"                "Por que decidimos?"
               │                             │
               ▼                             ▼
        domain-model.md              0001-domain-model.md
        state.md                     0002-specification-format.md
        event.md                     0003-trace-model.md
        transition.md                0004-validation-engine.md
        trace.md
        violation.md
        specification.md
        action.md
```

> **Nota:** A lista de ADRs acima representa as decisões arquiteturais atualmente documentadas no repositório.

### Conceitos

O diretório `docs/concepts/` descreve o significado e a semântica dos conceitos do domínio.

| Documento         | Finalidade                                                         |
| ----------------- | ------------------------------------------------------------------ |
| `domain-model.md`  | Visão geral do modelo de domínio e das relações entre os conceitos |
| `specification.md` | Define o modelo comportamental esperado                            |
| `state.md`         | Define os estados comportamentais                                  |
| `event.md`         | Define os eventos observáveis                                      |
| `transition.md`    | Define transições válidas entre estados                            |
| `trace.md`         | Define traces de execução observados                               |
| `violation.md`     | Define divergências comportamentais                                |
| `action.md`        | Define ações do sistema e sua distinção em relação a eventos       |

### Registros de Decisão de Arquitetura

O diretório `docs/architecture/adr/` registra decisões arquiteturais importantes e suas justificativas.

| Documento                       | Decisão                                              |
| ------------------------------- | ---------------------------------------------------- |
| `0001-domain-model.md`         | FSM como fundamento conceitual do domínio      |
| `0002-specification-format.md` | YAML como formato inicial de Specification     |
| `0003-trace-model.md`          | Modelo interno de Trace independente da fonte  |
| `0004-validation-engine.md`    | Responsabilidades e fluxo do Validation Engine |

A distinção é intencional:

> **Conceitos explicam o significado do sistema. ADRs explicam por que as decisões arquiteturais foram tomadas.**

## Modelo de validação

O Validation Engine verifica cada evento observado em relação ao estado atual e segue as transições correspondentes. Ele relata o primeiro evento que não pode seguir a Specification, incluindo o estado atual, os eventos esperados, a posição no Trace e o caminho válido percorrido até a divergência.

A completude do Trace é explícita. Um Trace parcial pode ser consistente enquanto a execução permanece em andamento. Completude desconhecida significa que a fonte não consegue determinar se haverá mais eventos. Um Trace completo é verificado em relação aos estados finais aceitos pela Specification quando eles estiverem definidos; caso contrário, o motor não presume um requisito de conclusão.

O motor de validação central opera sobre objetos de domínio e permanece independente de formatos externos e infraestrutura. Adapters convertem Specifications e fontes externas de Trace nesses objetos de domínio.

## Status do projeto

O VeyraSpec está atualmente **em desenvolvimento**.

O projeto está concentrado em estabelecer um modelo de domínio bem definido e um motor de validação determinístico antes de introduzir infraestrutura ou integrações adicionais.

A arquitetura e os conceitos de domínio podem evoluir conforme o projeto é implementado e testado.

## Filosofia do projeto

O VeyraSpec segue alguns princípios fundamentais:

* **Domínio em primeiro lugar** — definir o modelo comportamental antes de introduzir infraestrutura.
* **Comportamento explícito** — o comportamento esperado deve ser representado explicitamente.
* **Validação determinística** — a mesma Specification e o mesmo Trace devem produzir o mesmo resultado.
* **Independência tecnológica** — o núcleo não deve depender de tecnologias externas.
* **Diagnósticos úteis** — a validação deve explicar as violações, não apenas informar uma falha.
* **Evolução incremental** — a complexidade deve ser introduzida apenas quando o domínio exigir.

## Licença

Este projeto é open source. Consulte [LICENSE](LICENSE) para mais detalhes.
