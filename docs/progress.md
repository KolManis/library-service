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
