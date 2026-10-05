# webhook relay

## Как это работает
![how it work](howItWork.png)

## Основные сущности

### Subscriptions
| name | type| requires| default |
|--|--|-|-|
|id        |uuid|t|-|
|url       |text|t|-|
|secret    |text|t|-|
|active    |bool|t|t|
|created_at| timestamp | t | current time |

### SubscriptionEvents
--

### Events

| name | type| requires| default |
|--|--|-|-|
|id          |uuid      |t|-|
|event_id    |text      |t|-|
|event_type  |varchar   |t|-|
|payload     |jsonb     |t|-|
|create_at   |timestamp |t|current time|

### Delivery
| name | type| requires| default |
|--|--|-|-|
|id              |uuid      |t|-|
|event_id        |uuid      |t|-|
|subscription_id |uuid      |t|-|
|status          |varchar   |t|-|
|attempts        |int       |t|-|
|last_attempt_at |timestamp |t|-|

