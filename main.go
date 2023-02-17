package main

// этот телеграм бот будет принимать ссылки и иногда будет кидать пользователю эти же ссылки,
// но по одной (что бы пользлователь не забывал читать статьи, к примеру, которые он сохранил)

import (
	"context"
	tgClient "firstTGBot/clients/telegram"
	"firstTGBot/consumer/eventConsumer"
	"firstTGBot/events/telegram"
	"firstTGBot/storage/sqlite"
	"flag"
	"log"
)

const (
	tgBotHost         = "api.telegram.org"
	batchSize         = 100
	sqliteStoragePath = "data/sqlite/storage.db"
)

func main() {

	/* клиент, который будет общатся с телеграмом, он будет в пакете telegram.
	В теории, в будущем у нас может быть много различных клиентов, которые общаются \
	с теми или иными сервисами, по этому создадим для них заранее папку clients
	*/

	// ex, err := os.Executable()
	// if err != nil {
	// 	panic(err)
	// }

	bd, err := sqlite.New(sqliteStoragePath)
	if err != nil {
		log.Fatal("can not connect to storage: ", err)
	}
	// если мы используем такой контекст, который не ограничивает какими-либо таймаутами или дедлайнами,
	// то мы говорим, что сделаем когда-то в будущем контекст с таймаутом или дедлайном (context.WithTimeout(context.Background(), 5 * time.Second()))
	// background говорит о том, что таймауты и дедлайны вообще не нужны
	if err := bd.Init(context.TODO()); err != nil {
		log.Fatal("can not init storage: ", err)
	}

	eventProcessor := telegram.New(
		tgClient.New(tgBotHost, mustToken()),
		bd,
	)

	log.Print("service started")

	consumer := eventConsumer.New(eventProcessor, eventProcessor, batchSize)

	if err := consumer.Start(); err != nil {
		log.Fatal("service is stopped", err)
	}

}

/*
Нельзя просто так в коде объявить токен (из-за него могут угнать вашего бота), по этому нужно считать этот токен с консоли с помощью флагов.
Пишем отдельную функцию что бы остальной код был минималистичным. Авторы голанг запрещают писать getToken, по этому лучше написать просто Token.
Из-за того, что нащу программу бесполезно запускать без токена, можно просто упустить обработку ошибки и завершить работу программы на этом этапе.
Для таких функций, которые вместо того, что бы возвращать ошибку, аварийно завершают программу, нужно писать приставку must.
*/
func mustToken() string {
	//В переменной token будет лежать ссылка на значение флага (*String). Значение присваивается в функции flag.Parse()
	token := flag.String(
		"tg-bot-token",               // name (Имя флага. Во время запуска программы нам нужно указать токен в таком виде: bot -tg-bot-token 'my token')
		"",                           // value (флаг по умолчанию)
		"token for access to tg bot", // usage (подсказка для использования флага)
	)
	flag.Parse()

	if *token == "" { // такую проверку лучше не делать, но в этом примере можно.
		log.Fatal("token is not specified")
	}
	return *token
}
