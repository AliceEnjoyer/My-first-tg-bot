package telegram

import (
	"errors"
	"firstTGBot/lib/e"
	"firstTGBot/storage"
	"log"
	"net/url"
	"strings"
)

const (
	RndCmd   = "/rnd"
	HelpCmd  = "/help"
	StartCmd = "/start"
)

func (p *EventProcessor) doCmd(text, username string, chatID int) error {
	text = strings.TrimSpace(text) // убирает лишние пробелы
	log.Printf("got new command '%s' from '%s'", text, username)

	/*
		как будут выглядеть команды в чате с ботом:
		add page: http://...
		rnd page: /rnd
		help: /help
		start: /start (после этой команды будет выводится приветствие + /help)
	*/

	// Если text ссылка, то "делается команда" AddCmd (просто ссылка в хранилище добавляется)
	if isAddCmd(text) {
		return p.savePage(chatID, text, username)
	}

	switch text {
	case RndCmd:
		return p.sendRandom(chatID, username)
	case HelpCmd:
		return p.sendHelp(chatID)
	case StartCmd:
		return p.sendHello(chatID)
	default:
		return p.tg.SendMessages(chatID, msgUnknownCommand)
	}
}

func (p *EventProcessor) savePage(chatID int, pageURl, username string) (err error) {
	defer func() { err = e.WrapIfErr("can not do command: save page", err) }()
	page := &storage.Page{
		URL:      pageURl,
		UserName: username,
	}

	isExists, err := p.storage.IsExists(page)
	if err != nil {
		return err
	}

	if isExists {
		return p.tg.SendMessages(chatID, msgAlreadyExists) // если такой page существует, то отправляем слиенту сообщение
		// (сообщения хранится в файле messages)
	}

	if err = p.storage.Save(page); err != nil { // (((замыкания)))
		return err
	}

	if err = p.tg.SendMessages(chatID, msgSaved); err != nil {
		return err
	}

	return nil
}

// отправляет пользователю случайную статью
func (p *EventProcessor) sendRandom(chatID int, username string) (err error) {
	defer func() { err = e.WrapIfErr("can not do command: send random page", err) }()

	page, err := p.storage.PickRandom(username)

	isNoSvaedPages := errors.Is(err, storage.ErrNoSavedPages)

	if err != nil && !isNoSvaedPages {
		return err
	}

	if isNoSvaedPages {
		return p.tg.SendMessages(chatID, msgNoSavedPages)
	}

	if err := p.tg.SendMessages(chatID, page.URL); err != nil {
		return err
	}

	return p.storage.Remove(page)
}

func (p *EventProcessor) sendHelp(chatID int) error {
	return p.tg.SendMessages(chatID, msgHelp)
}

func (p *EventProcessor) sendHello(chatID int) error {
	return p.tg.SendMessages(chatID, msgHello)
}

// это типа обертки для isUrl (я, если честно, хз)
func isAddCmd(text string) bool {
	return isUrl(text) // так сказал вынести в отдельную функцию чел с ютуба
}

func isUrl(text string) bool {
	u, err := url.Parse(text) //  будет считывать ссылки только с указаным протоколом (просто google.com без https:// нельзя)
	return err == nil && u.Host != ""
}
