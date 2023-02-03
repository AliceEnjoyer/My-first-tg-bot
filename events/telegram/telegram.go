package telegram

import (
	"errors"
	"firstTGBot/clients/telegram"
	"firstTGBot/events"
	"firstTGBot/lib/e"
	"firstTGBot/storage"
)

var (
	ErrUnknownEventType = errors.New("unknown event type")
	ErrUnknownMetaType  = errors.New("unknown meta type")
)

// будет реализововать как Fetcher, так и Processor
type EventProcessor struct {
	tg      *telegram.Client
	offset  int
	storage storage.Storage // будет хранить ссылки (заметьте, что это интерфейс!!!)
}

// тип для Event, который содержит спец поля для телеграм (для Meta interface{})
type Meta struct {
	ChatId   int
	Username string
}

// создает EventProcessor
func New(client *telegram.Client, storage storage.Storage) *EventProcessor {
	return &EventProcessor{
		tg:      client,
		offset:  0,
		storage: storage,
	}
}

// реализуем интерфейс Fetcher и Processor

func (p *EventProcessor) Fetch(limit int) ([]events.Event, error) {
	updates, err := p.tg.Updates(p.offset, limit)
	if err != nil {
		return nil, e.Wrap("can not get events", err)
	}

	if len(updates) == 0 {
		return nil, nil
	}

	res := make([]events.Event, 0, len(updates)) // 0 - size, len(update) - capacity

	for _, u := range updates {
		res = append(res, event(u))
	}
	// тепер нам нужно обновить значения внутренего поля offset, поскольку при след вызове метода Fetch мы должны получить след порцию событий
	// Значение offset напрямую связано с ID апдэйта, и для того, что бы при след запросе получить след пачку обновлений, нам нужно взять
	// id полследнего апдэйта и добавить 1. И тогда, при след запросе, мы получим только те апдэйты, у которых id больше чем у последнего, уже полученых.
	p.offset = updates[len(updates)-1].ID + 1

	return res, nil

	// Ивенты - это понятие телеграма, и они относятся только к нему, в другом месенджере термина апдэйт может и не быть и работать мы можем с чем угодно,
	// а вот ивент - это более обшая сущность, в нее мы можем впихнуть любую инфу от всех мессенджеров, что бы потом ее обработать
}

func (p *EventProcessor) Process(event events.Event) error {
	switch event.Type {
	case events.Message:
		return p.processMessage(event)
	default:
		return e.Wrap("can not process message", ErrUnknownEventType)
	}
}

// для красоты выносим логику обработки сообщений в отдельную функцию
func (p *EventProcessor) processMessage(event events.Event) (err error) {
	defer func() { err = e.Wrap("can not process message", err) }()
	m, err := meta(event)
	if err != nil {
		return err
	}

	// В зависимости от типа сообщения, нам нужно выполнить те или другие действия с ним: если пользователь просто скинул ссылку,
	// то нужно ее сохранить, если пользователь отправил нам команду rnd, то мы должны найти ссылку и сохраненных и вернуть ему,
	// и так далее. Все эти группы действий чел из ютуба предложил назвать командыми, и весь код, который будет к ним относится
	// мы вынесем в отдельный файл commands
	if err := p.doCmd(event.Text, m.Username, m.ChatId); err != nil {
		return err
	}

	return nil
}

// извлекает мету из ивента
func meta(event events.Event) (Meta, error) {
	res, ok := event.Meta.(Meta)
	if !ok {
		return Meta{}, e.Wrap("can not get meta", ErrUnknownMetaType)
	}
	return res, nil
}

// преобразование типа update в тип event
func event(upd telegram.Update) events.Event {
	updType := fetchType(upd)
	res := events.Event{
		Type: updType,
		Text: fetchText(upd),
	}

	if updType == events.Message {
		res.Meta = Meta{
			ChatId:   upd.Message.Chat.ID,
			Username: upd.Message.From.UserName,
		}
	}
	return res
}

func fetchType(u telegram.Update) events.Type {
	if u.Message == nil { // если нам ничего не прислали, то возвращаем, что тип не известен
		return events.Unknown
	}
	return events.Message
}

func fetchText(u telegram.Update) string {
	if u.Message == nil { // если нам ничего не прислали, то возвращаем пустую строку
		return ""
	}
	return u.Message.Text
}
