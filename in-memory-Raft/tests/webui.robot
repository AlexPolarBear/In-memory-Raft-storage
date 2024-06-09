*** Settings ***
Documentation     Тестирование веб-интерфейса для In-memory Raft storage.

Library           SeleniumLibrary

Test Setup        Открыть браузер
Test Teardown     Close Browser


*** Variables ***
${SERVER}     http://localhost:8080
${BROWSER}    headlesschrome
${KEY}        testKey
${VALUE}      testValue


*** Test Cases ***
Добавление значения
    Input Text    id=putKey    ${KEY}
    Input Text    id=putValue    ${VALUE}
    Click Button    xpath=//button[text()='Put Value']
    ${result}=    Get Text    id=putValueResult
    Should Be Equal As Strings    ${result}    Value stored successfully
    
Получение значения
    Input Text    id=getKey    ${KEY}
    Click Button    xpath=//button[text()='Get Value']
    ${result}=    Get Text    id=getValueResult
    Should Not Be Empty    ${result}

Удаление ключа
    Input Text    id=deleteKey    ${KEY}
    Click Button    xpath=//button[text()='Delete Key']
    ${result}=    Get Text    id=deleteValueResult
    Should Be Equal As Strings    ${result}    Key deleted successfully

Загрузка журнала транзакций
    Click Button    id=loadLogButton
    ${result}=    Get Text    id=transactionLogResult
    Should Be Equal As Strings    ${result}    Transaction log loaded successfully
    
Сохранение журнала транзакций
    Click Button    id=saveLogButton
    ${result}=    Get Text    id=transactionLogResult
    Should Be Equal As Strings    ${result}    Transaction log saved successfully
    
    
*** Keywords ***
Открыть браузер
    Open Browser    ${SERVER}    ${BROWSER}
    Maximize Browser Window
    Set Selenium Speed    0.5 seconds
    ${status}=    Run Keyword And Return Status    Wait Until Element Is Visible    id=putKey    10s
    Run Keyword If    ${status}    Fail    Не удалось загрузить страницу
