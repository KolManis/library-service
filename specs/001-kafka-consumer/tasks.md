# Задачи: KAFKA-003 — консьюмер `book.decommissioned`

Порядок — снизу вверх по зависимостям, как в `COPY-001`: контракт → порты →
адаптеры → application → транспорт → точка входа → тесты. Каждый блок
самодостаточен и заканчивается зелёной сборкой.

Спека: [spec.md](spec.md)

## 1. Контракт события

- [ ] `api/events/v1/book_decommissioned.proto` — `BookDecommissioned`
      (`book_id`, `occurred_at`), пакет `events.v1`, `go_package` как в
      соседних файлах
- [ ] `buf lint && buf generate` — появился `pkg/events/v1/book_decommissioned.pb.go`
- [ ] `go build ./...`

## 2. Порты

- [ ] `ICopyRepository.DecommissionAllByBookID(ctx, bookID) error`
- [ ] `ILoanRepository.FindReservedByBookID(ctx, bookID) ([]*loan.Loan, error)`
      — возвращает пустой срез, если броней нет (не ошибка, по образцу
      `FindReservedByCopyID`)
- [ ] Доучить `fakeCopyRepo`/`fakeLoanRepo` в `reservecopy`, `issueloan`,
      `returnloan`, `decommissioncopy` — интерфейсы выросли, заглушки должны
      компилироваться (в прошлый раз про это забыли и поймали три ошибки
      компиляции)
- [ ] `go vet ./...`

## 3. Адаптеры out

- [ ] `CopyRepository.DecommissionAllByBookID` — `UPDATE copies SET
      decommissioned = true WHERE book_id = ?`. Ноль затронутых строк —
      **не** ошибка (книги может не быть, см. спеку)
- [ ] `LoanRepository.FindReservedByBookID` — джойн `loans` ↔ `copies` по
      `copy_id`, фильтр `copies.book_id = ? AND loans.status = 'reserved'`
- [ ] Оба метода через `dbFromContext(ctx, r.db)` — иначе выпадут из
      транзакции
- [ ] Godoc на обоих (экспортируемые методы, `docs/style.md`)

## 4. Application

- [ ] `internal/core/application/commands/decommissionbook/command.go` —
      `Command{BookID}`, `NewCommand` с проверкой на `uuid.Nil`
- [ ] `command_handler.go` — в `WithinTransaction`:
      `DecommissionAllByBookID` → `FindReservedByBookID` → для каждой брони
      `Cancel(now)` + `Update` + накопить события → один `outbox.Append`
- [ ] `command_handler_test.go` — сценарии: несколько броней отменены;
      броней нет (вхолостую); неизвестная книга (не ошибка); сбой
      репозитория пробрасывается; `issued`-бронь не попадает в выборку
- [ ] `go test ./internal/core/application/commands/decommissionbook/...`

## 5. Адаптер in — консьюмер

- [ ] `internal/adapters/in/consumers/book_decommissioned_consumer.go` —
      `sarama.ConsumerGroupHandler`: `Setup`, `Cleanup`, `ConsumeClaim`
- [ ] `ConsumeClaim`: декодировать protobuf → `decommissionbook.NewCommand`
      → `Handle` → `session.MarkMessage` **только после успеха**
- [ ] Битое сообщение: залогировать (`slog.Error` с offset) и закоммитить
      offset, не блокируя партицию (см. спеку)
- [ ] Ошибка обработки (БД недоступна): **не** коммитить offset — пусть
      придёт повторно

## 6. Точка входа

- [ ] `KAFKA_CONSUMER_GROUP` в `internal/config/config.go`, дефолт
      `library-service`
- [ ] `cmd/book-consumer/main.go` — `infra.NewLogger`/`NewPostgres`/
      `ShutdownContext`, `sarama.NewConsumerGroup`, цикл
      `for { cg.Consume(ctx, ...) }` с выходом по `ctx.Done()`
- [ ] `cg.Close()` с проверкой ошибки (`errcheck` иначе не пропустит)
- [ ] `go build ./... && golangci-lint run ./...`

## 7. Интеграционный тест

- [ ] `test/integration/book_consumer_test.go` — `BookConsumerSuite`,
      Kafka через `tckafka`, **уникальные топик и consumer group на тест**
      (`+ uuid.NewString()`)
- [ ] Основной сценарий: фикстуры (книга, 2 копии, бронь на одной) →
      опубликовать событие → дождаться обработки → проверить
      `decommissioned = true` у обеих копий, `cancelled` у брони, событие в
      `outbox`
- [ ] `issued`-бронь не тронута
- [ ] Повторная доставка того же сообщения — состояние не меняется, ошибок нет
- [ ] Событие про неизвестную книгу — обработано вхолостую
- [ ] Остановка консьюмера в `TearDown` — `cancel()` + дождаться выхода
      горутины, иначе она переживёт тест
- [ ] `go test -tags=integration ./test/integration/... -run TestBookConsumerSuite`

## 8. Закрытие фичи

- [ ] Полный прогон: `go build ./... && go vet ./... && go test ./... &&
      golangci-lint run ./... && go test -tags=integration ./test/integration/...`
- [ ] `docs/feature_list.json`: `KAFKA-003` → `done`, ссылка на эту спеку
- [ ] `docs/progress.md`: запись сессии
- [ ] Сверить результат со спекой: все 8 критериев приёмки закрыты
