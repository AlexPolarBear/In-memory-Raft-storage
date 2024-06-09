*** Settings ***
Documentation     Интеграционные тесты для In-memory Raft storage.

Library           RequestsLibrary
Library           Process
Library           OperatingSystem

Suite Setup       Запустить Сервер
Suite Teardown    Остановить Сервер

Test Setup        Инициализировать Данные
Test Teardown     DELETE On Session    raft    /keys/testKey


*** Variables ***
${SERVER}             http://localhost:8080
${RAFT_DIR}           ./raft_data
${RAFT_EXECUTABLE}    go run ./cmd/app/in-memory-raft.go
${RAFT_DATA_PATH}     ${RAFT_DIR}/test_node
${PROCESS_ID}         0


*** Test Cases ***
Некорректный Put Запрос
    [Documentation]    Тест на некорректный PUT запрос с невалидными данными.
    Create Session    raft    ${SERVER}
    &{data}=    Create Dictionary    key=    value=
    ${resp}=    POST On Session    raft    /keys    json=${data}
    Should Be Equal As Strings    ${resp.status_code}    400

Get Запрос Несуществующего Ключа
    [Documentation]    Тест на GET запрос несуществующего ключа.
    Create Session    raft    ${SERVER}
    ${resp}=    GET On Session    raft    /keys/nonExistentKey
    Should Be Equal As Strings    ${resp.status_code}    404

Delete Запрос Несуществующего Ключа
    [Documentation]    Тест на DELETE запрос несуществующего ключа.
    Create Session    raft    ${SERVER}
    ${resp}=    DELETE On Session    raft    /keys/nonExistentKey
    Should Be Equal As Strings    ${resp.status_code}    404

Присоединение Узла С Некорректными Данными
    [Documentation]    Тест на присоединение узла с некорректными данными.
    Create Session    raft    ${SERVER}
    &{data}=    Create Dictionary    addr=    id=
    ${resp}=    POST On Session    raft    /join    json=${data}
    Should Be Equal As Strings    ${resp.status_code}    400


*** Keywords ***
Запустить Сервер
    Create Directory    ${RAFT_DIR}
    ${process_id}=    Start Process    ${RAFT_EXECUTABLE}    -haddr localhost:8080    -raddr localhost:7000    -id "test_node"    ${RAFT_DATA_PATH}    shell=True
    Set Suite Variable    ${PROCESS_ID}

Остановить Сервер
    ${process_alive}=    Is Process Running    ${PROCESS_ID}
    Run Keyword If    ${process_alive}    Terminate Process    ${PROCESS_ID}    kill=True

Инициализировать Данные
    Create Session    raft    ${SERVER}
    &{data}=    Create Dictionary    key=testKey    value=testValue
    POST On Session    raft    /keys    json=${data}
