package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sync"
)

type Wallet struct {
    WalletID int     `json:"wallet_id"`
    Currency string  `json:"currency"`
    Sum      float64 `json:"sum"`
}

var rightDone int = 0

func main() {
    walletIDs := []int{1251513451454616, 5435452135151251, 1321513451454616} // Пример массива номеров кошельков

    var wg sync.WaitGroup
    wg.Add(10000)


    for i := 0; i < 10000; i++ {
		wid := rand.Intn(3)
        go func(id int) {
            defer wg.Done()

            // Создание и заполнение запроса
            url := "http://localhost:8082/api/invoice" // Измените URL на соответствующий REST API
            payload := Wallet{
                WalletID: id,
                Currency: "EUR",
                Sum:      3.0,
            }
            requestBody, err := json.Marshal(payload)
            if err != nil {
                fmt.Printf("Ошибка при маршалинге JSON: %s\n", err)
                return
            }

            // Отправка POST запроса
            resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
            if err != nil {
                fmt.Printf("Ошибка при отправке POST запроса: %s\n", err)
                return
            }
            defer resp.Body.Close()

            // Проверка ответа сервера
            if resp.StatusCode != http.StatusOK {
                fmt.Println("Ошибка: сервер не дал успешный ответ")
                return
            }

            // Чтение тела ответа
            body, err := ioutil.ReadAll(resp.Body)
            if err != nil {
                fmt.Printf("Ошибка при чтении ответа сервера: %s\n", err)
                return
            }

            // Вывод тела ответа
			rightDone++
            fmt.Printf("Ответ сервера: %s\n", string(body))
        }(walletIDs[wid])
    }
    wg.Wait()
	fmt.Println(rightDone)
	
}
// package main

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"math/rand"
// 	"net/http"
// )

// type Wallet struct {
// 	WalletID int     `json:"wallet_id"`
// 	Currency string  `json:"currency"`
// 	Sum      float64 `json:"sum"`
// }

// type Request struct{
// 	Id int
// 	ItsWallet Wallet
// }

// func main() {
// 	fmt.Println("starting..")
// 	// Входные данные - массив номеров кошельков
// 	wallets := []int{1251513451454616, 1321513451454616, 5435452135151251}

// 	// Проходим по каждому номеру кошелька
// 	for i := 0; i < 1000; i++ {
// 		//fmt.Println(walletID)
// 		wid := rand.Intn(3)

// 		go func(){

// 		// Создаем объект запроса для добавления денег
// 		invoiceRequest := Wallet{
// 			WalletID: wallets[wid],
// 			Currency: "EUR",
// 			Sum:      3.0,
// 		}


// 		// Отправляем запрос на добавление денег
// 		err := sendPostRequest("http://localhost:8082/api/invoice", invoiceRequest)
// 		if err != nil {
// 			fmt.Println("Ошибка при добавлении денег на счет:", err)
// 		}
// 		}()

// 		// go func(){
// 		// // Создаем объект запроса для снятия денег
// 		// withdrawRequest := Wallet{
// 		// 	WalletID: wallets[wid],
// 		// 	Currency: "EUR",
// 		// 	Sum:      1.0,
// 		// }


// 		// // Отправляем запрос на снятие денег
// 		// err := sendPostRequest("http://localhost:8082/api/withdraw", withdrawRequest)
// 		// if err != nil {
// 		// 	fmt.Println("Ошибка при снятии денег со счета:", err)
// 		// }
// 		// }()
// 	}
// }



// func sendPostRequest(url string, data Wallet) error {
// 	a := rand.Intn(100)

// 	fmt.Println(a, data)
// 	// Преобразуем данные запроса в JSON
// 	jsonData, err := json.Marshal(data)
// 	if err != nil {
// 		return err
// 	}


// 	// Создаем POST запрос
// 	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		return err
// 	}

// 	// Устанавливаем заголовок Content-Type
// 	req.Header.Set("Content-Type", "application/json")

// 	// Отправляем запрос
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	fmt.Println("EBAAAT")
// 	// Проверяем код состояния ответа
// 	if resp.StatusCode != http.StatusOK {
// 		return fmt.Errorf("ошибка запроса. Код состояния: %d", resp.StatusCode)
// 	}

// 	fmt.Println("Запрос успешно выполнен:", url)

// 	return nil
// }
