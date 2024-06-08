*** Settings ***
Documentation     Тестирование веб-интерфейса для In-memory Raft storage.

Library           SeleniumLibrary


*** Variables ***
${SERVER}     http://localhost:8080
${BROWSER}    headlesschrome


*** Test Cases ***
Добавление значения
    Открыть браузер
    Ввести ключ для добавления
    Ввести значение для добавления
    Нажать кнопку добавления
    Проверить результат добавления
    Закрыть браузер

Получение значения
    Открыть браузер
    Ввести ключ для получения
    Нажать кнопку получения
    Проверить результат получения
    Закрыть браузер

Удаление ключа
    Открыть браузер
    Ввести ключ для удаления
    Нажать кнопку удаления
    Проверить результат удаления
    Закрыть браузер

Загрузка журнала транзакций
    Открыть браузер
    Нажать кнопку загрузки журнала
    Проверить результат загрузки журнала
    Закрыть браузер

Сохранение журнала транзакций
    Открыть браузер
    Нажать кнопку сохранения журнала
    Проверить результат сохранения журнала
    Закрыть браузер


*** Keywords ***
Открыть браузер
    Open Browser    ${SERVER}    ${BROWSER}
    Maximize Browser Window
    Set Selenium Speed    0.5 seconds
    ${status}=    Run Keyword And Return Status    Wait Until Element Is Visible    id=putKey    10s
    Run Keyword Unless    ${status}    Fail    Не удалось загрузить страницу

Ввести ключ для добавления
    [Arguments]    ${key}=ключ
    Input Text    id=putKey    ${key}

Ввести значение для добавления
    [Arguments]    ${value}=значение
    Input Text    id=putValue    ${value}

Нажать кнопку добавления
    Click Button    xpath=//button[text()='Put Value']

Проверить результат добавления
    ${result}=    Get Text    id=putValueResult
    Should Be Equal As Strings    ${result}    Value stored successfully

Ввести ключ для получения
    [Arguments]    ${key}=ключ
    Input Text    id=getKey    ${key}

Нажать кнопку получения
    Click Button    xpath=//button[text()='Get Value']

Проверить результат получения
    ${result}=    Get Text    id=getValueResult
    Should Not Be Empty    ${result}

Ввести ключ для удаления
    [Arguments]    ${key}=ключ
    Input Text    id=deleteKey    ${key}

Нажать кнопку удаления
    Click Button    xpath=//button[text()='Delete Key']

Проверить результат удаления
    ${result}=    Get Text    id=deleteValueResult
    Should Be Equal As Strings    ${result}    Key deleted successfully

Нажать кнопку загрузки журнала
    Click Button    id=loadLogButton

Проверить результат загрузки журнала
    ${result}=    Get Text    id=transactionLogResult
    Should Be Equal As Strings    ${result}    Transaction log loaded successfully

Нажать кнопку сохранения журнала
    Click Button    id=saveLogButton

Проверить результат сохранения журнала
    ${result}=    Get Text    id=transactionLogResult
    Should Be Equal As Strings    ${result}    Transaction log saved successfully

Закрыть браузер
    Close Browser
