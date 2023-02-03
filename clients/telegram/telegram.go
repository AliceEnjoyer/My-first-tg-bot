package telegram

import (
	"encoding/json"
	"firstTGBot/lib/e"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
)

type Client struct {
	host     string      // хост API телеграма
	basePath string      // базовый путь - это тот префикс, с которого начинаются все запросы ("tg-bot.com/bot<token>", но без угловых скобок)
	client   http.Client // просто http клиент
}

// на случай если в телеграме изменят название методов
const (
	getUpdatesMethod  = "getUpdates"
	sendMessageMethod = "sendMessage"
)

// будет создавать клиент
func New(host, token string) *Client {
	return &Client{
		host:     host,
		basePath: newBaseToken(token),
		client:   http.Client{},
	}
}

/*
Человеку абсолютно не важно как генирируется этот путь,
ему важно понимать что он просто как-то генерируется, по этому ее
можно свести в отдельную функцию. Если когда-нибудь телеграм решит изменить формирования этого префикса,
то нам не нужно будет везде заменять этот код (если мы создавали бэйсПас в разных местах с помощью этой функции)
*/
func newBaseToken(token string) string {
	return "bot" + token
}

// Заниматся наш клиент будет двумя вещами: получение апдэйтов (новых сообщений), и отправка собственных сообщений пользователю
// экспортируемые функции должны быть выше не экспортированых функций

func (c *Client) SendMessages(chatID int, text string) error {
	q := url.Values{}
	q.Add("chat_id", strconv.Itoa(chatID)) // в какой чат мы хотим отправить сообщение
	q.Add("text", text)                    //  текст сообщения

	if _, err := c.doRequest(sendMessageMethod, q); err != nil {
		return e.Wrap("can not send message", err)
	}

	return nil
}

// Возвращает структуры, в которой есть все, что нам нужно знать об апдэйте
func (c *Client) Updates(offset, limit int) (updates []Update, err error) {
	defer func() { err = e.WrapIfErr("can not get updates", err) }() // делаем так, что бы не повторялся код

	q := url.Values{}                     // параметры запроса (какие штуки мы хотим получить ("telegram/GET offset", "telegram/GET limit"))
	q.Add("offset", strconv.Itoa(offset)) // offset = смещение. Типа когда эта переменная равна нулю, то мы получаем первую пачку апдэйтов,
	//////////////////////////////////////// когда равна один, то вторую, и так далее, но она привязана к ID послед апдейта
	q.Add("limit", strconv.Itoa(limit)) // количество апдэйтов, которые мы будем получать за один запрос

	// отправка запроса
	data, err := c.doRequest(getUpdatesMethod, q)
	if err != nil {
		return nil, err
	}

	var res UpdatesResponse
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}

	return res.Result, nil
}

func (c *Client) doRequest(method string, query url.Values) (data []byte, err error) {
	defer func() { err = e.WrapIfErr("can not do request", err) }() // делаем так, что бы не повторялся код

	u := url.URL{ // сформирует url, на который будет отправлятся запрос
		Scheme: "https",                       // указывает протокол, по которому будет отправлятся запрос
		Host:   c.host,                        // хост
		Path:   path.Join(c.basePath, method), // передает путь, функция Join() соиденяет пути так, что бы не было лишних '/'
		//метод - это, к примеру, GET в "site.com/GET offset"
	}

	req, err := http.NewRequest( // сформирует обьект запроса (тут мы только подготавливаем запрос)
		http.MethodGet, // передается http метод, которым мы хотим воспользоватся (http.MethodGet == "GET")
		u.String(),     // передача url в текстовом виде
		nil,            // тело запроса (все необходимое уже есть в виде параметров (query) + у метода GET тела обычно нет)
	)
	if err != nil {
		return nil, err
	}

	// передает в обьект запроса (req) параметры запроса
	req.URL.RawQuery = query.Encode() // метод Encode() приведет эти параметры к такому виду, что бы мы могли отправлять из на сервер

	// отправляем получившийся запрос
	resp, err := c.client.Do(req) // для отправки мы используем тот клиент, который мы заранее подготовили, у него есть метод Do(*http.Request)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil

}
