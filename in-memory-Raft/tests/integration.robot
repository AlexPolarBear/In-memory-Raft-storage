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
${KEY}                testKey
${VALUE}              testValue


*** Test Cases ***
Корректный Put Запрос
    [Documentation]    Тест на корректный PUT запрос для сохранения пары ключ-значение.
    Create Session    raft    ${SERVER}
    &{data}=    Create Dictionary    key=${KEY}    value=testValue
    ${resp}=    POST On Session    raft    /keys    json=${data}
    Should Be Equal As Strings    ${resp.status_code}    200

Корректный Get Запрос
    [Documentation]    Тест на корректный GET запрос для получения пары ключ-значение.
    Create Session    raft    ${SERVER}
    ${resp}=    GET On Session    raft    /keys/${KEY}
    Should Be Equal As Strings    ${resp.status_code}    200
    ${result}=    Set Variable    ${resp.content}  
    Should Contain    ${result}    ${VALUE}

Корректный Delete Запрос
    [Documentation]    Тест на корректный DELETE запрос для удаления пары ключ-значение.
    Create Session    raft    ${SERVER}
    ${resp}=    DELETE On Session    raft    /keys/${KEY}
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
    Set Suite Variable    ${PROCESS_ID}

Остановить Сервер
    ${process_alive}=    Is Process Running    ${PROCESS_ID}
    Run Keyword If    ${process_alive}    Terminate Process    ${PROCESS_ID}    kill=True

Инициализировать Данные
    Create Session    raft    ${SERVER}
    &{data}=    Create Dictionary    key=${KEY}    value=${VALUE}
    POST On Session    raft    /keys    json=${data}
