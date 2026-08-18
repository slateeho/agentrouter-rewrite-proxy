# AgentRouter Model Alias Proxy

> A local compatibility gateway that makes AgentRouter usable with a broader range of Claude- and OpenAI-compatible coding agents.

[English](#english) · [Русский](#русский)

---

# English

## Why this project exists

AgentRouter can offer unusually attractive access to expensive coding models, including promotional or discounted access for new users.

The problem is that having:

- an AgentRouter account,
- available credits,
- a valid API key,
- and access to a model

does not automatically mean that every coding agent can use AgentRouter directly.

Two compatibility problems appear repeatedly:

1. **AgentRouter WAF / transport compatibility**
2. **Client-side model recognition**

This proxy handles both locally while keeping the real upstream model and AgentRouter billing unchanged.

---

## 1. AgentRouter is not always a drop-in API endpoint

Many coding agents can speak Anthropic Messages or OpenAI Chat Completions, but AgentRouter has an additional compatibility layer in front of those APIs.

Its WAF can expect a recognized client fingerprint, headers and cookie flow.

As a result, a request can be valid at the API level and still fail before inference:

```text
AgentRouter account         ✅
API key                     ✅
Credits                     ✅
Upstream model available    ✅
Anthropic/OpenAI request    ✅

Direct client connection    ❌
```

The proxy handles that transport layer locally:

```text id="hlxuo0"
┌─────────────────────────────┐
│ Coding Agent / IDE          │
└──────────────┬──────────────┘
               │
               │ Standard Anthropic/OpenAI API
               ▼
┌─────────────────────────────┐
│ Local Compatibility Proxy   │
│ 127.0.0.1:8318              │
│                             │
│ • WAF-compatible transport  │
│ • cookie handling           │
│ • upstream authentication   │
└──────────────┬──────────────┘
               │
               ▼
┌─────────────────────────────┐
│ AgentRouter                 │
└─────────────────────────────┘
```

The coding agent only needs to understand a normal Anthropic- or OpenAI-compatible endpoint.

---

## 2. Some clients do not recognize the upstream model name

Coding agents often treat the model ID as more than an arbitrary API string.

A model name may be used for:

- capability detection,
- model selection,
- context-window defaults,
- timeout profiles,
- fast-model routing,
- subagent selection,
- prompt strategy,
- cache behavior,
- UI presentation.

A perfectly functional upstream model can therefore be rejected or handled incorrectly simply because the client does not recognize its model ID.

The proxy exposes a stable client-facing Claude identity:

```text id="mck2wg"
Client-visible model
claude-haiku-4-5
        │
        ▼
Local Proxy
        │
        │ request rewrite
        ▼
claude-opus-5
        │
        ▼
AgentRouter
```

AgentRouter may return a canonical internal identifier such as:

```text id="0k4ub4"
anthropic/claude-opus-5-ps-aws-dst
```

The proxy rewrites the response back to:

```text id="6z8zkf"
claude-haiku-4-5
```

From the client's perspective, the model identity remains stable throughout the session.

---

## Why use `claude-haiku-4-5` as the client-facing alias?

Claude Haiku 4.5 is a useful compatibility identity because it belongs to Anthropic's fast Claude family and is widely recognizable by Claude-oriented tooling.

For a coding client, that gives a familiar combination:

```text id="37pew8"
known Claude model
      +
Anthropic protocol
      +
fast-model family
      +
coding-oriented usage
      +
recognized model ID
```

Some clients use model-specific profiles internally.

When they recognize `claude-haiku-4-5`, they may select an existing Claude-compatible execution path instead of falling back to an unknown-model path or refusing the model entirely.

This can affect client-side behavior such as:

```text id="djutcr"
timeouts
token budgets
subagent selection
tool configuration
context assumptions
cache settings
UI presentation
```

### Important: the alias does not make Opus 5 faster

The actual execution remains:

```text id="edquq4"
Client identity     = Claude Haiku 4.5
Actual inference    = Claude Opus 5
AgentRouter billing = Claude Opus 5
Actual model speed  = Claude Opus 5
Actual intelligence = Claude Opus 5
```

The alias is a **client compatibility mechanism**, not an inference optimization.

It does not:

```text id="nedtnd"
turn Opus into Haiku
reduce upstream model price
change AgentRouter billing
unlock unavailable models
increase quota
bypass model entitlement
```

If a client has a faster or more efficient code path associated with Haiku, the alias may allow that **client-side path** to be used.

The upstream inference itself is still Opus 5.

---

## Wider coding-agent compatibility

AgentRouter already works with several established coding tools.

The interesting use case for this proxy is the much larger ecosystem of provider-neutral or multi-provider coding agents.

Examples include:

| Agent | Why it is interesting |
|---|---|
| **Qwen Code** | Mature coding CLI, Anthropic/OpenAI support, broad provider selection |
| **Claude Code-compatible backends** | Excellent established agent UX with configurable upstream gateways |
| **Autohand** | Terminal-native and provider-neutral |
| **Reasonix** | Cost- and cache-oriented agent architecture |
| **Jazz** | Lightweight terminal agent with model/provider flexibility |
| **ForgeCode** | Broad provider support and task-specific model selection |
| **Neovate** | Open-source CLI, headless operation, skills and MCP |
| **Pi** | Extremely model-neutral with a large provider/model ecosystem |
| **Hermes** | Broad provider support and reusable agent skills |
| **Goose** | Mature open-source agent with extensive provider and MCP support |
| **Crush** | Strong multi-model terminal harness |
| **Command Code** | Interesting agent UX, but not designed to consume arbitrary AgentRouter credentials directly |

The proxy does not require these projects to implement AgentRouter-specific transport logic.

Where a client can already consume an Anthropic-compatible or OpenAI-compatible endpoint, the proxy can provide the AgentRouter-specific compatibility layer underneath it.

---

## Supported API surfaces

The proxy exposes both major API styles:

### Anthropic Messages

```text id="oawjol"
POST http://127.0.0.1:8318/v1/messages
```

Example client model:

```text id="ow96w5"
claude-haiku-4-5
```

Actual upstream model:

```text id="j2pnij"
claude-opus-5
```

### OpenAI Chat Completions

```text id="dy16cz"
POST http://127.0.0.1:8318/v1/chat/completions
```

The same client alias can be used through the OpenAI-compatible path.

---

## Model discovery

Clients querying:

```text id="xgxm76"
GET /v1/models
```

see the configured compatibility model:

```json id="3ykj6g"
{
  "data": [
    {
      "id": "claude-haiku-4-5",
      "object": "model",
      "owned_by": "agentrouter"
    }
  ],
  "object": "list"
}
```

They do not need to know which AgentRouter model is actually used upstream.

---

## Request and response rewriting

### Client request

```json id="6f4df1"
{
  "model": "claude-haiku-4-5",
  "messages": [
    {
      "role": "user",
      "content": "Fix this Kubernetes manifest"
    }
  ]
}
```

### Proxy sends upstream

```json id="25plln"
{
  "model": "claude-opus-5",
  "messages": [
    {
      "role": "user",
      "content": "Fix this Kubernetes manifest"
    }
  ]
}
```

### AgentRouter may return

```json id="fg8s82"
{
  "model": "anthropic/claude-opus-5-ps-aws-dst"
}
```

### Client receives

```json id="19gqyi"
{
  "model": "claude-haiku-4-5"
}
```

Only model identity metadata is rewritten.

Model output, tool calls, usage data and other response content remain intact.

---

## Anthropic SSE support

Model rewriting also works for streaming Anthropic responses.

Upstream:

```text id="djah13"
event: message_start
data: {"type":"message_start","message":{"model":"anthropic/claude-opus-5-ps-aws-dst"}}
```

Client:

```text id="2u8egt"
event: message_start
data: {"type":"message_start","message":{"model":"claude-haiku-4-5"}}
```

The proxy preserves the surrounding SSE protocol and modifies only the relevant model field.

---

## Prompt caching is preserved

Model aliasing should not interfere with Anthropic prompt caching.

The proxy preserves request content such as:

```text id="z8cqxm"
system prompts
messages
tools
cache_control
thinking blocks
tool calls
usage metadata
```

A typical coding-agent request contains a large stable prefix:

```text id="uu2sd3"
system instructions
repository context
QWEN.md / CLAUDE.md
tool definitions
project conventions
architecture
conversation history
────────────────────────────
small new user request
```

If the upstream AgentRouter channel supports prompt caching, usage may contain:

```json id="dygqxv"
{
  "cache_creation_input_tokens": 12000,
  "cache_read_input_tokens": 11800
}
```

The model alias itself does **not** increase the cache-hit rate.

Cache efficiency still depends on:

- the real upstream provider,
- prompt structure,
- stable prefixes,
- cache-control configuration,
- cache lifetime.

The proxy simply avoids destroying those structures while applying model compatibility rewriting.

---

## Credential isolation

The client does not need the real AgentRouter API key.

A local credential can be used between the coding agent and the proxy:

```text id="j0078h"
Coding Agent
    │
    │ local disposable token
    ▼
Local Proxy
    │
    │ real AgentRouter credential
    ▼
AgentRouter
```

The upstream credential remains inside the local proxy environment.

This is especially useful when experimenting with multiple coding agents or IDE extensions.

---

## Architecture

```text id="irrzle"
┌──────────────────────────────────────────┐
│ Coding Agent / IDE                       │
│                                          │
│ model: claude-haiku-4-5                  │
│ Anthropic or OpenAI-compatible protocol  │
└───────────────────┬──────────────────────┘
                    │
                    ▼
┌──────────────────────────────────────────┐
│ AgentRouter Model Alias Proxy            │
│ 127.0.0.1:8318                           │
│                                          │
│ • local authentication                   │
│ • model advertisement                    │
│ • request model rewriting                │
│ • response model rewriting               │
│ • Anthropic SSE rewriting                │
│ • WAF-compatible transport               │
│ • upstream credential isolation          │
└───────────────────┬──────────────────────┘
                    │
                    │ claude-opus-5
                    ▼
┌──────────────────────────────────────────┐
│ AgentRouter                              │
│                                          │
│ actual model: Claude Opus 5              │
│ actual accounting: Claude Opus 5         │
└──────────────────────────────────────────┘
```

---

## What this project provides

```text id="3h1hiv"
✅ AgentRouter WAF compatibility
✅ Claude model-name compatibility
✅ Anthropic Messages API
✅ OpenAI Chat Completions API
✅ SSE-aware model rewriting
✅ static model advertisement
✅ local credential isolation
✅ stable client-facing model identity
✅ compatibility with more coding-agent frontends
```

---

## What this project does NOT provide

```text id="l043at"
❌ cheaper Opus inference through a Haiku alias
❌ hidden model access
❌ quota bypass
❌ billing bypass
❌ entitlement bypass
❌ unauthorized upstream models
```

AgentRouter always receives the real configured upstream model.

If the proxy maps:

```text id="wcs5rc"
claude-haiku-4-5 → claude-opus-5
```

then AgentRouter executes and accounts for:

```text id="stctiu"
claude-opus-5
```

---

## Why this matters

The coding-agent ecosystem is becoming increasingly provider-neutral.

A good agent frontend should not have to contain special transport code for every inference provider.

This project keeps the integration boundary simple:

```text id="xc3i2u"
Coding Agent
      │
      │ standard Anthropic/OpenAI API
      ▼
Compatibility Proxy
      │
      │ AgentRouter-specific transport
      ▼
AgentRouter
```

That makes AgentRouter easier to use with existing tools while keeping its upstream authentication, model selection and accounting intact.

---

## Sources

- AgentRouter documentation: https://docs.agentrouter.org/
- AgentRouter third-party integration guide: https://co.agentrouter.org/portal/guide
- Anthropic — Claude Haiku 4.5: https://www.anthropic.com/news/claude-haiku-4-5

---

# Русский

## Зачем существует этот проект

AgentRouter может давать очень привлекательный доступ к дорогим coding-моделям, в том числе за счет промо-условий и скидок для новых пользователей.

Но наличие:

- аккаунта AgentRouter,
- доступного баланса,
- действующего API key,
- и разрешенной модели

еще не означает, что любой coding agent сможет напрямую работать с AgentRouter.

На практике возникают две отдельные проблемы совместимости:

1. **AgentRouter WAF / transport compatibility**
2. **Распознавание модели самим клиентом**

Этот proxy решает обе проблемы локально, не изменяя фактическую upstream-модель и ее тарификацию в AgentRouter.

---

## 1. AgentRouter не всегда является обычным drop-in API endpoint

Многие coding agents умеют работать с Anthropic Messages или OpenAI Chat Completions, однако перед API AgentRouter расположен дополнительный слой защиты и совместимости.

WAF может ожидать определенный fingerprint клиента, набор заголовков и cookie flow.

Поэтому запрос может быть полностью корректным с точки зрения API и все равно не дойти до inference:

```text id="lailkp"
AgentRouter account         ✅
API key                     ✅
Credits                     ✅
Upstream model available    ✅
Anthropic/OpenAI request    ✅

Direct client connection    ❌
```

Proxy берет этот transport layer на себя:

```text id="7fjo3l"
┌─────────────────────────────┐
│ Coding Agent / IDE          │
└──────────────┬──────────────┘
               │
               │ Standard Anthropic/OpenAI API
               ▼
┌─────────────────────────────┐
│ Local Compatibility Proxy   │
│ 127.0.0.1:8318              │
│                             │
│ • WAF-compatible transport  │
│ • cookie handling           │
│ • upstream authentication   │
└──────────────┬──────────────┘
               │
               ▼
┌─────────────────────────────┐
│ AgentRouter                 │
└─────────────────────────────┘
```

Самому coding agent достаточно понимать обычный Anthropic- или OpenAI-compatible endpoint.

---

## 2. Клиент может не распознавать upstream Model ID

В современных coding agents имя модели часто является не просто строкой, которая передается в API.

Model ID может использоваться для:

- определения возможностей модели;
- выбора модели;
- настроек context window;
- timeout-профилей;
- fast-model routing;
- выбора subagent;
- prompt strategy;
- cache behavior;
- отображения модели в UI.

Поэтому полностью рабочая upstream-модель может быть отклонена или неправильно обработана только потому, что клиент не знает ее Model ID.

Proxy предоставляет стабильное Claude-имя, которое клиент уже умеет использовать:

```text id="oh7vgr"
Client-visible model
claude-haiku-4-5
        │
        ▼
Local Proxy
        │
        │ request rewrite
        ▼
claude-opus-5
        │
        ▼
AgentRouter
```

AgentRouter при этом может вернуть внутреннее canonicalized имя:

```text id="tsok8e"
anthropic/claude-opus-5-ps-aws-dst
```

Proxy преобразует его обратно в:

```text id="bnm104"
claude-haiku-4-5
```

Для клиента модель остается одной и той же на протяжении всей сессии.

---

## Почему в качестве client alias используется `claude-haiku-4-5`

Claude Haiku 4.5 — удобная compatibility identity, поскольку это хорошо узнаваемая модель из fast-линейки Claude.

Для coding client получается знакомое сочетание:

```text id="1nmuqg"
known Claude model
      +
Anthropic protocol
      +
fast-model family
      +
coding-oriented usage
      +
recognized model ID
```

Некоторые coding agents имеют внутренние model-specific profiles.

Если клиент распознает `claude-haiku-4-5`, он может использовать уже существующий Claude-compatible execution path вместо режима неизвестной модели или полного отказа от запуска.

Это потенциально влияет на клиентские настройки:

```text id="b8qccp"
timeouts
token budgets
subagent selection
tool configuration
context assumptions
cache settings
UI presentation
```

### Важно: alias не делает Opus 5 быстрее

Фактическая схема остается следующей:

```text id="dhsv7g"
Client identity     = Claude Haiku 4.5
Actual inference    = Claude Opus 5
AgentRouter billing = Claude Opus 5
Actual model speed  = Claude Opus 5
Actual intelligence = Claude Opus 5
```

Alias — это **механизм клиентской совместимости**, а не оптимизация inference.

Он не:

```text id="hbz87c"
превращает Opus в Haiku
снижает стоимость upstream-модели
изменяет AgentRouter billing
открывает недоступные модели
увеличивает quota
обходит model entitlement
```

Если конкретный клиент имеет более быстрый или оптимизированный внутренний путь для Haiku, alias позволяет использовать именно этот **client-side path**.

Сам inference по-прежнему выполняет Opus 5.

---

## Совместимость с более широкой экосистемой coding agents

AgentRouter уже работает с несколькими известными coding tools.

Практический интерес этого proxy — в более широкой экосистеме provider-neutral и multi-provider agents.

Например:

| Agent | Почему интересен |
|---|---|
| **Qwen Code** | Зрелый coding CLI, Anthropic/OpenAI и широкий выбор providers |
| **Claude Code-compatible backends** | Сильный agent UX и возможность использовать альтернативный gateway |
| **Autohand** | Terminal-native и provider-neutral |
| **Reasonix** | Архитектура с акцентом на стоимость и caching |
| **Jazz** | Легковесный terminal agent с гибким выбором модели |
| **ForgeCode** | Много providers и разные модели для разных задач |
| **Neovate** | Open-source CLI, headless mode, skills и MCP |
| **Pi** | Очень высокая model neutrality и большой выбор providers |
| **Hermes** | Большое количество providers и reusable agent skills |
| **Goose** | Зрелый open-source agent с большим количеством providers и MCP |
| **Crush** | Сильный multi-model terminal harness |
| **Command Code** | Интересный agent UX, но не рассчитан на произвольное использование AgentRouter API credentials |

Этим проектам не требуется реализовывать AgentRouter-specific transport самостоятельно.

Если клиент уже способен использовать Anthropic-compatible или OpenAI-compatible API, специфичная для AgentRouter совместимость может быть вынесена в локальный proxy.

---

## Поддерживаемые API

Proxy предоставляет оба основных API-варианта.

### Anthropic Messages

```text id="nacr2c"
POST http://127.0.0.1:8318/v1/messages
```

Client model:

```text id="9h4lgj"
claude-haiku-4-5
```

Actual upstream model:

```text id="v16npx"
claude-opus-5
```

### OpenAI Chat Completions

```text id="8x5zrp"
POST http://127.0.0.1:8318/v1/chat/completions
```

Тот же alias может использоваться через OpenAI-compatible интерфейс.

---

## Model discovery

Клиент, выполняющий:

```text id="zaqnj8"
GET /v1/models
```

видит compatibility model:

```json id="j2ssc7"
{
  "data": [
    {
      "id": "claude-haiku-4-5",
      "object": "model",
      "owned_by": "agentrouter"
    }
  ],
  "object": "list"
}
```

Знать реальную upstream-модель клиенту не требуется.

---

## Request / response rewriting

### Клиент отправляет

```json id="vxyhsm"
{
  "model": "claude-haiku-4-5",
  "messages": [
    {
      "role": "user",
      "content": "Fix this Kubernetes manifest"
    }
  ]
}
```

### Proxy отправляет AgentRouter

```json id="jl34sn"
{
  "model": "claude-opus-5",
  "messages": [
    {
      "role": "user",
      "content": "Fix this Kubernetes manifest"
    }
  ]
}
```

### AgentRouter может вернуть

```json id="od6m33"
{
  "model": "anthropic/claude-opus-5-ps-aws-dst"
}
```

### Клиент получает

```json id="sd24dw"
{
  "model": "claude-haiku-4-5"
}
```

Переписываются только необходимые поля, связанные с идентификатором модели.

Ответ модели, tool calls, usage и остальные данные сохраняются.

---

## Anthropic SSE

Model alias работает и для streaming-ответов Anthropic.

Upstream:

```text id="a1kkf0"
event: message_start
data: {"type":"message_start","message":{"model":"anthropic/claude-opus-5-ps-aws-dst"}}
```

Client:

```text id="1xl12g"
event: message_start
data: {"type":"message_start","message":{"model":"claude-haiku-4-5"}}
```

SSE framing сохраняется, меняется только соответствующее поле `model`.

---

## Prompt caching сохраняется

Model aliasing не должен ломать Anthropic prompt caching.

Proxy сохраняет:

```text id="6m7opp"
system prompts
messages
tools
cache_control
thinking blocks
tool calls
usage metadata
```

Типичный coding-agent request содержит большой стабильный prefix:

```text id="kf88pw"
system instructions
repository context
QWEN.md / CLAUDE.md
tool definitions
project conventions
architecture
conversation history
────────────────────────────
small new user request
```

Если используемый AgentRouter channel поддерживает prompt caching, usage может выглядеть так:

```json id="r0ty4o"
{
  "cache_creation_input_tokens": 12000,
  "cache_read_input_tokens": 11800
}
```

Сам alias не повышает cache-hit rate.

Эффективность caching зависит от:

- реального upstream provider;
- структуры prompt;
- стабильности prefix;
- `cache_control`;
- времени жизни cache.

Задача proxy — сохранить эти механизмы при выполнении compatibility rewriting.

---

## Изоляция credentials

Coding agent не требуется реальный AgentRouter API key.

Между клиентом и локальным proxy можно использовать отдельный локальный token:

```text id="ggf9pv"
Coding Agent
    │
    │ local disposable token
    ▼
Local Proxy
    │
    │ real AgentRouter credential
    ▼
AgentRouter
```

Настоящий upstream credential остается внутри окружения proxy.

Это особенно удобно при использовании нескольких coding agents и IDE extensions.

---

## Архитектура

```text id="xorf3k"
┌──────────────────────────────────────────┐
│ Coding Agent / IDE                       │
│                                          │
│ model: claude-haiku-4-5                  │
│ Anthropic or OpenAI-compatible protocol  │
└───────────────────┬──────────────────────┘
                    │
                    ▼
┌──────────────────────────────────────────┐
│ AgentRouter Model Alias Proxy            │
│ 127.0.0.1:8318                           │
│                                          │
│ • local authentication                   │
│ • model advertisement                    │
│ • request model rewriting                │
│ • response model rewriting               │
│ • Anthropic SSE rewriting                │
│ • WAF-compatible transport               │
│ • upstream credential isolation          │
└───────────────────┬──────────────────────┘
                    │
                    │ claude-opus-5
                    ▼
┌──────────────────────────────────────────┐
│ AgentRouter                              │
│                                          │
│ actual model: Claude Opus 5              │
│ actual accounting: Claude Opus 5         │
└──────────────────────────────────────────┘
```

---

## Что дает проект

```text id="8hptt4"
✅ AgentRouter WAF compatibility
✅ Claude model-name compatibility
✅ Anthropic Messages API
✅ OpenAI Chat Completions API
✅ SSE-aware model rewriting
✅ static model advertisement
✅ local credential isolation
✅ stable client-facing model identity
✅ поддержка большего числа coding-agent frontends
```

---

## Что проект НЕ делает

```text id="z4ojm6"
❌ более дешевый Opus через Haiku alias
❌ hidden model access
❌ quota bypass
❌ billing bypass
❌ entitlement bypass
❌ unauthorized upstream models
```

AgentRouter всегда получает реальный настроенный upstream Model ID.

Если proxy настроен как:

```text id="q4u8nm"
claude-haiku-4-5 → claude-opus-5
```

AgentRouter выполняет и учитывает:

```text id="qzom1y"
claude-opus-5
```

---

## Почему это полезно

Современные coding agents все чаще отделяют agent UX от конкретного inference provider.

Такой подход хорошо сочетается с локальным compatibility gateway:

```text id="nf7n98"
Coding Agent
      │
      │ standard Anthropic/OpenAI API
      ▼
Compatibility Proxy
      │
      │ AgentRouter-specific transport
      ▼
AgentRouter
```

Agent остается provider-neutral, а вся специфичная для AgentRouter логика — transport, WAF, credentials и model aliasing — остается в одном локальном слое.

---

## Источники

- AgentRouter documentation: https://docs.agentrouter.org/
- AgentRouter third-party integration guide: https://co.agentrouter.org/portal/guide
- Anthropic — Claude Haiku 4.5: https://www.anthropic.com/news/claude-haiku-4-5
