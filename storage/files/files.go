package files

import (
	"encoding/gob"
	"errors"
	"firstTGBot/lib/e"
	"firstTGBot/storage"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

// тип, который будет реализововать интерфейс хранилища
type Storage struct {
	BasePath string // в какой папке будут хранится файлы
}

// Этот доступ будет работать для созданных директорий с помошью функции os.MkdirAll()
// в таком формате у всех пользователей будут права на чтение и запись
const defaultPerm = 0774

func New(BasePath string) Storage {
	return Storage{BasePath: BasePath}
}

func (s Storage) Save(page *storage.Page) (err error) {
	defer func() { err = e.WrapIfErr("can not save page", err) }()

	filePath := filepath.Join(s.BasePath, page.UserName) // создает путь сохранения ссылок с помошью функции filepath.Join(), которая подходит для всез ОС
	// (Windows - "\", Linux & Macos - "/" для ссылок на файлы)

	// создаем  папку для сохранения и проверяем на ошибки
	if err := os.MkdirAll(filePath, defaultPerm); err != nil {
		return err
	}

	// формирование имени файла
	fileName, err := fileName(page)
	if err != nil {
		return err
	}

	filePath = filepath.Join(filePath, fileName) // дописываем имя файла к пути

	// создание файла
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }() // явно указываем, что мы игнорируем ошибки закрывания файла

	// сериализация ссылки
	if err := gob.NewEncoder(file).Encode(page); err != nil { // передаем в какой io.Writer мы энкодим (сериализируем) структуру page
		return err
	}

	return nil
}

func (s Storage) PickRandom(userName string) (page *storage.Page, err error) {
	defer func() { err = e.WrapIfErr("can not pick random page", err) }()

	path := filepath.Join(s.BasePath, userName)

	files, err := os.ReadDir(path) // возвращает слайс из файлов
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, storage.ErrNoSavedPages
	}

	// получаем случайный файл
	rand.Seed(time.Now().UnixNano())
	file := files[rand.Intn(len(files))]

	return s.decodePage(filepath.Join(path, file.Name()))

}

func (s Storage) Remove(p *storage.Page) error {
	fName, err := fileName(p)
	if err != nil {
		return e.Wrap("can not remove file", err)
	}
	path := filepath.Join(s.BasePath, p.UserName, fName)
	if err := os.Remove(path); err != nil {
		msg := fmt.Sprintf("can not remove file %s", path)
		return e.Wrap(msg, err)
	}
	return nil
}

func (s Storage) IsExists(p *storage.Page) (bool, error) {
	fName, err := fileName(p)
	if err != nil {
		return false, e.Wrap("can not chek if file exists", err)
	}
	path := filepath.Join(s.BasePath, p.UserName, fName)
	switch _, err = os.Stat(path); { // проверяет статус файла
	case errors.Is(err, os.ErrNotExist):
		return false, nil
	case err != nil:
		msg := fmt.Sprintf("can not chek if file %s exists", path)
		return false, e.Wrap(msg, err)
	}
	return true, nil
}

// открытие и декодирование файла (десериализация)
func (s Storage) decodePage(filePath string) (*storage.Page, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, e.Wrap("can not decode page", err)
	}
	defer func() { _ = f.Close() }() // явно указываем, что мы игнорируем ошибки закрывания файла

	var p storage.Page // куда мцы будем возвращать нашу декодированую страницу
	if err := gob.NewDecoder(f).Decode(&p); err != nil {
		return nil, e.Wrap("can not decode page", err)
	}
	return &p, nil
}

// Будет делать уникальные имена для файлов, которые будут хранится в storage.
// Делаем это в отдельной функции, что бы когда мы захотели что-то изменить при
// создании имени нам не надо было искать каждую фнкцию storage.Page.Hash()
func fileName(p *storage.Page) (string, error) {
	return p.Hash()
}
