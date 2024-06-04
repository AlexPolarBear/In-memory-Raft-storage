# etcd

Это простая key-value реализация _etcd_. Она была необходима для замеров производительности и последующего сравнения.

Реализация клиента представлена в файле ```etcd.go```. Доступ к такому хранилищу осуществляется с помощью API запросов на localhost:8000. Сами запросы: ```/get```, ```/put```, ```/delete```. Для запроса по ключу добавляется строка ```?key=``` где и прописывается ключ.

Сервер запускался в Docker образе, с помощью команды, прописанной в ```server.sh```.

В файле ```etcd_test.go``` написаны тесты и бенч-тесты для проверки работоспособности и производительности хранилища. Тесты запускались следующей командой:

```bash
go test -bench . -benchmem -cpu 1,2,4
```

В таблице приведены усредненные значения по cpu. Были получены следующие результаты:

```bash
goos: linux
goarch: amd64
pkg: etcd
cpu: Intel(R) Pentium(R) Silver N5000 CPU @ 1.10GHz
```

| name of the benchmark (with cpu) | number of times the loop has been executed | average runtime, expressed in nanoseconds per operation | number of bytes required by the operation | number of allocations done by the operation |
| :--- | :---: | :---: | :---: | :---: |
| Get | 807 | 1 573 137 ns/op | 8 155 B/op | 123 allocs/op |
| Put | 605 | 1 957 670 ns/op | 7 267 B/op | 116 allocs/op |
| Delete | 610 | 1 796 799 ns/op | 7 060 | 114 B/op allocs/op |
| HTTP Get | 533 | 2 014 792 ns/op | 3 729 B/op | 49 allocs/op |
| HTTP Put | 363 | 3 239 796 ns/op | 15 606 B/op | 107 allocs/op |
| HTTP Delete | 360 | 3 041 975 ns/op | 14 546 B/op | 98 allocs/op |

<!-- | BenchmarkPut | 644 | 2107605 ns/op | 7249 B/op | 116 allocs/op |
| BenchmarkPut-2 | 537 | 1905340 ns/op | 7287 B/op | 116 allocs/op |
| BenchmarkPut-4 | 634 | 1860065 ns/op | 7265 B/op | 116 allocs/op |
| BenchmarkGet | 862 | 1465934 ns/op | 8100 B/op | 123 allocs/op |
| BenchmarkGet-2 | 811 | 1568790 ns/op | 8149 B/op | 123 allocs/op |
| BenchmarkGet-4 | 747 | 1684686 ns/op | 8217 B/op | 123 allocs/op |
| BenchmarkDelete | 676 | 1726418 ns/op | 7057 B/op | 114 allocs/op |
| BenchmarkDelete-2 | 576 | 1820482 ns/op | 7057 B/op | 113 allocs/op |
| BenchmarkDelete-4 | 566 | 1853499 ns/op | 7067 B/op | 114 allocs/op |
| BenchmarkHTTPPut | 390 | 3177504 ns/op | 15585 B/op | 107 allocs/op |
| BenchmarkHTTPPut-2 | 357 | 3443161 ns/op | 15628 B/op | 107 allocs/op |
| BenchmarkHTTPPut-4 | 352 | 3098723 ns/op | 15616 B/op | 107 allocs/op |
| BenchmarkHTTPGet | 504 | 2009110 ns/op | 3728 B/op | 49 allocs/op |
| BenchmarkHTTPGet-2 | 522 | 2015120 ns/op | 3729 B/op | 49 allocs/op |
| BenchmarkHTTPGet-4 | 574 | 2020145 ns/op | 3730 B/op | 49 allocs/op |
| BenchmarkHTTPDelete | 369 | 29001223 ns/op | 14536 B/op | 98 allocs/op |
| BenchmarkHTTPelete-2 | 332 | 3014236 ns/op | 14538 B/op | 98 allocs/op |
| BenchmarkHTTPDelete-4 | 379 | 3211568 ns/op | 14565 B/op | 98 allocs/op | -->
