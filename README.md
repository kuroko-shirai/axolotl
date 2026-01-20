# axolotlt

__Axolotl__ — это библиотека на Go, которая обеспечивает интеллектуальную маршрутизацию команд чтения между мастер- и реплика-нодами Redis в автономной (standalone) конфигурации. На основе мониторинга текущей загрузки CPU библиотека динамически выбирает: выполнять чтение с мастеров или с реплик — оптимизируя производительность и защищая ёмкость записи.

## Возможности

- Динамическая маршрутизация чтения: автоматически направляет команды `GET`, `HGET`, `SMEMBERS` и другие на основе реальной загрузки CPU.
- Поддержка различных режимов выполнения: одиночные команды, пакетные (`DoMulti`), кэшированные (`DoCache`, `DoMultiCache`).
- Ориентация на standalone-режим: разработана специально для независимых инстансов __Redis__ с явным разделением ролей «мастер/реплика».
- Безопасность по умолчанию: проверяет, что команда предназначена только для чтения; отклоняет попытки записи через балансировщик.
- Готовность к наблюдаемости: легко интегрируется с внешними системами мониторинга через интерфейс `Monitor`.

## Пример работы

1. Конфигурация `config.yaml`

Создайте файл с конфигурацией
```yaml
redis:
  username: "ваш_пользователь"
  password: "ваш_пароль_redis"
  masters:
    addresses: ["10.0.0.1:6379", "10.0.0.2:6379"]
    maxThreshold: 15.0
  replicas:
    addresses: ["10.0.0.3:6379", "10.0.0.4:6379"]
    maxThreshold: 10.0
```

2. Инициализация клиента

```go 
// Загрузка конфигурации
cfg := loadConfig("config.yaml")

// Запуск монитора
monitor, _ := monitor.New(monitor.Config{
    Username:  cfg.Username,
    Password:  cfg.Password,
    Addresses: append(cfg.Masters.Addresses, cfg.Replicas.Addresses...),
    Delay:     time.Second,
})
go monitor.Run(ctx)

// Создание адаптивного маршрутизатора
cobweb, _ := cobweb.New(&cobweb.Config{
    Masters:  &cluster.Config{...},
    Replicas: &cluster.Config{...},
    Monitor:  &monitor,
})

// Формирование команды Redis
cmd := redisClient.Get("user:123")

// Выполнение с адаптивной маршрутизацией
results, _ := cobweb.Execute(ctx, cobweb.SingleCmd{Cmd: cmd})
value := results[0].ToString()
```
