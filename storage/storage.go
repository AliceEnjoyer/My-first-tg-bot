package storage

import (
	"crypto/sha1"
	"errors"
	"firstTGBot/lib/e"
	"fmt"
	"io"
)

// Выносим эту ошибку для того, что бы ее можно было проверить в других пакетах
var ErrNoSavedPages = errors.New("no saved pages")

// хранит инфу, которыую обработал Processor
// может хранить инфу как в файловой системе, так и в бд
type Storage interface {
	Save(p *Page) error                        // сохраняет ссылки
	PickRandom(userName string) (*Page, error) // выбирает одну рандомную ссылку
	Remove(p *Page) error                      // удаляет ссылку
	IsExists(p *Page) (bool, error)            // проверяет есть ли ссылка в хранении
}

// основной типа данных, с которым будет работать Storage
// Под этим типом мы будем понимать страницу, на которую ведет ссылка, которую мыв скинули боту.
type Page struct {
	URL      string // ссылка, которую сы скинули боту
	UserName string // имя пользователя, который скинул ссылку, что бы пониматьь кому эту ссылку отдавать
	// Created  time.Time // когда эта ссылка была создана
}

// для того, что бы были уникальные имена файлов, в которых будут сохранятся ссылки
func (p Page) Hash() (string, error) {
	h := sha1.New() // создаем хэш объект

	// добавляем текст, по которому будет создаватся хэш
	if _, err := io.WriteString(h, p.URL); err != nil {
		return "", e.Wrap("can not calculate hash", err)
	}
	if _, err := io.WriteString(h, p.UserName); err != nil {
		return "", e.Wrap("can not calculate hash", err)
	}

	// возвращаем хэш, правильно при этом его преобразовав в string
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
