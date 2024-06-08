*** Settings ***
Documentation     Интеграционные тесты для In-memory Raft storage.

Library           RequestsLibrary
Library           Process
Library           OperatingSystem

Suite Setup       Запустить Сервер
Suite Teardown    Остановить Сервер

Test Setup        Инициализировать Данные
Test Teardown     Очистить Данные


*** Variables ***
${SERVER}             http://localhost:8080
${RAFT_DIR}           ./raft_data
${RAFT_EXECUTABLE}    go run ./cmd/app/in-memory-raft.go
${RAFT_DATA_PATH}     ${RAFT_DIR}/test_node


*** Test Cases ***
Корректный Put Запрос
    [Documentation]    Тест на корректный PUT запрос для сохранения пары ключ-значение.
    Create Session    raft    ${SERVER}
    &{data}=    Create Dictionary    key=testKey    value=testValue
    ${resp}=    POST On Session    raft    /keys    json=${data}
    Should Be Equal As Strings    ${resp.status_code}    200

Корректный Get Запрос
    [Documentation]    Тест на корректный GET запрос для получения пары ключ-значение.
    Create Session    raft    ${SERVER}
    ${resp}=    GET On Session    raft    /keys/testKey
    Should Be Equal As Strings    ${resp.status_code}    200
    ${result}=    ${resp.content}
    Dictionary Should Contain Key    ${result}    testKey

Корректный Delete Запрос
    [Documentation]    Тест на корректный DELETE запрос для удаления пары ключ-значение.
    Create Session    raft    ${SERVER}
    ${resp}=    DELETE On Session    raft    /keys/testKey
    Should Be Equal As Strings    ${resp.status_code}    200

Присоединение Нового Узла
    [Documentation]    Тест на присоединение нового узла к кластеру.
    Create Session    raft    ${SERVER}
    &{data}=    Create Dictionary    addr=localhost:8081    id=node2
    ${resp}=    POST On Session    raft    /join    json=${data}
    Should Be Equal As Strings    ${resp.status_code}    200


*** Keywords ***
Запустить Сервер
    Create Directory    ${RAFT_DIR}
    ${process_id}=    Start Process    ${RAFT_EXECUTABLE}    -haddr localhost:8080    -raddr localhost:7000    -id "test_node"    ${RAFT_DATA_PATH}    shell=True
    Set Suite Variable    ${process_id}

Остановить Сервер
    ${process_alive}=    Is Process Running    ${process_id}
    Run Keyword If    ${process_alive}    Terminate Process    ${process_id}    kill=True

Инициализировать Данные
    Create Session    raft    ${SERVER}
    &{data}=    Create Dictionary    key=testKey    value=testValue
    POST On Session    raft    /keys    json=${data}

Очистить Данные
    DELETE On Session    raft    /keys/testKey
