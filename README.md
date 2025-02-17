# go-musthave-shortener-tpl

<p align="center">
  <a href="https://go.dev/">
    <img src="https://img.shields.io/badge/golang-v1.22.7-lightblue" height="25" alt="Go version"/>
  </a>
  <a href="https://codecov.io/github/GlebRadaev/shlink">
    <img src="https://codecov.io/github/GlebRadaev/shlink/branch/main/graph/badge.svg?token=QSF4QTYP52" height="25" alt="Code Coverage">
  </a>
  <a href="https://github.com/GlebRadaev/shlink/actions/workflows/shortenertest.yml">
    <img src="https://github.com/GlebRadaev/shlink/actions/workflows/shortenertest.yml/badge.svg?branch=main" height="25" alt="Autotests">
  </a>
</p>

Шаблон репозитория для трека «Сервис сокращения URL».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m main template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/main .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).
