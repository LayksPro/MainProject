package main

import (
	"fmt"
	"log"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/levigross/grequests"
)

func main() {
	// Открытие файла Excel
	filePath := "путь_к_файлу.xlsx"
	xlsx, err := excelize.OpenFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	// Чтение данных из файла Excel
	rows := xlsx.GetRows("Лист1")

	// Подключение к API Zabbix
	url := "http://your_zabbix_url/api_jsonrpc.php"
	username := "your_username"
	password := "your_password"
	authToken, err := zabbixLogin(url, username, password)
	if err != nil {
		log.Fatal(err)
	}

	// Создание нового узла в Zabbix
	for _, row := range rows {
		hostname := row[0]  // Предполагается, что имя хоста находится в первой колонке
		ipAddress := row[1] // Предполагается, что IP-адрес находится во второй колонке

		_, err := createZabbixHost(url, authToken, hostname, ipAddress)
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("Узел %s успешно создан в Zabbix\n", hostname)
		}
	}

	// Выход из API Zabbix
	err = zabbixLogout(url, authToken)
	if err != nil {
		log.Println(err)
	}
}

func zabbixLogin(url, username, password string) (string, error) {
	// Подготовка запроса на авторизацию
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "user.login",
		"params": map[string]string{
			"user":     username,
			"password": password,
		},
		"id":   1,
		"auth": nil,
	}

	// Отправка запроса на авторизацию
	resp, err := grequests.Post(url, &grequests.RequestOptions{
		JSON: payload,
	})
	if err != nil {
		return "", err
	}

	// Обработка ответа
	var loginResp struct {
		Result string `json:"result"`
		Error  struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	err = resp.JSON(&loginResp)
	if err != nil {
		return "", err
	}

	// Проверка наличия ошибок
	if loginResp.Error.Code != 0 {
		return "", fmt.Errorf("ошибка авторизации: %s", loginResp.Error.Message)
	}

	return loginResp.Result, nil
}

func createZabbixHost(url, authToken, hostname, ipAddress string) (string, error) {
	// Подготовка запроса на создание хоста
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "host.create",
		"params": map[string]interface{}{
			"host": hostname,
			"interfaces": []map[string]interface{}{
				{
					"type":  1,
					"main":  1,
					"useip": 1,
					"ip":    ipAddress,
					"dns":   "",
					"port":  "10050",
				},
			},
			"groups": []map[string]string{
				{
					"groupid": "1", // Идентификатор группы хостов в Zabbix (по умолчанию: 1)
				},
			},
		},
		"auth": authToken,
		"id":   1,
	}

	// Отправка запроса на создание хоста
	resp, err := grequests.Post(url, &grequests.RequestOptions{
		JSON: payload,
	})
	if err != nil {
		return "", err
	}

	// Обработка ответа
	var createResp struct {
		Result map[string]string `json:"result"`
		Error  struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	err = resp.JSON(&createResp)
	if err != nil {
		return "", err
	}

	// Проверка наличия ошибок
	if createResp.Error.Code != 0 {
		return "", fmt.Errorf("ошибка создания хоста: %s", createResp.Error.Message)
	}

	return createResp.Result["hostids"], nil
}

func zabbixLogout(url, authToken string) error {
	// Подготовка запроса на выход из API
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "user.logout",
		"params":  []interface{}{},
		"auth":    authToken,
		"id":      1,
	}

	// Отправка запроса на выход из API
	_, err := grequests.Post(url, &grequests.RequestOptions{
		JSON: payload,
	})
	if err != nil {
		return err
	}

	return nil
}
