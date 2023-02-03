package telegram

// структура для парсинга json, которое принимается от API телеграма
type Update struct {
	ID      int              `json:"update_id"`
	Message *IncomingMessage `json:"message"` // сообщение из чата телеграм бота (является ссылкой, потому что мы можем не получить сообщения (будет nil))
}

// структура для отправки ответа по json
type UpdatesResponse struct {
	Ok     bool     `json:"ok"` // если тут будет true, то сервер считает еще Result
	Result []Update `json:"result"`
}

// структура сообщения пользователя для парсинга
type IncomingMessage struct {
	Text string `json:"text"` // текст сообщения
	From From   `json:"from"` // имя пользователя
	Chat Chat   `json:"chat"` // id чата
}

// структура имени пользователя для парсинга json
type From struct {
	UserName string `json:"username"`
}

// структура id чата для парсинга json
type Chat struct {
	ID int `json:"id"`
}
