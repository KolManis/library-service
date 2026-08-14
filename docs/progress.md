# Progress Log

## 2026-07-31

Task: RETURN-001 (довести до конца), QUERY-001 (начать), внедрение харнесса
(этот файл, `CLAUDE.md`, `feature_list.json`)

Changed:
- `issueloan` — ревью и правки стиля (issue_handler.go, command_handler.go).
- `returnloan` — реализована команда целиком: домен `Fine` (internal/core/domain/aggregates/fine),
  порты `IFineRepository`/`ITransactor`, адаптеры (`fine_repository.go`, `transactor.go`,
  правка `loan_repository.go` под `dbFromContext`), HTTP-хендлер, unit-тесты,
  wiring в `cmd/api/main.go`. Исправлены: асимметрия нейминга `toModel`/`toDomain`,
  мёртвый код (`toFineDomain`), лишний defensive nil-check на `dueAt` (не по стилю
  проекта — домен и так гарантирует инвариант, см. `MarkOverdue`).
- `getreaderloans` — начато: `Query`/`NewQuery` и порт `IReaderLoansReader` + `LoanView`
  готовы. `query_handler.go` пуст, репозиторий/HTTP/wiring/тесты не начаты.
- `DESIGN.md` — решение 2.6 (расчёт суммы штрафа), раздел 11 расширен
  (Kibana/Grafana/Jaeger/Kafka UI, cosign/подписание образов), заметка про mockery
  в разделе «Тесты».

Verified:
- `go build ./...`, `go vet ./...`, `gofmt -l` — чисто.
- `go test ./...` — все unit-тесты зелёные (reservecopy, issueloan, returnloan, fine, loan).
- Integration-тесты (testcontainers) в этой сессии не запускались.

Remaining:
- RETURN-001: код и тесты готовы, лежит в `feature/return-loan` на origin, PR в
  `main` не открыт.
- QUERY-001: `query_handler.go`, `FindByReaderID` в `LoanRepository`, HTTP-хендлер,
  wiring в `main.go`, тесты — см. `feature_list.json` → QUERY-001.

Next session: продолжить QUERY-001 с `query_handler.go` (тонкий проксирующий
хендлер к `IReaderLoansReader`), затем `FindByReaderID` в `loan_repository.go`.
После завершения QUERY-001 этап 2 из DESIGN.md закрыт полностью — дальше RACE-001
(этап 3).

## 2026-08-02

Task: QUERY-001 (завершить), единый code style репозитория

Changed:
- `getreaderloans` — дописаны `query_handler.go`, `FindByReaderID` в
  `LoanRepository`, `reader_loans_handler.go`, wiring в `main.go`, 3 unit-теста
  (успех/пусто/ошибка репозитория). QUERY-001 закрыт полностью — этап 2 из
  DESIGN.md закрыт.
- `docs/style.md` — новый файл: правила по комментариям (по умолчанию не пишем,
  5 явных исключений), импортам (2 блока, без ручного `-local`-разделения),
  переносу параметров (3+ — по одному на строку), неймингу (`fake<Порт>`, поле
  хендлера не `handler`). Правило зафиксировано после ревью реальных расхождений
  между `reservecopy`/`issueloan`/`reserve_handler.go` (старые, с комментариями
  через строчку) и `returnloan`/`return_handler.go` (новые, лаконичные).
- Применено по всему репозиторию: убраны комментарии-дублёры в HTTP-хендлерах и
  command-хендлерах (`reserve_handler.go`, `issue_handler.go`, `reservecopy`,
  `issueloan`), выровнена нумерация шагов (`Шаг N:` → `N.`), унифицированы
  импорты в `issue_handler.go`/`return_handler.go`, добавлены комментарии на
  `fineModel`/`TableName()` и `fine.NewFine` по образцу `loan`-аггрегата,
  `reader_loans_handler.go` упрощён (не форматируем даты в строки вручную —
  `encoding/json` сам сериализует `time.Time`), тестовый fake переименован
  `fakeReader` → `fakeReaderLoansReader`.

Verified:
- `go build ./...`, `go vet ./...`, `go test ./...` — чисто, все зелёные.
- `gofmt -l` по всем изменённым файлам (с поправкой на CRLF от Windows-редактора)
  — чисто.

Правило комментариев из `docs/style.md` пересмотрено в этой же сессии: вместо
«без комментариев по умолчанию» — **godoc на каждом экспортируемом**
идентификаторе (функция/метод/тип/константа, комментарий начинается с имени —
стандарт `go doc`/IDE-подсказок/`golint`). Приватное — как было, без комментария,
если не несёт что-то не считываемое из кода. Применено по всему репозиторию:
`loan.go`, `status.go`, `fine.go`, все ports, все command/query-хендлеры, все
HTTP-хендлеры, все репозитории — везде добавлены недостающие godoc-комментарии.
`go doc` проверен вручную (`go doc ./internal/.../returnloan Handler` и т.п.) —
рендерится корректно.

Remaining:
- QUERY-001 не запушен в origin (только локально на `feature/get-reader-loans`).
- RETURN-001 всё ещё не смёржен в main (см. запись от 2026-07-31).

## 2026-08-03

Task: RACE-001 (этап 3), branch protection на `main`, план SDET-практики (отдельный PR)

Changed:
- `RETURN-001`/`QUERY-001`/план SDET — смёржены в `main` (PR #7, #8, #9). PR #8
  смёржен squash'ем (наши 3 отдельных коммита схлопнулись в один) — единственный
  раз так, до этого и после — обычный merge commit.
- Настроен GitHub ruleset на `main`: require PR before merging, required status
  checks (`lint`/`unit`/`integration`), block force pushes, restrict deletions.
  Linear history сознательно не включали — конфликтует с обычным merge commit,
  которым мержим PR.
- `RACE-001`: `test/integration/reserve_race_test.go` (100 горутин на 1 свободный
  экземпляр книги), `migrations/00002_uniq_active_loan_per_copy.sql` (частичный
  уникальный индекс), `ports.ErrConcurrentReservation` + обработка SQLSTATE
  `23505` в `LoanRepository.Create` (`errors.As` на `*pgconn.PgError`), маппинг
  на 409 в `reserve_handler.go`. Убрано дублирование тестовых фикстур-переменных
  (`raceBookID` и т.п. дублировали уже существующие `fixtureBookID` и т.п. в том
  же пакете) и мёртвая переменная `raceCopyID1`.

Verified:
- Цикл красный → зелёный → откат миграции → красный → восстановление прогнан
  вживую через реальный Docker/testcontainers (не только по коду): без индекса
  тест реально ловит овербукинг (`actual: 2` вместо `1`), с индексом — зелёный.
- Весь integration-сьют (`CopyRepositorySuite`, `LoanRepositorySuite`,
  `MigrationsSuite`, `RaceSuite`) зелёный после фикса.
- `go build`/`go vet`/unit-тесты/`gofmt` — чисто.

Remaining:
- Этап 3 из DESIGN.md закрыт полностью.
- Дальше по roadmap: `KAFKA-001` (этап 4) и `WORKER-001` (этап 5) — оба
  `priority: low`, не начаты.

Next session: начать `KAFKA-001` — proto-контракты событий, outbox-таблица,
продюсер `loan.changed`/`fine.created`, консьюмер `book.decommissioned`.
- Style-правки затрагивают файлы вне scope QUERY-001 (`reserve_handler.go`,
  `reservecopy/command_handler.go`) — стоит коммитить отдельным коммитом/PR, не
  мешать со сменой фичи.

Next session: запушить `feature/get-reader-loans`, открыть PR. Начать RACE-001
(этап 3) — тест на овербукинг до добавления частичного уникального индекса.

## 2026-08-14

Task: KAFKA-001 (доменные события + outbox, этап 4 часть 1/3)

Changed:
- `internal/core/domain/events` — интерфейс `DomainEvent`.
- `loan.LoanChanged`/`fine.FineCreated` — события, поле `events` + `PullEvents()`
  на обоих агрегатах, подняты во всех переходах (`Reserve`/`Issue`/`Return`/
  `Expire`/`MarkOverdue`/`NewFine`).
- `migrations/00003_outbox.sql`, `ports.IOutboxRepository`,
  `adapters/out/repositories/outbox_repository.go`.
- `reservecopy`/`issueloan`/`returnloan` — все три хендлера обёрнуты в
  `ITransactor`, вызывают `outbox.Append` внутри той же транзакции, что и
  сохранение агрегата. `main.go` и все fake-тесты (`fakeOutboxRepo`,
  `fakeTransactor`) обновлены под новые сигнатуры `NewHandler`.
- `CopyRepository.FindFreeCopyID` переведён на `dbFromContext` (консистентность
  с остальными репозиториями).
- `test/integration/outbox_atomicity_test.go` — форсированный сбой
  `outbox.Append` через `failingOutboxRepo`, проверка что `Create` тоже
  откатился (прямой SQL по `loans`/`outbox`).
- `DESIGN.md` §11 — новый подраздел «Приоритет — сделать раньше, а не
  откладывать»: `.golangci.yml` (`errorlint`/`contextcheck`/`sqlclosecheck`/
  `forbidigo`), `slog`, graceful shutdown, единый error-mapping слой. Заведена
  `INFRA-EARLY-001` (`priority: medium`) под это.

Verified:
- `go build`/`go vet`/unit-тесты — чисто.
- Полный integration-сьют на реальном Postgres: `CopyRepositorySuite`,
  `LoanRepositorySuite`, `MigrationsSuite`, `RaceSuite`,
  `OutboxAtomicitySuite` — все зелёные.
- Атомарность outbox подтверждена не только код-ревью, а реальным прогоном:
  форсированная ошибка `Append` → ни `Loan`, ни `outbox`-строка не появились.

Remaining:
- `KAFKA-001` закрыт полностью (этап 4 часть 1/3).
- `INFRA-EARLY-001` — не начат, приоритет выше остального бэклога.
- `KAFKA-002` (продюсер, `buf`/protobuf) и `KAFKA-003` (консьюмер,
  `Loan.Cancelled`) — не начаты, зависят друг от друга по порядку.

Next session: по обсуждению — сделать `INFRA-EARLY-001` перед `KAFKA-002`
(дешевле внедрить линтеры/`slog`/graceful shutdown сейчас, чем после того как
`KAFKA-002`/`003` добавят продюсера и консьюмера).
